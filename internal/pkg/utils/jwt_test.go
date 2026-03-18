package utils

import (
	"SneakerFlash/internal/config"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateTokensAndParse(t *testing.T) {
	config.Conf.JWT.Secret = "test-secret"
	config.Conf.JWT.Expried = 3600
	config.Conf.JWT.RefreshExpried = 7200

	access, refresh, err := GenerateTokens(7, "alice")
	if err != nil {
		t.Fatalf("GenerateTokens() error = %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatalf("GenerateTokens() returned empty token, access=%q refresh=%q", access, refresh)
	}

	accessClaims, err := ParshToken(access)
	if err != nil {
		t.Fatalf("ParshToken(access) error = %v", err)
	}
	if accessClaims.UserID != 7 || accessClaims.Username != "alice" || accessClaims.TokenType != tokenTypeAccess {
		t.Fatalf("unexpected access claims: %+v", accessClaims)
	}

	refreshClaims, err := ParshToken(refresh)
	if err != nil {
		t.Fatalf("ParshToken(refresh) error = %v", err)
	}
	if refreshClaims.TokenType != tokenTypeRefresh {
		t.Fatalf("refresh token type = %q, want %q", refreshClaims.TokenType, tokenTypeRefresh)
	}
}

func TestParshToken_InvalidSignature(t *testing.T) {
	config.Conf.JWT.Secret = "test-secret"
	config.Conf.JWT.Expried = 3600
	config.Conf.JWT.RefreshExpried = 7200
	configuredSecret := []byte("another-secret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:    1,
		Username:  "bob",
		TokenType: tokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})

	raw, err := token.SignedString(configuredSecret)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := ParshToken(raw); err == nil {
		t.Fatal("ParshToken() error = nil, want invalid signature error")
	}
}
