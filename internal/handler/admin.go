package handler

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"SneakerFlash/internal/pkg/logger"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"
	"context"
	"errors"
	"log/slog"
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
	auditSvc  *service.AuditService
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

func NewAdminHandler(adminSvc *service.AdminService, riskSvc *service.RiskService, couponSvc *service.CouponService, auditSvc *service.AuditService) *AdminHandler {
	return &AdminHandler{
		adminSvc:  adminSvc,
		riskSvc:   riskSvc,
		couponSvc: couponSvc,
		auditSvc:  auditSvc,
	}
}

// Stats 管理台总览统计
// @Summary 管理台统计
// @Tags 管理后台
// @Produce json
// @Security BearerAuth
// @Success 200 {object} app.Response{data=service.AdminStats}
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Router /admin/stats [get]
func (h *AdminHandler) Stats(c *gin.Context) {
	appG := app.Gin{C: c}
	stats, err := h.adminSvc.Stats(c.Request.Context())
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.Success(stats)
}

// ListUsers 管理台用户列表
// @Summary 管理台用户列表
// @Tags 管理后台
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Success 200 {object} app.Response{data=app.PageData}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Router /admin/users [get]
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

// ListOrders 管理台订单列表
// @Summary 管理台订单列表
// @Tags 管理后台
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Param status query int false "订单状态：0未支付 1已支付 2失败 3已取消"
// @Success 200 {object} app.Response{data=app.PageData}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Router /admin/orders [get]
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
		if !model.ValidOrderStatus(orderStatus) {
			appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
			return
		}
		status = &orderStatus
	}

	orders, total, err := h.adminSvc.ListAllOrders(c.Request.Context(), status, page, pageSize)
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.SuccessWithPage(orders, total, page, pageSize)
}

// ListCoupons 管理台优惠券模板列表
// @Summary 管理台优惠券列表
// @Tags 管理后台
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Success 200 {object} app.Response{data=app.PageData}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Router /admin/coupons [get]
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

// CreateCoupon 管理台创建优惠券模板
// @Summary 创建优惠券模板
// @Tags 管理后台
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body adminCouponCreateReq true "优惠券模板"
// @Success 200 {object} app.Response{data=model.Coupon}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Router /admin/coupons [post]
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
	h.recordAudit(c, model.AdminResourceCoupons, "create", strconv.Itoa(int(coupon.ID)), req, "")
	appG.Success(coupon)
}

// UpdateCoupon 管理台更新优惠券模板
// @Summary 更新优惠券模板
// @Tags 管理后台
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "优惠券ID"
// @Param payload body adminCouponUpdateReq true "优惠券模板补丁"
// @Success 200 {object} app.Response{data=model.Coupon}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Failure 404 {object} app.Response "资源不存在"
// @Router /admin/coupons/{id} [put]
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
	h.recordAudit(c, model.AdminResourceCoupons, "update", strconv.Itoa(int(coupon.ID)), req, "")
	appG.Success(coupon)
}

// DeleteCoupon 管理台删除优惠券模板
// @Summary 删除优惠券模板
// @Tags 管理后台
// @Produce json
// @Security BearerAuth
// @Param id path int true "优惠券ID"
// @Success 200 {object} app.Response{data=MessageResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Failure 404 {object} app.Response "资源不存在"
// @Router /admin/coupons/{id} [delete]
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
	h.recordAudit(c, model.AdminResourceCoupons, "delete", strconv.Itoa(id), nil, "")
	appG.Success(gin.H{"message": "ok"})
}

// ListProducts 管理台商品列表
// @Summary 管理台商品列表
// @Tags 管理后台
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Success 200 {object} app.Response{data=app.PageData}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Router /admin/products [get]
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

// ListBlacklist 管理台黑名单
// @Summary 查询黑名单
// @Tags 管理后台
// @Produce json
// @Security BearerAuth
// @Success 200 {object} app.Response{data=RiskListResponse}
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Router /admin/risk/blacklist [get]
func (h *AdminHandler) ListBlacklist(c *gin.Context) {
	h.listRisk(c, h.riskSvc.ListBlacklist)
}

// AddBlacklist 管理台新增黑名单
// @Summary 新增黑名单
// @Tags 管理后台
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body riskEntryReq true "名单项"
// @Success 200 {object} app.Response{data=MessageResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Router /admin/risk/blacklist [post]
func (h *AdminHandler) AddBlacklist(c *gin.Context) {
	h.changeRisk(c, "create", model.AdminResourceRisk, h.riskSvc.AddBlacklist)
}

// RemoveBlacklist 管理台删除黑名单
// @Summary 删除黑名单
// @Tags 管理后台
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body riskEntryReq true "名单项"
// @Success 200 {object} app.Response{data=MessageResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Router /admin/risk/blacklist [delete]
func (h *AdminHandler) RemoveBlacklist(c *gin.Context) {
	h.changeRisk(c, "delete", model.AdminResourceRisk, h.riskSvc.RemoveBlacklist)
}

// ListGraylist 管理台灰名单
// @Summary 查询灰名单
// @Tags 管理后台
// @Produce json
// @Security BearerAuth
// @Success 200 {object} app.Response{data=RiskListResponse}
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Router /admin/risk/graylist [get]
func (h *AdminHandler) ListGraylist(c *gin.Context) {
	h.listRisk(c, h.riskSvc.ListGraylist)
}

// AddGraylist 管理台新增灰名单
// @Summary 新增灰名单
// @Tags 管理后台
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body riskEntryReq true "名单项"
// @Success 200 {object} app.Response{data=MessageResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Router /admin/risk/graylist [post]
func (h *AdminHandler) AddGraylist(c *gin.Context) {
	h.changeRisk(c, "create", model.AdminResourceRisk, h.riskSvc.AddGraylist)
}

// RemoveGraylist 管理台删除灰名单
// @Summary 删除灰名单
// @Tags 管理后台
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body riskEntryReq true "名单项"
// @Success 200 {object} app.Response{data=MessageResponse}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Router /admin/risk/graylist [delete]
func (h *AdminHandler) RemoveGraylist(c *gin.Context) {
	h.changeRisk(c, "delete", model.AdminResourceRisk, h.riskSvc.RemoveGraylist)
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

func (h *AdminHandler) changeRisk(c *gin.Context, action, resource string, fn func(context.Context, string, string) error) {
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
	h.recordAudit(c, resource, action, req.Value, req, "")
	appG.Success(gin.H{"message": "ok"})
}

// ListAuditLogs 管理台审计日志列表
// @Summary 审计日志列表
// @Tags 管理后台
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Param actor_name query string false "操作者用户名"
// @Param resource query string false "资源类型"
// @Param action query string false "动作"
// @Success 200 {object} app.Response{data=app.PageData}
// @Failure 400 {object} app.Response "参数错误"
// @Failure 401 {object} app.Response "未登录"
// @Failure 403 {object} app.Response "需要管理员权限"
// @Router /admin/audit [get]
func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	appG := app.Gin{C: c}
	page, pageSize, ok := parsePage(c)
	if !ok {
		appG.Error(http.StatusBadRequest, e.INVALID_PARAMS)
		return
	}

	logs, total, err := h.auditSvc.List(c.Request.Context(), repository.AuditLogFilter{
		ActorName: strings.TrimSpace(c.Query("actor_name")),
		Resource:  strings.TrimSpace(c.Query("resource")),
		Action:    strings.TrimSpace(c.Query("action")),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		appG.Error(http.StatusInternalServerError, e.ERROR)
		return
	}
	appG.SuccessWithPage(logs, total, page, pageSize)
}

func (h *AdminHandler) recordAudit(c *gin.Context, resource, action, resourceID string, body any, errMsg string) {
	if h.auditSvc == nil {
		return
	}
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	role, _ := c.Get("role")
	requestID, _ := c.Get("request_id")
	actorID, okID := userID.(uint)
	actorName, okName := username.(string)
	actorRole, okRole := role.(string)
	requestIDStr, _ := requestID.(string)
	if !okID || !okName || !okRole {
		slog.WarnContext(c.Request.Context(), "审计日志缺少操作者上下文", slog.String("resource", resource), slog.String("action", action))
		return
	}
	result := "success"
	if errMsg != "" {
		result = "failed"
	}
	if err := h.auditSvc.Record(c.Request.Context(), service.AuditLogInput{
		ActorID:      actorID,
		ActorName:    actorName,
		ActorRole:    actorRole,
		Resource:     resource,
		Action:       action,
		ResourceID:   resourceID,
		RequestID:    requestIDStr,
		RequestPath:  c.FullPath(),
		RequestIP:    c.ClientIP(),
		RequestBody:  body,
		Result:       result,
		ErrorMessage: errMsg,
	}); err != nil {
		slog.WarnContext(
			logger.ContextWithAttrs(c.Request.Context(), slog.String("request_id", requestIDStr)),
			"写入审计日志失败",
			slog.String("resource", resource),
			slog.String("action", action),
			slog.Any("err", err),
		)
	}
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
