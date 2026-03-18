package service

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// UserService 用户服务，处理注册、登录、Token 刷新。
type UserService struct {
	repo *repository.UserRepo
}

type UserProfile struct {
	ID              uint      `json:"id"`
	Username        string    `json:"username"`
	Balance         float64   `json:"balance"`
	Avatar          string    `json:"avatar"`
	TotalSpentCents int64     `json:"total_spent_cents"`
	GrowthLevel     int       `json:"growth_level"`
	Role            string    `json:"role"`
	Permissions     []string  `json:"permissions,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

var (
	ErrUserExited    = errors.New("用户已存在")
	ErrUserNotFound  = errors.New("用户不存在")
	ErrPasswordWrong = errors.New("密码错误")
	ErrTokenInvalid  = errors.New("token 无效")
	ErrTokenExpired  = errors.New("token 过期")
)

func NewUserService(repo *repository.UserRepo) *UserService {
	return &UserService{
		repo: repo,
	}
}

// Register 注册用户，直接插入并依赖唯一键防重，密码使用哈希存储。
func (s *UserService) Register(ctx context.Context, username, password string) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	// 加密用户密码
	hashPwd, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	// 存储 user 对象进入数据库
	user := &model.User{
		Username: username,
		Password: hashPwd,
		Balance:  0,
		Role:     model.UserRoleUser,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || isMySQLDuplicate(err) {
			return ErrUserExited
		}
		return err
	}
	return nil
}

// Login 校验密码后签发 access/refresh token。
func (s *UserService) Login(ctx context.Context, username, password string) (string, string, error) {
	if ctx == nil {
		return "", "", fmt.Errorf("context is nil")
	}
	// 查找用户
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return "", "", ErrUserNotFound
	}

	// 校验密码
	if !utils.CheckPassword(password, user.Password) {
		return "", "", ErrPasswordWrong
	}

	// 签发 token
	role := model.NormalizeUserRole(user.Role)
	access, refresh, err := utils.GenerateTokens(user.ID, username, role)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

// GetProfile 查询用户信息。
func (s *UserService) GetProfile(ctx context.Context, userID uint) (*UserProfile, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	role := model.NormalizeUserRole(user.Role)
	return &UserProfile{
		ID:              user.ID,
		Username:        user.Username,
		Balance:         user.Balance,
		Avatar:          user.Avatar,
		TotalSpentCents: user.TotalSpentCents,
		GrowthLevel:     user.GrowthLevel,
		Role:            role,
		Permissions:     model.PermissionsForRole(role),
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}, nil
}

// Refresh 使用 refresh token 续签新的 access token。
func (s *UserService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("context is nil")
	}
	claims, err := utils.ParshToken(refreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrTokenExpired
		}
		return "", ErrTokenInvalid
	}
	if claims.TokenType != "refresh" {
		return "", ErrTokenInvalid
	}

	user, err := s.repo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrTokenInvalid
		}
		return "", err
	}
	role := model.NormalizeUserRole(user.Role)
	access, _, err := utils.GenerateTokens(user.ID, user.Username, role)
	if err != nil {
		return "", fmt.Errorf("生成 token 失败: %w", err)
	}
	return access, nil
}

// UpdateProfile 更新用户名或头像；用户名变更会先查重。
func (s *UserService) UpdateProfile(ctx context.Context, userID uint, username, avatar *string) (*UserProfile, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}

	if username != nil && *username != user.Username {
		_, err := s.repo.GetByUsername(ctx, *username)
		switch {
		case err == nil:
			return nil, ErrUserExited
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return nil, err
		default:
			updates["username"] = *username
			user.Username = *username
		}
	}

	if avatar != nil {
		updates["avatar"] = *avatar
		user.Avatar = *avatar
	}

	if len(updates) == 0 {
		return s.GetProfile(ctx, userID)
	}

	if err := s.repo.UpdateProfile(ctx, userID, updates); err != nil {
		return nil, err
	}

	return s.GetProfile(ctx, userID)
}

// PromoteToAdmin 将指定用户提权为管理员；已是管理员时保持幂等。
func (s *UserService) PromoteToAdmin(ctx context.Context, username string) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}

	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username is empty")
	}

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if model.IsAdminRole(user.Role) {
		return nil
	}

	return s.repo.UpdateProfile(ctx, user.ID, map[string]any{"role": model.UserRoleAdmin})
}
