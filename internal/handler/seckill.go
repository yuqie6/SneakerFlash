package handler

import (
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
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

// 执行秒杀
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
	orderNum, err := h.svc.Seckill(userID, req.ProductID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSeckillRepeat):
			appG.Error(http.StatusOK, e.ERROR_REPEAT_BUY)
		case errors.Is(err, service.ErrSeckillFull):
			appG.Error(http.StatusOK, e.ERROR_SECKILL_FULL)
		case errors.Is(err, service.ErrSeckillBusy):
			appG.ErrorMsg(http.StatusServiceUnavailable, e.ERROR, err.Error())
		default:
			appG.Error(http.StatusInternalServerError, e.ERROR)
		}
		return
	}

	// 4. 秒杀成功
	appG.Success(gin.H{"order_num": orderNum})
}
