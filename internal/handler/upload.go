package handler

import (
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	svc *service.UploadService
}

func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

// 上传图片（头像、商品图通用）
func (h *UploadHandler) UploadImage(c *gin.Context) {
	appG := app.Gin{C: c}

	file, err := c.FormFile("file")
	if err != nil {
		appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, "未找到文件")
		return
	}

	path, err := h.svc.SaveImage(file)
	if err != nil {
		appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, err.Error())
		return
	}

	appG.Success(gin.H{
		"url": path,
	})
}
