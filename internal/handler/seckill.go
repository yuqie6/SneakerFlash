package handler

import (
	"SneakerFlash/internal/service"
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
	// 1. 获取当前用户
	uid, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}
	userID, ok := uid.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息无效"})
		return
	}

	// 2, 解析请求
	var req SeckillReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. 调用秒杀服务
	orderNum, err := h.svc.Seckill(userID, req.ProductID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	// 4. 秒杀成功
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "抢购成功, 订单生成中",
		"data": gin.H{
			"order_num": orderNum,
		},
	})
}
