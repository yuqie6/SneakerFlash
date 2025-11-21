package middlerware

import (
	"SneakerFlash/internal/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTauth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. 获取 header 中的 authorization
		authHeader := ctx.GetHeader("authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			ctx.Abort()
			return
		}

		// 2. 格式校验
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "token 格式有误"})
			ctx.Abort()
			return
		}

		// 解析 token
		claims, err := utils.ParshToken(parts[1])
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "token 无效"})
			ctx.Abort()
			return
		}

		// 4. 存入用户信息到context
		ctx.Set("userID", claims.UserID)
		ctx.Set("username", claims.Username)

		ctx.Next()
	}
}
