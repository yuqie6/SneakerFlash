package handler

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

type RegisterReq struct {
	Username string `json:"user_name" binding:"required"`
	Password string `json:"user_password" binding:"required"`
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// 用户注册接口
func (h *UserHandler) Register(c *gin.Context) {
	appG := app.Gin{C: c}

	var req RegisterReq
	// 1. 参数校验
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	// 2. 调用业务逻辑
	if err := h.svc.Register(req.Username, req.Password); err != nil {
		if errors.Is(err, service.ErrUserExited) {
			appG.Error(http.StatusOK, e.ERROR_EXIST_USER)
			return
		}
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}

	appG.Success(gin.H{"message": "注册成功"})
}

// 用户登录接口
func (h *UserHandler) Login(c *gin.Context) {
	appG := app.Gin{C: c}

	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	access, refresh, err := h.svc.Login(req.Username, req.Password)
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		appG.Error(http.StatusUnauthorized, e.ERROR_NOT_EXIST_USER)
		return
	case errors.Is(err, service.ErrPasswordWrong):
		appG.ErrorMsg(http.StatusUnauthorized, e.ERROR_AUTH_CHECK_TOKEN_FAIL, "用户名或密码错误")
		return
	case err != nil:
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}

	appG.Success(gin.H{
		"access_token":  access,
		"refresh_token": refresh,
		"expires_in":    config.Conf.JWT.Expried,
	})
}

// 获取个人信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	appG := app.Gin{C: c}

	userID, exists := c.Get("userID")
	if !exists {
		appG.Error(http.StatusUnauthorized, e.ERROR_NOT_EXIST_USER)
		return
	}

	user, err := h.svc.GetProfile(userID.(uint))
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}

	appG.Success(user)
}

// 刷新 access token
func (h *UserHandler) Refresh(c *gin.Context) {
	appG := app.Gin{C: c}

	var req RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	token, err := h.svc.Refresh(req.RefreshToken)
	switch {
	case errors.Is(err, service.ErrTokenInvalid), errors.Is(err, service.ErrTokenExpired):
		appG.Error(http.StatusUnauthorized, e.UNAUTHORIZED)
		return
	case err != nil:
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}

	appG.Success(gin.H{
		"access_token": token,
		"expires_in":   config.Conf.JWT.Expried,
	})
}
