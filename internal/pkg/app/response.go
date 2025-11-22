// 封装响应器
package app

import (
	"SneakerFlash/internal/pkg/e"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Gin struct {
	C *gin.Context
}

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func (g *Gin) response(httpCode, errCode int, msg string, data any) {
	g.C.JSON(httpCode, Response{
		Code: errCode,
		Msg:  msg,
		Data: data,
	})
}

// 基础响应
func (g *Gin) Response(httpCode, errCode int, data any) {
	g.response(httpCode, errCode, e.GetMsg(errCode), data)
}

// 自定义提示的错误响应
func (g *Gin) ErrorMsg(httpCode, errCode int, msg string) {
	g.response(httpCode, errCode, msg, nil)
}

// 成功响应
func (g *Gin) Success(data any) {
	g.Response(http.StatusOK, e.SUCCESS, data)
}

// 错误响应
func (g *Gin) Error(httpCode, errCode int) {
	g.Response(httpCode, errCode, nil)
}
