package handler

import (
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CouponHandler struct {
	svc *service.CouponService
}

func NewCouponHandler(svc *service.CouponService) *CouponHandler {
	return &CouponHandler{svc: svc}
}

// ListMyCoupons 我的优惠券列表
// @Summary 我的优惠券
// @Tags Coupon
// @Produce json
// @Security BearerAuth
// @Param status query string false "available/used/expired"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} app.Response{data=[]service.MyCoupon}
// @Router /coupons/mine [get]
func (h *CouponHandler) ListMyCoupons(c *gin.Context) {
	appG := app.Gin{C: c}
	ctx := c.Request.Context()

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

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.svc.ListUserCoupons(ctx, userID, status, page, pageSize)
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.SuccessWithPage(list, total, page, pageSize)
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
	ctx := c.Request.Context()

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

	uc, err := h.svc.PurchaseCoupon(ctx, userID, req.CouponID)
	if err != nil {
		appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, err.Error())
		return
	}
	appG.Success(uc)
}
