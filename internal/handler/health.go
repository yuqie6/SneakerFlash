package handler

import (
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	svc *service.HealthService
}

func NewHealthHandler(svc *service.HealthService) *HealthHandler {
	return &HealthHandler{svc: svc}
}

// Health 存活检查
// @Summary 存活检查
// @Description 快速确认 HTTP 服务进程可响应
// @Tags 系统
// @Produce json
// @Success 200 {object} app.Response{data=HealthStatusResponse}
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	appG := app.Gin{C: c}
	appG.Success(h.svc.Health(c.Request.Context()))
}

// Ready 就绪检查
// @Summary 就绪检查
// @Description 校验 MySQL、Redis 与 Kafka Producer 是否可用
// @Tags 系统
// @Produce json
// @Success 200 {object} app.Response{data=ReadinessStatusResponse}
// @Failure 503 {object} app.Response{data=ReadinessStatusResponse}
// @Router /ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	status, err := h.svc.Ready(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, app.Response{
			Code: e.ERROR,
			Msg:  "服务未就绪",
			Data: status,
		})
		return
	}

	appG := app.Gin{C: c}
	appG.Success(status)
}
