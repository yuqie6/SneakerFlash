package middlerware

import (
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTauth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		appG := app.Gin{C: ctx}
		// 1. 获取 header 中的 authorization
		authHeader := ctx.GetHeader("authorization")
		if authHeader == "" {
			appG.ErrorMsg(http.StatusUnauthorized, e.ERROR_AUTH_CHECK_TOKEN_FAIL, "请先登录")
			ctx.Abort()
			return
		}

		// 2. 格式校验
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			appG.ErrorMsg(http.StatusUnauthorized, e.ERROR_AUTH_CHECK_TOKEN_FAIL, "token 格式有误")
			ctx.Abort()
			return
		}

		// 解析 token
		claims, err := utils.ParshToken(parts[1])
		if err != nil {
			appG.ErrorMsg(http.StatusUnauthorized, e.ERROR_AUTH_CHECK_TOKEN_FAIL, "token 无效")
			ctx.Abort()
			return
		}

		// 4. 存入用户信息到context
		ctx.Set("userID", claims.UserID)
		ctx.Set("username", claims.Username)

		ctx.Next()
	}
}
