package middlerware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"SneakerFlash/internal/pkg/logger"

	"github.com/gin-gonic/gin"
)

// SlogMiddlerware 将 HTTP 请求指标写入 slog，便于统一追踪。
func SlogMiddlerware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		requestID := generateRequestID()

		ctx := logger.ContextWithAttrs(
			c.Request.Context(),
			slog.String("request_id", requestID),
		)
		c.Request = c.Request.WithContext(ctx)
		c.Set("request_id", requestID)

		c.Next()

		latency := time.Since(startTime)
		status := c.Writer.Status()
		size := c.Writer.Size()
		if size < 0 {
			size = 0
		}

		attrs := []slog.Attr{
			slog.String("request_id", requestID),
			slog.Int("status", status),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Duration("latency", latency),
			slog.Int("size", size),
			slog.String("client_ip", c.ClientIP()),
		}
		if query != "" {
			attrs = append(attrs, slog.String("query", query))
		}
		if ua := c.Request.UserAgent(); ua != "" {
			attrs = append(attrs, slog.String("user_agent", ua))
		}
		if ref := c.Request.Referer(); ref != "" {
			attrs = append(attrs, slog.String("referer", ref))
		}
		if errDetail := c.Errors.ByType(gin.ErrorTypePrivate).String(); errDetail != "" {
			attrs = append(attrs, slog.String("error", errDetail))
		}

		level := slog.LevelInfo
		switch {
		case status >= http.StatusInternalServerError:
			level = slog.LevelError
		case status >= http.StatusBadRequest:
			level = slog.LevelWarn
		}

		slog.Default().LogAttrs(ctx, level, "HTTP request", attrs...)
	}
}

func generateRequestID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	return hex.EncodeToString([]byte(time.Now().Format("150405.000000000")))
}
