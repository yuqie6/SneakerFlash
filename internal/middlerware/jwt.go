package middlerware

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/pkg/utils"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTauth 校验 Bearer Token，验证类型为 access，解析后将 userID/username 注入上下文。
func JWTauth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		appG := app.Gin{C: ctx}
		tokenValue, err := resolveAccessToken(ctx)
		if err != nil {
			appG.ErrorMsg(http.StatusUnauthorized, e.UNAUTHORIZED, err.Error())
			ctx.Abort()
			return
		}

		claims, err := utils.ParshToken(tokenValue)
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
		role := model.NormalizeUserRole(claims.Role)
		ctx.Set("role", role)
		ctx.Set("permissions", model.PermissionsForRole(role))

		ctx.Next()
	}
}

func resolveAccessToken(ctx *gin.Context) (string, error) {
	authHeader := ctx.GetHeader("authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return "", errors.New("token 格式有误")
		}
		return parts[1], nil
	}

	if !strings.HasPrefix(ctx.FullPath(), "/api/v1/stream/") {
		return "", errors.New("未提供 token")
	}

	tokenValue := strings.TrimSpace(ctx.Query("access_token"))
	if tokenValue == "" {
		return "", errors.New("未提供 token")
	}
	return tokenValue, nil
}
