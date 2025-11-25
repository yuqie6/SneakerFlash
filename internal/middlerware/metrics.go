package middlerware

import (
	"SneakerFlash/internal/pkg/metrics"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsMiddleware 采集 HTTP 基本指标：请求总数和延迟。
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		metrics.ObserveHTTP(path, c.Request.Method, statusFamily(c.Writer.Status()), latency)
	}
}

func statusFamily(code int) string {
	switch {
	case code >= 500:
		return "5xx"
	case code >= 400:
		return "4xx"
	case code >= 300:
		return "3xx"
	case code >= 200:
		return "2xx"
	default:
		return "1xx"
	}
}

// MetricsHandler 暴露 Prometheus 文本格式。
func MetricsHandler(c *gin.Context) {
	metrics.Handler(c.Writer, c.Request)
}
