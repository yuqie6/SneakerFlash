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

type UpdateProfileReq struct {
	Username *string `json:"user_name" binding:"omitempty,min=1,max=50"`
	Avatar   *string `json:"avatar" binding:"omitempty"`
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册后返回成功提示
// @Tags 用户
// @Accept json
// @Produce json
// @Param payload body RegisterReq true "注册信息"
// @Success 200 {object} app.Response{data=MessageResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Router /register [post]
func (h *UserHandler) Register(c *gin.Context) {
	appG := app.Gin{C: c}
	svc := h.svc.WithContext(c.Request.Context())

	var req RegisterReq
	// 1. 参数校验
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	// 2. 调用业务逻辑
	if err := svc.Register(req.Username, req.Password); err != nil {
		if errors.Is(err, service.ErrUserExited) {
			appG.Error(http.StatusOK, e.ERROR_EXIST_USER)
			return
		}
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}

	appG.Success(gin.H{"message": "注册成功"})
}

// Login 用户登录
// @Summary 用户登录
// @Description 登录后下发 access/refresh token
// @Tags 用户
// @Accept json
// @Produce json
// @Param payload body RegisterReq true "登录信息"
// @Success 200 {object} app.Response{data=TokenPairResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "认证失败"
// @Router /login [post]
func (h *UserHandler) Login(c *gin.Context) {
	appG := app.Gin{C: c}
	svc := h.svc.WithContext(c.Request.Context())

	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	access, refresh, err := svc.Login(req.Username, req.Password)
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

// GetProfile 获取个人信息
// @Summary 获取当前用户信息
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Success 200 {object} app.Response{data=UserResponse}
// @Failure 401 {object} app.Response "未登录"
// @Router /profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	appG := app.Gin{C: c}
	svc := h.svc.WithContext(c.Request.Context())

	userID, exists := c.Get("userID")
	if !exists {
		appG.Error(http.StatusUnauthorized, e.ERROR_NOT_EXIST_USER)
		return
	}

	user, err := svc.GetProfile(userID.(uint))
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}

	appG.Success(user)
}

// Refresh 刷新 access token
// @Summary 刷新 access token
// @Description 使用 refresh_token 刷新新的 access_token
// @Tags 用户
// @Accept json
// @Produce json
// @Param payload body RefreshReq true "刷新参数"
// @Success 200 {object} app.Response{data=AccessTokenResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "token 失效"
// @Router /refresh [post]
func (h *UserHandler) Refresh(c *gin.Context) {
	appG := app.Gin{C: c}
	svc := h.svc.WithContext(c.Request.Context())

	var req RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	token, err := svc.Refresh(req.RefreshToken)
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

// UpdateProfile 更新个人信息
// @Summary 更新当前用户信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body UpdateProfileReq true "支持部分字段"
// @Success 200 {object} app.Response{data=UserResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Router /profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	appG := app.Gin{C: c}
	svc := h.svc.WithContext(c.Request.Context())
	userID, exists := c.Get("userID")
	if !exists {
		appG.Error(http.StatusUnauthorized, e.ERROR_NOT_EXIST_USER)
		return
	}

	var req UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	if req.Username == nil && req.Avatar == nil {
		appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, "请提供要更新的字段")
		return
	}

	user, err := svc.UpdateProfile(userID.(uint), req.Username, req.Avatar)
	switch {
	case errors.Is(err, service.ErrUserExited):
		appG.Error(http.StatusOK, e.ERROR_EXIST_USER)
		return
	case err != nil:
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}

	appG.Success(user)
}
