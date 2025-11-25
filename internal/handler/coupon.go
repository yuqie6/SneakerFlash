package handler

import (
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CouponHandler struct {
	svc    *service.CouponService
	vipSvc *service.VIPService
}

func NewCouponHandler(svc *service.CouponService, vipSvc *service.VIPService) *CouponHandler {
	return &CouponHandler{svc: svc, vipSvc: vipSvc}
}

// ListMyCoupons 我的优惠券列表
// @Summary 我的优惠券
// @Tags Coupon
// @Produce json
// @Security BearerAuth
// @Param status query string false "available/used/expired"
// @Success 200 {object} app.Response{data=[]service.MyCoupon}
// @Router /coupons/mine [get]
func (h *CouponHandler) ListMyCoupons(c *gin.Context) {
	appG := app.Gin{C: c}
	svc := h.svc.WithContext(c.Request.Context())
	vipSvc := h.vipSvc.WithContext(c.Request.Context())

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

	// 登录时触发当月 VIP 配额发券
	if profile, err := vipSvc.Profile(userID); err == nil {
		_ = svc.IssueVIPMonthly(userID, profile.EffectiveLevel)
	}

	status := c.Query("status")
	list, err := svc.ListUserCoupons(userID, status)
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.Success(list)
}

type PurchaseCouponReq struct {
	CouponID uint `json:"coupon_id" binding:"required"`
}

// PurchaseCoupon 购买优惠券（当前模拟支付，直接发券）
// @Summary 购买优惠券（模拟支付）
// @Tags Coupon
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body PurchaseCouponReq true "coupon_id"
// @Success 200 {object} app.Response{data=service.MyCoupon}
// @Router /coupons/purchase [post]
func (h *CouponHandler) PurchaseCoupon(c *gin.Context) {
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

	var req PurchaseCouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	uc, err := svc.PurchaseCoupon(userID, req.CouponID)
	if err != nil {
		appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, err.Error())
		return
	}
	appG.Success(uc)
}
