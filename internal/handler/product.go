package handler

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/service"
	"errors"
	"fmt"
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
	StartTime string  `json:"start_time" binding:"required"`
	Image     string  `json:"image"`
}

type UpdateProductReq struct {
	Name      *string  `json:"name" binding:"omitempty"`
	Price     *float64 `json:"price" binding:"omitempty,gt=0"`
	Stock     *int     `json:"stock" binding:"omitempty,gt=0"`
	StartTime *string  `json:"start_time" binding:"omitempty"`
	Image     *string  `json:"image"`
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{
		svc: svc,
	}
}

func parseStartTime(raw string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006/01/02 15:04:05",
		"2006/01/02/15:04",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time format")
}

// Create 发布商品
// @Summary 发布商品
// @Tags 商品
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body CreateProductReq true "商品参数"
// @Success 200 {object} app.Response{data=model.Product}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Router /products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	appG := app.Gin{C: c}
	svc := h.svc.WithContext(c.Request.Context())
	uidAny, ok := c.Get("userID")
	if !ok {
		appG.Error(http.StatusUnauthorized, e.UNAUTHORIZED)
		return
	}
	userID := uidAny.(uint)

	var req CreateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	startTime, err := parseStartTime(req.StartTime)
	if err != nil || !startTime.After(time.Now()) {
		appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, "开始时间必须晚于当前时间（格式示例：2025-11-24 22:11:00）")
		return
	}

	p := &model.Product{
		UserID:    userID,
		Name:      req.Name,
		Price:     req.Price,
		Stock:     req.Stock,
		StartTime: startTime,
		Image:     req.Image,
	}

	if err := svc.CreateProduct(p); err != nil {
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

// GetProduct 获取商品详情
// @Summary 获取商品详情
// @Tags 商品
// @Produce json
// @Param id path int true "商品ID"
// @Success 200 {object} app.Response{data=model.Product}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 404 {object} app.Response "未找到"
// @Router /product/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	appG := app.Gin{C: c}
	svc := h.svc.WithContext(c.Request.Context())
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	p, err := svc.GetProductByID(uint(id))
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

// ListProducts 获取商品列表
// @Summary 获取商品列表
// @Tags 商品
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页条数" default(10)
// @Success 200 {object} app.Response{data=ProductListResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Router /products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	appG := app.Gin{C: c}
	svc := h.svc.WithContext(c.Request.Context())
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

	list, total, err := svc.ListProducts(page, size)
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

// UpdateProduct 更新商品（仅创建者）
// @Summary 更新商品
// @Tags 商品
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "商品ID"
// @Param payload body UpdateProductReq true "商品参数"
// @Success 200 {object} app.Response{data=IDResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 404 {object} app.Response "未找到"
// @Router /products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	appG := app.Gin{C: c}
	svc := h.svc.WithContext(c.Request.Context())
	uidAny, ok := c.Get("userID")
	if !ok {
		appG.Error(http.StatusUnauthorized, e.UNAUTHORIZED)
		return
	}
	userID := uidAny.(uint)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	var req UpdateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.Stock != nil {
		updates["stock"] = *req.Stock
	}
	if req.StartTime != nil {
		if t, parseErr := parseStartTime(*req.StartTime); parseErr == nil {
			updates["start_time"] = t
		} else {
			appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, "开始时间格式不正确")
			return
		}
	}
	if req.Image != nil {
		updates["image"] = *req.Image
	}

	if err := svc.UpdateProduct(userID, uint(id), updates); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			appG.Error(http.StatusNotFound, e.ERROR_NOT_EXIST_PRODUCT)
		} else {
			appG.Error(http.StatusInternalServerError, e.ERROR)
		}
		return
	}
	appG.Success(gin.H{"id": id})
}

// DeleteProduct 删除商品（仅创建者）
// @Summary 删除商品
// @Tags 商品
// @Produce json
// @Security BearerAuth
// @Param id path int true "商品ID"
// @Success 200 {object} app.Response{data=IDResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 404 {object} app.Response "未找到"
// @Router /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	appG := app.Gin{C: c}
	svc := h.svc.WithContext(c.Request.Context())
	uidAny, ok := c.Get("userID")
	if !ok {
		appG.Error(http.StatusUnauthorized, e.UNAUTHORIZED)
		return
	}
	userID := uidAny.(uint)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	if err := svc.DeleteProduct(userID, uint(id)); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			appG.Error(http.StatusNotFound, e.ERROR_NOT_EXIST_PRODUCT)
		} else {
			appG.Error(http.StatusInternalServerError, e.ERROR)
		}
		return
	}
	appG.Success(gin.H{"id": id})
}

// ListMyProducts 获取当前用户发布的商品
// @Summary 我的商品列表
// @Tags 商品
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param size query int false "每页条数" default(10)
// @Success 200 {object} app.Response{data=ProductListWithSizeResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Router /products/mine [get]
func (h *ProductHandler) ListMyProducts(c *gin.Context) {
	appG := app.Gin{C: c}
	svc := h.svc.WithContext(c.Request.Context())
	uidAny, ok := c.Get("userID")
	if !ok {
		appG.Error(http.StatusUnauthorized, e.UNAUTHORIZED)
		return
	}
	userID := uidAny.(uint)

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}
	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil || size <= 0 {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	list, total, err := svc.ListUserProducts(userID, page, size)
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.Success(gin.H{
		"items": list,
		"total": total,
		"page":  page,
		"size":  size,
	})
}
