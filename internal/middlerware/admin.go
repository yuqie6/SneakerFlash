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
		roleStr, _ := role.(string)
		if !exists || !model.IsAdminRole(roleStr) {
			appG.ErrorMsg(http.StatusForbidden, e.UNAUTHORIZED, "需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

func AdminResourceAuth(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		appG := app.Gin{C: c}

		role, _ := c.Get("role")
		roleStr, _ := role.(string)
		if !model.HasAdminResource(roleStr, resource) {
			appG.ErrorMsg(http.StatusForbidden, e.UNAUTHORIZED, "缺少后台资源权限")
			c.Abort()
			return
		}

		c.Set("admin_resource", resource)
		c.Next()
	}
}
