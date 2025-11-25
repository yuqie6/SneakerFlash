package handler

import (
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VIPHandler struct {
	svc *service.VIPService
}

func NewVIPHandler(svc *service.VIPService) *VIPHandler {
	return &VIPHandler{svc: svc}
}

// GetProfile 查询 VIP 信息（成长等级+付费 VIP）。
// @Summary 获取 VIP 信息
// @Tags VIP
// @Produce json
// @Security BearerAuth
// @Success 200 {object} app.Response{data=service.VIPProfile}
// @Failure 401 {object} app.Response "未登录"
// @Router /vip/profile [get]
func (h *VIPHandler) GetProfile(c *gin.Context) {
	appG := app.Gin{C: c}
	svc := h.svc.WithContext(c.Request.Context())

	userIDAny, exists := c.Get("userID")
	if !exists {
		appG.Error(http.StatusUnauthorized, e.UNAUTHORIZED)
		return
	}
	userID, ok := userIDAny.(uint)
	if !ok {
		appG.Error(http.StatusUnauthorized, e.UNAUTHORIZED)
		return
	}

	profile, err := svc.Profile(userID)
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.Success(profile)
}

type PurchaseVIPReq struct {
	PlanID int `json:"plan_id" binding:"required"`
}

// Purchase 购买付费 VIP（当前模拟支付成功，直接生效）。
// @Summary 购买付费 VIP
// @Tags VIP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body PurchaseVIPReq true "付费套餐"
// @Success 200 {object} app.Response{data=service.VIPProfile}
// @Failure 400 {object} app.Response "参数错误或套餐不存在"
// @Failure 401 {object} app.Response "未登录"
// @Router /vip/purchase [post]
func (h *VIPHandler) Purchase(c *gin.Context) {
	appG := app.Gin{C: c}
	svc := h.svc.WithContext(c.Request.Context())

	userIDAny, exists := c.Get("userID")
	if !exists {
		appG.Error(http.StatusUnauthorized, e.UNAUTHORIZED)
		return
	}
	userID, ok := userIDAny.(uint)
	if !ok {
		appG.Error(http.StatusUnauthorized, e.UNAUTHORIZED)
		return
	}

	var req PurchaseVIPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	profile, err := svc.PurchasePaidVIP(userID, req.PlanID)
	if err != nil {
		appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, err.Error())
		return
	}
	appG.Success(profile)
}
