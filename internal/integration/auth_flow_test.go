//go:build integration

package integration

import (
	"SneakerFlash/internal/handler"
	"SneakerFlash/internal/middlerware"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthFlow_RegisterLoginProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := setupIntegrationDB(t)

	userSvc := service.NewUserService(repository.NewUserRepo(gdb))
	userHandler := handler.NewUserHandler(userSvc)

	router := gin.New()
	api := router.Group("/api/v1")
	api.POST("/register", userHandler.Register)
	api.POST("/login", userHandler.Login)

	auth := api.Group("/")
	auth.Use(middlerware.JWTauth())
	auth.GET("/profile", userHandler.GetProfile)

	registerBody := []byte(`{"user_name":"alice","user_password":"password-123"}`)
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewReader(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")
	registerRec := httptest.NewRecorder()
	router.ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusOK {
		t.Fatalf("register status = %d, want 200 body=%s", registerRec.Code, registerRec.Body.String())
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(registerBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want 200 body=%s", loginRec.Code, loginRec.Body.String())
	}

	var loginResp struct {
		Code int `json:"code"`
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("decode login body: %v", err)
	}
	if loginResp.Data.AccessToken == "" {
		t.Fatal("login access token is empty")
	}

	profileReq := httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)
	profileReq.Header.Set("Authorization", "Bearer "+loginResp.Data.AccessToken)
	profileRec := httptest.NewRecorder()
	router.ServeHTTP(profileRec, profileReq)
	if profileRec.Code != http.StatusOK {
		t.Fatalf("profile status = %d, want 200 body=%s", profileRec.Code, profileRec.Body.String())
	}
}
