package handler

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{
		svc: svc,
	}
}

// 发布商品
func (h *ProductHandler) Create(c *gin.Context) {
	appG := app.Gin{C: c}
	var p model.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	if err := h.svc.CreateProduct(&p); err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}

	appG.Success(p)
}

// 获取商品详情
func (h *ProductHandler) GetProduct(c *gin.Context) {
	appG := app.Gin{C: c}
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	p, err := h.svc.GetProductByID(uint(id))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound), errors.Is(err, gorm.ErrRecordNotFound):
			appG.Error(http.StatusNotFound, e.ERROR_NOT_EXIST_PRODUCT)
		default:
			appG.Error(http.StatusInternalServerError, e.ERROR)
		}
		return
	}

	appG.Success(p)
}

// 获取商品列表
func (h *ProductHandler) ListProducts(c *gin.Context) {
	appG := app.Gin{C: c}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}
	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	list, total, err := h.svc.ListProducts(page, size)
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}

	appG.Success(gin.H{
		"data":  list,
		"total": total,
		"page":  page,
	})
}
