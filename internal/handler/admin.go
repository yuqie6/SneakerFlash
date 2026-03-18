package handler

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/service"
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	adminSvc  *service.AdminService
	riskSvc   *service.RiskService
	couponSvc *service.CouponService
}

type riskEntryReq struct {
	Type  string `json:"type" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type adminCouponCreateReq struct {
	Type          string `json:"type" binding:"required"`
	Title         string `json:"title" binding:"required"`
	Description   string `json:"description"`
	AmountCents   int64  `json:"amount_cents"`
	DiscountRate  int    `json:"discount_rate"`
	MinSpendCents int64  `json:"min_spend_cents"`
	ValidFrom     string `json:"valid_from" binding:"required"`
	ValidTo       string `json:"valid_to" binding:"required"`
	Purchasable   bool   `json:"purchasable"`
	PriceCents    int64  `json:"price_cents"`
	Status        string `json:"status"`
}

type adminCouponUpdateReq struct {
	Type          *string `json:"type"`
	Title         *string `json:"title"`
	Description   *string `json:"description"`
	AmountCents   *int64  `json:"amount_cents"`
	DiscountRate  *int    `json:"discount_rate"`
	MinSpendCents *int64  `json:"min_spend_cents"`
	ValidFrom     *string `json:"valid_from"`
	ValidTo       *string `json:"valid_to"`
	Purchasable   *bool   `json:"purchasable"`
	PriceCents    *int64  `json:"price_cents"`
	Status        *string `json:"status"`
}

func NewAdminHandler(adminSvc *service.AdminService, riskSvc *service.RiskService, couponSvc *service.CouponService) *AdminHandler {
	return &AdminHandler{
		adminSvc:  adminSvc,
		riskSvc:   riskSvc,
		couponSvc: couponSvc,
	}
}

func (h *AdminHandler) Stats(c *gin.Context) {
	appG := app.Gin{C: c}
	stats, err := h.adminSvc.Stats(c.Request.Context())
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.Success(stats)
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	appG := app.Gin{C: c}
	page, pageSize, ok := parsePage(c)
	if !ok {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	users, total, err := h.adminSvc.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.SuccessWithPage(users, total, page, pageSize)
}

func (h *AdminHandler) ListOrders(c *gin.Context) {
	appG := app.Gin{C: c}
	page, pageSize, ok := parsePage(c)
	if !ok {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	var status *model.OrderStatus
	if rawStatus := c.Query("status"); rawStatus != "" {
		parsed, err := strconv.Atoi(rawStatus)
		if err != nil {
			appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
			return
		}
		orderStatus := model.OrderStatus(parsed)
		status = &orderStatus
	}

	orders, total, err := h.adminSvc.ListAllOrders(c.Request.Context(), status, page, pageSize)
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.SuccessWithPage(orders, total, page, pageSize)
}

func (h *AdminHandler) ListCoupons(c *gin.Context) {
	appG := app.Gin{C: c}
	page, pageSize, ok := parsePage(c)
	if !ok {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	coupons, total, err := h.couponSvc.ListTemplates(c.Request.Context(), page, pageSize)
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.SuccessWithPage(coupons, total, page, pageSize)
}

func (h *AdminHandler) CreateCoupon(c *gin.Context) {
	appG := app.Gin{C: c}

	var req adminCouponCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	validFrom, err := parseAdminTime(req.ValidFrom)
	if err != nil {
		appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, "valid_from 格式不正确")
		return
	}
	validTo, err := parseAdminTime(req.ValidTo)
	if err != nil {
		appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, "valid_to 格式不正确")
		return
	}

	coupon, err := h.couponSvc.CreateTemplate(c.Request.Context(), service.CouponTemplateInput{
		Type:          model.CouponType(strings.TrimSpace(req.Type)),
		Title:         req.Title,
		Description:   req.Description,
		AmountCents:   req.AmountCents,
		DiscountRate:  req.DiscountRate,
		MinSpendCents: req.MinSpendCents,
		ValidFrom:     validFrom,
		ValidTo:       validTo,
		Purchasable:   req.Purchasable,
		PriceCents:    req.PriceCents,
		Status:        req.Status,
	})
	if err != nil {
		if isCouponClientError(err) {
			appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, err.Error())
			return
		}
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.Success(coupon)
}

func (h *AdminHandler) UpdateCoupon(c *gin.Context) {
	appG := app.Gin{C: c}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	var req adminCouponUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	patch := service.CouponTemplatePatch{
		Title:         req.Title,
		Description:   req.Description,
		AmountCents:   req.AmountCents,
		DiscountRate:  req.DiscountRate,
		MinSpendCents: req.MinSpendCents,
		Purchasable:   req.Purchasable,
		PriceCents:    req.PriceCents,
		Status:        req.Status,
	}
	if req.Type != nil {
		couponType := model.CouponType(strings.TrimSpace(*req.Type))
		patch.Type = &couponType
	}
	if req.ValidFrom != nil {
		validFrom, err := parseAdminTime(*req.ValidFrom)
		if err != nil {
			appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, "valid_from 格式不正确")
			return
		}
		patch.ValidFrom = &validFrom
	}
	if req.ValidTo != nil {
		validTo, err := parseAdminTime(*req.ValidTo)
		if err != nil {
			appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, "valid_to 格式不正确")
			return
		}
		patch.ValidTo = &validTo
	}

	coupon, err := h.couponSvc.UpdateTemplate(c.Request.Context(), uint(id), patch)
	if err != nil {
		if errors.Is(err, service.ErrCouponNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			appG.Error(http.StatusNotFound, e.INVALID_PARAMS)
			return
		}
		if isCouponClientError(err) {
			appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, err.Error())
			return
		}
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.Success(coupon)
}

func (h *AdminHandler) DeleteCoupon(c *gin.Context) {
	appG := app.Gin{C: c}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	if err := h.couponSvc.DeleteTemplate(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, service.ErrCouponNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			appG.Error(http.StatusNotFound, e.INVALID_PARAMS)
			return
		}
		if errors.Is(err, service.ErrCouponTemplateInUse) {
			appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, err.Error())
			return
		}
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.Success(gin.H{"message": "ok"})
}

func (h *AdminHandler) ListProducts(c *gin.Context) {
	appG := app.Gin{C: c}
	page, pageSize, ok := parsePage(c)
	if !ok {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	products, total, err := h.adminSvc.ListAllProducts(c.Request.Context(), page, pageSize)
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.SuccessWithPage(products, total, page, pageSize)
}

func (h *AdminHandler) ListBlacklist(c *gin.Context) {
	h.listRisk(c, h.riskSvc.ListBlacklist)
}

func (h *AdminHandler) AddBlacklist(c *gin.Context) {
	h.changeRisk(c, h.riskSvc.AddBlacklist)
}

func (h *AdminHandler) RemoveBlacklist(c *gin.Context) {
	h.changeRisk(c, h.riskSvc.RemoveBlacklist)
}

func (h *AdminHandler) ListGraylist(c *gin.Context) {
	h.listRisk(c, h.riskSvc.ListGraylist)
}

func (h *AdminHandler) AddGraylist(c *gin.Context) {
	h.changeRisk(c, h.riskSvc.AddGraylist)
}

func (h *AdminHandler) RemoveGraylist(c *gin.Context) {
	h.changeRisk(c, h.riskSvc.RemoveGraylist)
}

func (h *AdminHandler) listRisk(c *gin.Context, fn func(context.Context) (ips, users []string, err error)) {
	appG := app.Gin{C: c}
	ips, users, err := fn(c.Request.Context())
	if err != nil {
		if errors.Is(err, service.ErrRiskEntryTypeInvalid) {
			appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, err.Error())
			return
		}
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.Success(gin.H{"ip": ips, "user": users})
}

func (h *AdminHandler) changeRisk(c *gin.Context, fn func(context.Context, string, string) error) {
	appG := app.Gin{C: c}

	var req riskEntryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	if err := fn(c.Request.Context(), req.Type, req.Value); err != nil {
		if errors.Is(err, service.ErrRiskEntryTypeInvalid) {
			appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, err.Error())
			return
		}
		if strings.Contains(err.Error(), "不能为空") {
			appG.ErrorMsg(http.StatusBadRequest, e.INVALID_PARAMS, err.Error())
			return
		}
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.Success(gin.H{"message": "ok"})
}

func parsePage(c *gin.Context) (int, int, bool) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		return 0, 0, false
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || pageSize <= 0 {
		return 0, 0, false
	}
	return page, pageSize, true
}

func parseAdminTime(raw string) (time.Time, error) {
	zonedLayouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, layout := range zonedLayouts {
		if parsed, err := time.Parse(layout, strings.TrimSpace(raw)); err == nil {
			return parsed, nil
		}
	}

	localLayouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
	}

	for _, layout := range localLayouts {
		if parsed, err := time.ParseInLocation(layout, strings.TrimSpace(raw), time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errors.New("invalid time format")
}

func isCouponClientError(err error) bool {
	return errors.Is(err, service.ErrCouponTitleRequired) ||
		errors.Is(err, service.ErrCouponTemplateStatus) ||
		errors.Is(err, service.ErrCouponInvalidAmount) ||
		errors.Is(err, service.ErrCouponInvalidPeriod) ||
		errors.Is(err, service.ErrCouponInvalidRate) ||
		errors.Is(err, service.ErrCouponTypeInvalid) ||
		errors.Is(err, service.ErrCouponTemplateInUse)
}
