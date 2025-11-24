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

// UploadImage 上传图片（头像、商品图通用）
// @Summary 上传图片
// @Tags 文件
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param file formData file true "图片文件"
// @Success 200 {object} app.Response{data=UploadURLResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Router /upload [post]
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
