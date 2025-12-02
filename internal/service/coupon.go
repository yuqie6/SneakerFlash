package service

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// 业务错误定义
var (
	ErrCouponNotFound       = errors.New("优惠券不存在")
	ErrCouponNotAvailable   = errors.New("优惠券不可用")
	ErrCouponExpired        = errors.New("优惠券已过期")
	ErrCouponNotPurchasable = errors.New("该优惠券不支持购买")
	ErrCouponBelowThreshold = errors.New("订单金额未达到优惠券使用门槛")
	ErrCouponInvalidRate    = errors.New("优惠券折扣率无效")
	ErrCouponTypeInvalid    = errors.New("不支持的优惠券类型")
)

type CouponService struct {
	db             *gorm.DB
	couponRepo     *repository.CouponRepo
	userCouponRepo *repository.UserCouponRepo
}

func NewCouponService(db *gorm.DB) *CouponService {
	return &CouponService{
		db:             db,
		couponRepo:     repository.NewCouponRepo(db),
		userCouponRepo: repository.NewUserCouponRepo(db),
	}
}

type MyCoupon struct {
	ID            uint               `json:"id"`
	CouponID      uint               `json:"coupon_id"`
	Type          model.CouponType   `json:"type"`
	Title         string             `json:"title"`
	Description   string             `json:"description"`
	AmountCents   int64              `json:"amount_cents"`
	DiscountRate  int                `json:"discount_rate"`
	MinSpendCents int64              `json:"min_spend_cents"`
	Status        model.CouponStatus `json:"status"`
	ValidFrom     time.Time          `json:"valid_from"`
	ValidTo       time.Time          `json:"valid_to"`
	ObtainedFrom  string             `json:"obtained_from"`
}

var vipMonthlyQuota = map[int]int{
	1: 1,
	2: 2,
	3: 3,
	4: 4,
}

type vipCouponTemplate struct {
	Title         string
	Type          model.CouponType
	AmountCents   int64
	DiscountRate  int
	MinSpendCents int64
}

var vipTemplates = map[int]vipCouponTemplate{
	1: {Title: "VIP L1 月度券", Type: model.CouponTypeFullCut, AmountCents: 500, MinSpendCents: 3000},
	2: {Title: "VIP L2 月度券", Type: model.CouponTypeFullCut, AmountCents: 1000, MinSpendCents: 5000},
	3: {Title: "VIP L3 月度券", Type: model.CouponTypeDiscount, DiscountRate: 90, MinSpendCents: 0},
	4: {Title: "VIP L4 月度券", Type: model.CouponTypeDiscount, DiscountRate: 85, MinSpendCents: 0},
}

// ApplyCoupon 校验并计算优惠后的金额，返回优惠后金额和需要核销的用户券记录。
func (s *CouponService) ApplyCoupon(ctx context.Context, userID uint, userCouponID uint, originAmount int64) (*model.UserCoupon, *model.Coupon, int64, error) {
	if ctx == nil {
		return nil, nil, 0, fmt.Errorf("context is nil")
	}
	now := time.Now()

	// 查询并锁定用户券
	uc, err := s.userCouponRepo.GetByIDForUpdate(ctx, userCouponID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, 0, ErrCouponNotFound
		}
		return nil, nil, 0, err
	}

	// 业务校验：归属、状态、有效期
	if uc.UserID != userID {
		return nil, nil, 0, ErrCouponNotFound
	}
	if uc.Status != model.CouponStatusAvailable {
		return nil, nil, 0, ErrCouponNotAvailable
	}
	if now.Before(uc.ValidFrom) || now.After(uc.ValidTo) {
		return nil, nil, 0, ErrCouponExpired
	}

	// 读取券模板
	c, err := s.couponRepo.GetByID(ctx, uc.CouponID)
	if err != nil {
		return nil, nil, 0, err
	}

	// 校验门槛
	if originAmount < c.MinSpendCents {
		return nil, nil, 0, ErrCouponBelowThreshold
	}

	// 计算优惠后金额
	newAmount := originAmount
	switch c.Type {
	case model.CouponTypeFullCut:
		newAmount = originAmount - c.AmountCents
	case model.CouponTypeDiscount:
		if c.DiscountRate <= 0 || c.DiscountRate >= 100 {
			return nil, nil, 0, ErrCouponInvalidRate
		}
		newAmount = originAmount * int64(c.DiscountRate) / 100
	default:
		return nil, nil, 0, ErrCouponTypeInvalid
	}
	if newAmount < 0 {
		newAmount = 0
	}
	return uc, c, newAmount, nil
}

func (s *CouponService) MarkUsed(ctx context.Context, userCouponID uint, orderID uint) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return s.userCouponRepo.MarkUsed(ctx, userCouponID, orderID)
}

func (s *CouponService) ReleaseByOrder(ctx context.Context, orderID uint) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return s.userCouponRepo.ReleaseByOrder(ctx, orderID)
}

// ListUserCoupons 查询用户优惠券列表，支持分页。
func (s *CouponService) ListUserCoupons(ctx context.Context, userID uint, status string, page, pageSize int) ([]MyCoupon, int64, error) {
	if ctx == nil {
		return nil, 0, fmt.Errorf("context is nil")
	}
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	now := time.Now()

	ucs, total, err := s.userCouponRepo.ListByUserAndStatus(ctx, userID, status, now, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(ucs) == 0 {
		return nil, total, nil
	}

	ids := make([]uint, 0, len(ucs))
	for _, uc := range ucs {
		ids = append(ids, uc.CouponID)
	}

	cs, err := s.couponRepo.ListByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}

	cmap := make(map[uint]model.Coupon, len(cs))
	for _, c := range cs {
		cmap[c.ID] = c
	}

	out := make([]MyCoupon, 0, len(ucs))
	for _, uc := range ucs {
		c := cmap[uc.CouponID]
		// 实时修正状态：如果 status=available 但已过期，返回 expired
		effectiveStatus := uc.Status
		if uc.Status == model.CouponStatusAvailable && now.After(uc.ValidTo) {
			effectiveStatus = model.CouponStatusExpired
		}
		out = append(out, MyCoupon{
			ID:            uc.ID,
			CouponID:      uc.CouponID,
			Type:          c.Type,
			Title:         c.Title,
			Description:   c.Description,
			AmountCents:   c.AmountCents,
			DiscountRate:  c.DiscountRate,
			MinSpendCents: c.MinSpendCents,
			Status:        effectiveStatus,
			ValidFrom:     uc.ValidFrom,
			ValidTo:       uc.ValidTo,
			ObtainedFrom:  uc.ObtainedFrom,
		})
	}
	return out, total, nil
}

// PurchaseCoupon 购买优惠券，事务保护。
func (s *CouponService) PurchaseCoupon(ctx context.Context, userID, couponID uint) (*MyCoupon, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	var result *MyCoupon
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCouponRepo := repository.NewCouponRepo(tx)
		txUserCouponRepo := repository.NewUserCouponRepo(tx)

		c, err := txCouponRepo.GetByID(ctx, couponID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCouponNotFound
			}
			return err
		}
		if !c.Purchasable {
			return ErrCouponNotPurchasable
		}

		now := time.Now()
		uc := &model.UserCoupon{
			UserID:       userID,
			CouponID:     couponID,
			Status:       model.CouponStatusAvailable,
			ObtainedFrom: "purchase",
			ValidFrom:    c.ValidFrom,
			ValidTo:      c.ValidTo,
			IssuedAt:     now,
		}
		if err := txUserCouponRepo.Create(ctx, uc); err != nil {
			return err
		}

		result = &MyCoupon{
			ID:            uc.ID,
			CouponID:      uc.CouponID,
			Type:          c.Type,
			Title:         c.Title,
			Description:   c.Description,
			AmountCents:   c.AmountCents,
			DiscountRate:  c.DiscountRate,
			MinSpendCents: c.MinSpendCents,
			Status:        uc.Status,
			ValidFrom:     uc.ValidFrom,
			ValidTo:       uc.ValidTo,
			ObtainedFrom:  uc.ObtainedFrom,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// IssueVIPMonthly 按月配额为指定等级的用户发券（幂等：当月超配额不再发）。
func (s *CouponService) IssueVIPMonthly(ctx context.Context, userID uint, level int) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}

	if level < 1 {
		level = 1
	}
	if level > 4 {
		level = 4
	}
	quota := vipMonthlyQuota[level]
	if quota <= 0 {
		return nil
	}
	start, end := monthPeriod(time.Now())
	existing, err := s.userCouponRepo.CountByPeriod(ctx, userID, "vip_month", start, end)
	if err != nil {
		return err
	}
	if existing >= int64(quota) {
		return nil
	}
	tpl, ok := vipTemplates[level]
	if !ok {
		return nil
	}
	coupon, err := s.ensureTemplate(ctx, tpl)
	if err != nil {
		return err
	}
	need := quota - int(existing)
	now := time.Now()
	ucs := make([]model.UserCoupon, 0, need)
	for range need {
		ucs = append(ucs, model.UserCoupon{
			UserID:       userID,
			CouponID:     coupon.ID,
			Status:       model.CouponStatusAvailable,
			ObtainedFrom: "vip_month",
			ValidFrom:    start,
			ValidTo:      end,
			IssuedAt:     now,
		})
	}
	return s.userCouponRepo.BatchCreate(ctx, ucs)
}

// ensureTemplate 确保券模板存在，使用 FirstOrCreate 保证并发安全。
func (s *CouponService) ensureTemplate(ctx context.Context, tpl vipCouponTemplate) (*model.Coupon, error) {
	coupon := &model.Coupon{
		Title:         tpl.Title,
		Type:          tpl.Type,
		AmountCents:   tpl.AmountCents,
		DiscountRate:  tpl.DiscountRate,
		MinSpendCents: tpl.MinSpendCents,
		Purchasable:   false,
		ValidFrom:     time.Now().AddDate(-1, 0, 0),
		ValidTo:       time.Now().AddDate(1, 0, 0),
		Status:        model.CouponTemplateStatusActive,
	}
	return s.couponRepo.FirstOrCreate(ctx, coupon, "title = ?", tpl.Title)
}

func monthPeriod(now time.Time) (time.Time, time.Time) {
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0)
	return start, end
}

// MarkExpiredCoupons 批量将已过期但 status 仍为 available 的券标记为 expired。
func (s *CouponService) MarkExpiredCoupons(ctx context.Context) (int64, error) {
	if ctx == nil {
		return 0, fmt.Errorf("context is nil")
	}
	return s.userCouponRepo.MarkExpiredBatch(ctx, time.Now())
}
