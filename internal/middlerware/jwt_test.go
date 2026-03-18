package middlerware

import (
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/testutil"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestJWTauth(t *testing.T) {
	testutil.SetupTestConfig()
	gin.SetMode(gin.TestMode)

	validAccess, _, err := utils.GenerateTokens(42, "alice", "admin")
	if err != nil {
		t.Fatalf("GenerateTokens() error = %v", err)
	}
	_, validRefresh, err := utils.GenerateTokens(7, "bob", "user")
	if err != nil {
		t.Fatalf("GenerateTokens() refresh error = %v", err)
	}

	tests := []struct {
		name       string
		header     string
		wantStatus int
		wantUserID any
	}{
		{name: "missing token", wantStatus: http.StatusUnauthorized},
		{name: "invalid format", header: "Token abc", wantStatus: http.StatusUnauthorized},
		{name: "refresh token not allowed", header: "Bearer " + validRefresh, wantStatus: http.StatusUnauthorized},
		{name: "access token success", header: "Bearer " + validAccess, wantStatus: http.StatusOK, wantUserID: uint(42)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/protected", JWTauth(), func(c *gin.Context) {
				userID, _ := c.Get("userID")
				username, _ := c.Get("username")
				role, _ := c.Get("role")
				c.JSON(http.StatusOK, gin.H{
					"user_id":  userID,
					"username": username,
					"role":     role,
				})
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus != http.StatusOK {
				return
			}

			var payload map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("decode body: %v", err)
			}

			if got := payload["user_id"]; got != float64(tt.wantUserID.(uint)) {
				t.Fatalf("user_id = %v, want %v", got, tt.wantUserID)
			}
			if payload["username"] != "alice" {
				t.Fatalf("username = %v, want alice", payload["username"])
			}
			if payload["role"] != "admin" {
				t.Fatalf("role = %v, want admin", payload["role"])
			}
		})
	}
}

func TestAdminAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		role       string
		wantStatus int
	}{
		{name: "admin allowed", role: "admin", wantStatus: http.StatusOK},
		{name: "user denied", role: "user", wantStatus: http.StatusForbidden},
		{name: "missing role denied", wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/admin", func(c *gin.Context) {
				if tt.role != "" {
					c.Set("role", tt.role)
				}
				AdminAuth()(c)
			}, func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestBlackListMiddleware_UsesBearerUserIDBeforeJWTauth(t *testing.T) {
	testutil.SetupTestConfig()
	testutil.SetupTestRedis(t)
	gin.SetMode(gin.TestMode)

	validAccess, _, err := utils.GenerateTokens(42, "alice", "user")
	if err != nil {
		t.Fatalf("GenerateTokens() error = %v", err)
	}
	if err := redis.RDB.SAdd(context.Background(), "risk:user:black", "42").Err(); err != nil {
		t.Fatalf("SAdd() error = %v", err)
	}

	router := gin.New()
	router.GET("/protected", BlackListMiddleware(redis.RDB), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validAccess)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}
