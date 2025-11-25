package handler

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/service"
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	orderSvc   *service.OrderService
	productSvc *service.ProductService
}

type CreateOrderReq struct {
	ProductID uint `json:"product_id" binding:"required"`
}

type PaymentCallbackReq struct {
	PaymentID  string `json:"payment_id" binding:"required"`
	Status     string `json:"status" binding:"required"` // paid / failed / refunded
	NotifyData string `json:"notify_data"`
}

func NewOrderHandler(orderSvc *service.OrderService, productSvc *service.ProductService) *OrderHandler {
	return &OrderHandler{
		orderSvc:   orderSvc,
		productSvc: productSvc,
	}
}

// CreateOrder 创建订单并初始化支付单（幂等：user+product）
// @Summary 创建订单并初始化支付
// @Tags 订单
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body CreateOrderReq true "订单参数"
// @Success 200 {object} app.Response{data=OrderWithPaymentResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 404 {object} app.Response "商品不存在"
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	appG := app.Gin{C: c}
	orderSvc := h.orderSvc.WithContext(c.Request.Context())
	productSvc := h.productSvc.WithContext(c.Request.Context())

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

	var req CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	// 查询商品价格，计算支付金额（分）
	product, err := productSvc.GetProductByID(req.ProductID)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			appG.Error(http.StatusNotFound, e.ERROR_NOT_EXIST_PRODUCT)
			return
		}
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	amountCents := int64(math.Round(product.Price * 100))
	if amountCents <= 0 {
		appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, "商品价格异常")
		return
	}

	orderWithPayment, err := orderSvc.CreateOrderAndInitPayment(userID, req.ProductID, amountCents)
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}

	appG.Success(gin.H{
		"order":   orderWithPayment.Order,
		"payment": orderWithPayment.Payment,
	})
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
	orderSvc := h.orderSvc.WithContext(c.Request.Context())

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

	orders, total, err := orderSvc.ListOrders(userID, statusPtr, page, size)
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
// @Success 200 {object} app.Response{data=OrderWithPaymentResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 404 {object} app.Response "未找到"
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	appG := app.Gin{C: c}
	orderSvc := h.orderSvc.WithContext(c.Request.Context())

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
	orderWithPayment, err := orderSvc.GetOrderWithPayment(userID, uint(id))
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
	orderSvc := h.orderSvc.WithContext(c.Request.Context())

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

	orderWithPayment, err := orderSvc.HandlePaymentResult(req.PaymentID, targetStatus, req.NotifyData)
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
