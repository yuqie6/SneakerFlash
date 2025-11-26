package handler

import (
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/pkg/metrics"
	"SneakerFlash/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SeckillHandler struct {
	svc *service.SeckillService
}

func NewSeckillHandler(svc *service.SeckillService) *SeckillHandler {
	return &SeckillHandler{
		svc: svc,
	}
}

type SeckillReq struct {
	ProductID uint `json:"product_id" binding:"required"`
}

type SeckillResponse struct {
	OrderNum  string `json:"order_num"`
	OrderID   uint   `json:"order_id"`
	PaymentID string `json:"payment_id"`
}

// Seckill 执行秒杀
// @Summary 秒杀抢购
// @Tags 秒杀
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body SeckillReq true "秒杀参数"
// @Success 200 {object} app.Response{data=SeckillResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 429 {object} app.Response "被限流"
// @Router /seckill [post]
func (h *SeckillHandler) Seckill(c *gin.Context) {
	appG := app.Gin{C: c}
	// 1. 获取当前用户
	uid, exists := c.Get("userID")
	if !exists {
		appG.Error(http.StatusUnauthorized, e.ERROR_AUTH_CHECK_TOKEN_FAIL)
		return
	}
	userID, ok := uid.(uint)
	if !ok {
		appG.Error(http.StatusUnauthorized, e.ERROR_AUTH_CHECK_TOKEN_FAIL)
		return
	}

	// 2, 解析请求
	var req SeckillReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	// 3. 调用秒杀服务
	svc := h.svc.WithContext(c.Request.Context())
	result, err := svc.Seckill(userID, req.ProductID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSeckillRepeat):
			metrics.IncSeckillResult("repeat")
			appG.Error(http.StatusOK, e.ERROR_REPEAT_BUY)
		case errors.Is(err, service.ErrSeckillFull):
			metrics.IncSeckillResult("sold_out")
			appG.Error(http.StatusOK, e.ERROR_SECKILL_FULL)
		case errors.Is(err, service.ErrSeckillNotStart):
			metrics.IncSeckillResult("not_started")
			appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, err.Error())
		case errors.Is(err, service.ErrProductNotFound):
			appG.Error(http.StatusNotFound, e.ERROR_NOT_EXIST_PRODUCT)
		case errors.Is(err, service.ErrSeckillBusy):
			metrics.IncSeckillResult("busy")
			appG.ErrorMsg(http.StatusServiceUnavailable, e.ERROR, err.Error())
		default:
			metrics.IncSeckillResult("error")
			appG.Error(http.StatusInternalServerError, e.ERROR)
		}
		return
	}

	// 4. 秒杀成功
	metrics.IncSeckillResult("success")
	appG.Success(SeckillResponse{
		OrderNum:  result.OrderNum,
		OrderID:   result.OrderID,
		PaymentID: result.PaymentID,
	})
}
