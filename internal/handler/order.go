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

// 创建订单并初始化支付单（幂等：user+product）
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	appG := app.Gin{C: c}

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
	product, err := h.productSvc.GetProductByID(req.ProductID)
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

	orderWithPayment, err := h.orderSvc.CreateOrderAndInitPayment(userID, req.ProductID, amountCents)
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}

	appG.Success(gin.H{
		"order":   orderWithPayment.Order,
		"payment": orderWithPayment.Payment,
	})
}

// 订单列表
func (h *OrderHandler) ListOrders(c *gin.Context) {
	appG := app.Gin{C: c}

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

	orders, total, err := h.orderSvc.ListOrders(userID, statusPtr, page, size)
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

// 订单详情
func (h *OrderHandler) GetOrder(c *gin.Context) {
	appG := app.Gin{C: c}

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

	orderWithPayment, err := h.orderSvc.GetOrderWithPayment(userID, uint(id))
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

// 支付回调（模拟验签已完成）
func (h *OrderHandler) PaymentCallback(c *gin.Context) {
	appG := app.Gin{C: c}

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

	orderWithPayment, err := h.orderSvc.HandlePaymentResult(req.PaymentID, targetStatus, req.NotifyData)
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
