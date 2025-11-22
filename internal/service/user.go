package service

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/repository"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
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

// 业务 1: 用户注册
// 参数: username, password
func (s *UserService) Register(username, password string) error {
	// 检查用户是否存在
	_, err := s.repo.GetByUsername(username)
	if err == nil {
		return ErrUserExited
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
	}
	return s.repo.Create(user)
}

// 业务 2: 用户登录
// 入参: username, password
// 出参: access token, refresh token, err
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

// 业务 3: 获取个人信息
func (s *UserService) GetProfile(userID uint) (*model.User, error) {
	return s.repo.GetByID(userID)
}

// 刷新 token
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
