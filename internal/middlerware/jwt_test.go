package middlerware

import (
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/testutil"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestJWTauth(t *testing.T) {
	testutil.SetupTestConfig()
	gin.SetMode(gin.TestMode)

	validAccess, _, err := utils.GenerateTokens(42, "alice")
	if err != nil {
		t.Fatalf("GenerateTokens() error = %v", err)
	}
	_, validRefresh, err := utils.GenerateTokens(7, "bob")
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
				c.JSON(http.StatusOK, gin.H{
					"user_id":  userID,
					"username": username,
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
		})
	}
}
