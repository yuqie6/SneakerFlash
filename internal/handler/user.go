package handler

import (
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/service"
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

// 用户注册接口
func (h *UserHandler) Register(c *gin.Context) {
	appG := app.Gin{C: c}

	var req RegisterReq
	// 1. 参数校验
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVAILID_PARAMS)
		return
	}

	// 2. 调用业务逻辑
	if err := h.svc.Register(req.Username, req.Password); err != nil {
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
		appG.Error(http.StatusBadRequest, e.INVAILID_PARAMS)
		return
	}

	token, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		appG.Error(http.StatusUnauthorized, e.ERROR_NOT_EXIST_USER)
		return
	}

	appG.Success(gin.H{"token": token})
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
