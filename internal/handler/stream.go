package handler

import (
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/service"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StreamHandler struct {
	streamSvc *service.StreamService
}

func NewStreamHandler(streamSvc *service.StreamService) *StreamHandler {
	return &StreamHandler{streamSvc: streamSvc}
}

// OrderEvents 订阅订单状态推送
// @Summary 订阅订单状态推送
// @Tags 推送
// @Produce text/event-stream
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Param access_token query string false "access token"
// @Success 200 {string} string "SSE stream established"
// @Failure 401 {object} app.Response "未登录"
// @Router /stream/orders/{id} [get]
func (h *StreamHandler) OrderEvents(c *gin.Context) {
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

	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil || orderID <= 0 {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	events, unsubscribe, err := h.streamSvc.SubscribeOrder(userID, uint(orderID))
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	defer unsubscribe()
	streamEvents(c, events)
}

// ProductEvents 订阅商品库存推送
// @Summary 订阅商品库存推送
// @Tags 推送
// @Produce text/event-stream
// @Security BearerAuth
// @Param id path int true "商品ID"
// @Param access_token query string false "access token"
// @Success 200 {string} string "SSE stream established"
// @Failure 401 {object} app.Response "未登录"
// @Router /stream/products/{id} [get]
func (h *StreamHandler) ProductEvents(c *gin.Context) {
	appG := app.Gin{C: c}
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil || productID <= 0 {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	events, unsubscribe, err := h.streamSvc.SubscribeProduct(uint(productID))
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	defer unsubscribe()
	streamEvents(c, events)
}

func streamEvents(c *gin.Context, events <-chan []byte) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case payload, ok := <-events:
			if !ok {
				return false
			}
			c.SSEvent("message", string(payload))
			return true
		}
	})
}
