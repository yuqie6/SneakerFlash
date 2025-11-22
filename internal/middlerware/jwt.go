package middlerware

import (
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/pkg/utils"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTauth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		appG := app.Gin{C: ctx}
		// 1. 获取 header 中的 authorization
		authHeader := ctx.GetHeader("authorization")
		if authHeader == "" {
			appG.Error(http.StatusUnauthorized, e.UNAUTHORIZED)
			ctx.Abort()
			return
		}

		// 2. 格式校验
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			appG.ErrorMsg(http.StatusUnauthorized, e.UNAUTHORIZED, "token 格式有误")
			ctx.Abort()
			return
		}

		// 解析 token
		claims, err := utils.ParshToken(parts[1])
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				appG.ErrorMsg(http.StatusUnauthorized, e.UNAUTHORIZED, "token 已过期")
			} else {
				appG.ErrorMsg(http.StatusUnauthorized, e.UNAUTHORIZED, "token 无效")
			}
			ctx.Abort()
			return
		}
		if claims.TokenType != "access" {
			appG.ErrorMsg(http.StatusUnauthorized, e.UNAUTHORIZED, "token 类型无效")
			ctx.Abort()
			return
		}

		// 4. 存入用户信息到context
		ctx.Set("userID", claims.UserID)
		ctx.Set("username", claims.Username)

		ctx.Next()
	}
}
