package handler

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/service"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	svc *service.ProductService
}

type CreateProductReq struct {
	Name      string  `json:"name" binding:"required"`
	Price     float64 `json:"price" binding:"required,gt=0"`
	Stock     int     `json:"stock" binding:"required,gt=0"`
	StartTime string  `json:"start_time" binding:"required,datetime=2006-01-02 15:04:05"`
	Image     string  `json:"image"`
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{
		svc: svc,
	}
}

// 发布商品
func (h *ProductHandler) Create(c *gin.Context) {
	appG := app.Gin{C: c}
	var req CreateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	startTime, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
	if err != nil || !startTime.After(time.Now()) {
		appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, "开始时间必须晚于当前时间")
		return
	}

	p := &model.Product{
		Name:      req.Name,
		Price:     req.Price,
		Stock:     req.Stock,
		StartTime: startTime,
		Image:     req.Image,
	}

	if err := h.svc.CreateProduct(p); err != nil {
		switch {
		case errors.Is(err, service.ErrProductDuplicate):
			appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, "商品已存在，请勿重复提交")
		default:
			appG.Error(http.StatusInternalServerError, e.ERROR)
		}
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
		"items": list,
		"total": total,
		"page":  page,
	})
}
