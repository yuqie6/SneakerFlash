package utils

import (
	"SneakerFlash/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

// GenerateTokens 签发 access 与 refresh token，TTL 读取配置。
func GenerateTokens(userID uint, username string) (accessToken, refreshToken string, err error) {
	accessToken, err = generateToken(userID, username, tokenTypeAccess, config.Conf.JWT.Expried)
	if err != nil {
		return "", "", err
	}
	refreshTTL := config.Conf.JWT.RefreshExpried
	if refreshTTL == 0 {
		refreshTTL = config.Conf.JWT.Expried * 7
	}
	refreshToken, err = generateToken(userID, username, tokenTypeRefresh, refreshTTL)
	return
}

func generateToken(userID uint, username, tokenType string, ttlSeconds int) (string, error) {
	now := time.Now()
	expireTime := now.Add(time.Duration(ttlSeconds) * time.Second)

	claims := Claims{
		UserID:    userID,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "sneaker-flash",
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tokenClaims.SignedString([]byte(config.Conf.JWT.Secret))
}

// ParshToken 解析并校验 JWT（HS256），返回自定义 Claims。
func ParshToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		return []byte(config.Conf.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}
