package handler

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	orderSvc *service.OrderService
}

type PaymentCallbackReq struct {
	PaymentID  string `json:"payment_id" binding:"required"`
	Status     string `json:"status" binding:"required"` // paid / failed / refunded
	NotifyData string `json:"notify_data"`
}

type ApplyCouponReq struct {
	CouponID *uint `json:"coupon_id" binding:"omitempty"`
}

type PollOrderResponse struct {
	Status    string                    `json:"status"`
	OrderNum  string                    `json:"order_num"`
	PaymentID string                    `json:"payment_id,omitempty"`
	Order     *service.OrderWithPayment `json:"order,omitempty"`
	Message   string                    `json:"message,omitempty"`
}

func NewOrderHandler(orderSvc *service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderSvc: orderSvc,
	}
}

// ListOrders 订单列表
// @Summary 查询订单列表
// @Tags 订单
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param size query int false "每页条数" default(10)
// @Param status query int false "订单状态：0未支付 1已支付 2失败"
// @Success 200 {object} app.Response{data=OrderListResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Router /orders [get]
func (h *OrderHandler) ListOrders(c *gin.Context) {
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

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}
	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil || size <= 0 {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	var statusPtr *model.OrderStatus
	if statusStr := c.Query("status"); statusStr != "" {
		statusInt, convErr := strconv.Atoi(statusStr)
		if convErr != nil {
			appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
			return
		}
		status := model.OrderStatus(statusInt)
		statusPtr = &status
	}

	orders, total, err := h.orderSvc.ListOrders(ctx, userID, statusPtr, page, size)
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}

	appG.Success(gin.H{
		"items": orders,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetOrder 订单详情
// @Summary 获取订单详情
// @Tags 订单
// @Produce json
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Success 200 {object} app.Response{data=PollOrderResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 404 {object} app.Response "未找到"
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
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

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}
	orderWithPayment, err := h.orderSvc.GetOrderWithPayment(ctx, userID, uint(id))
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			appG.Error(http.StatusNotFound, e.INVALID_PARAMS)
			return
		}
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}

	appG.Success(orderWithPayment)
}

// PollOrder 轮询订单创建状态（异步秒杀）。
// @Summary 轮询订单创建状态
// @Tags 订单
// @Produce json
// @Security BearerAuth
// @Param order_num path string true "订单号"
// @Success 200 {object} app.Response{data=OrderWithPaymentResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Router /orders/poll/{order_num} [get]
func (h *OrderHandler) PollOrder(c *gin.Context) {
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

	orderNum := c.Param("order_num")
	if orderNum == "" {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	result, err := h.orderSvc.PollOrder(ctx, userID, orderNum)
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}

	resp := PollOrderResponse{
		Status:    string(result.Status),
		OrderNum:  result.OrderNum,
		PaymentID: result.PaymentID,
		Order:     result.Order,
		Message:   result.Message,
	}
	appG.Success(resp)
}

// ApplyCoupon 在订单支付前应用/更换优惠券
// @Summary 订单应用优惠券
// @Tags 订单
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Param payload body ApplyCouponReq true "优惠券参数，coupon_id 为空表示不使用优惠券"
// @Success 200 {object} app.Response{data=OrderWithPaymentResponse}
// @Failure 400 {object} app.Response "参数错误或状态不允许"
// @Failure 401 {object} app.Response "未登录"
// @Failure 404 {object} app.Response "订单不存在"
// @Router /orders/{id}/apply-coupon [post]
func (h *OrderHandler) ApplyCoupon(c *gin.Context) {
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

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	var req ApplyCouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	result, err := h.orderSvc.ApplyCoupon(ctx, userID, uint(id), req.CouponID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderNotFound):
			appG.Error(http.StatusNotFound, e.INVALID_PARAMS)
		case errors.Is(err, service.ErrOrderNotPayable):
			appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, "订单状态不可支付")
		default:
			appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, err.Error())
		}
		return
	}

	appG.Success(result)
}

// PaymentCallback 支付回调（模拟验签已完成）
// @Summary 支付回调
// @Tags 订单
// @Accept json
// @Produce json
// @Param payload body PaymentCallbackReq true "支付回调参数"
// @Success 200 {object} app.Response{data=OrderWithPaymentResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 404 {object} app.Response "支付单不存在"
// @Router /payment/callback [post]
func (h *OrderHandler) PaymentCallback(c *gin.Context) {
	appG := app.Gin{C: c}
	ctx := c.Request.Context()

	var req PaymentCallbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	var targetStatus model.PaymentStatus
	switch req.Status {
	case string(model.PaymentStatusPaid):
		targetStatus = model.PaymentStatusPaid
	case string(model.PaymentStatusFailed):
		targetStatus = model.PaymentStatusFailed
	case string(model.PaymentStatusRefunded):
		targetStatus = model.PaymentStatusRefunded
	default:
		appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, "status 仅支持 paid/failed/refunded")
		return
	}

	orderWithPayment, err := h.orderSvc.HandlePaymentResult(ctx, req.PaymentID, targetStatus, req.NotifyData)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPaymentNotFound):
			appG.Error(http.StatusNotFound, e.INVALID_PARAMS)
		case errors.Is(err, service.ErrUnsupportedPayStatus):
			appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		default:
			appG.Error(http.StatusInternalServerError, e.ERROR)
		}
		return
	}

	appG.Success(orderWithPayment)
}
