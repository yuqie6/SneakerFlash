package middlerware

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		appG := app.Gin{C: c}

		role, exists := c.Get("role")
		if !exists || role != model.UserRoleAdmin {
			appG.ErrorMsg(http.StatusForbidden, e.UNAUTHORIZED, "需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}
