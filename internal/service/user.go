package service

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/repository"
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type UserService struct {
	repo *repository.UserRepo
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

// WithContext 绑定请求上下文，让用户相关数据库操作日志带上 request_id。
func (s *UserService) WithContext(ctx context.Context) *UserService {
	if ctx == nil {
		return s
	}
	return &UserService{repo: s.repo.WithContext(ctx)}
}

// Register 注册用户，直接插入并依赖唯一键防重，密码使用哈希存储。
func (s *UserService) Register(username, password string) error {
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
	}
	if err := s.repo.Create(user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || isMySQLDuplicate(err) {
			return ErrUserExited
		}
		return err
	}
	return nil
}

// Login 校验密码后签发 access/refresh token。
func (s *UserService) Login(username, password string) (string, string, error) {
	// 查找用户
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return "", "", ErrUserNotFound
	}

	// 校验密码
	if !utils.CheckPassword(password, user.Password) {
		return "", "", ErrPasswordWrong
	}

	// 签发 token
	access, refresh, err := utils.GenerateTokens(user.ID, username)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

// GetProfile 查询用户信息。
func (s *UserService) GetProfile(userID uint) (*model.User, error) {
	return s.repo.GetByID(userID)
}

// Refresh 使用 refresh token 续签新的 access token。
func (s *UserService) Refresh(refreshToken string) (string, error) {
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

	access, _, err := utils.GenerateTokens(claims.UserID, claims.Username)
	if err != nil {
		return "", fmt.Errorf("生成 token 失败: %w", err)
	}
	return access, nil
}

// UpdateProfile 更新用户名或头像；用户名变更会先查重。
func (s *UserService) UpdateProfile(userID uint, username, avatar *string) (*model.User, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}

	if username != nil && *username != user.Username {
		_, err := s.repo.GetByUsername(*username)
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
		return user, nil
	}

	if err := s.repo.UpdateProfile(userID, updates); err != nil {
		return nil, err
	}

	return user, nil
}
