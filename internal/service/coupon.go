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
func (s *CouponService) ApplyCoupon(ctx context.Context, userID uint, couponID uint, originAmount int64) (*model.UserCoupon, *model.Coupon, int64, error) {
	if ctx == nil {
		return nil, nil, 0, fmt.Errorf("context is nil")
	}
	now := time.Now()
	userCouponRepo := s.userCouponRepo.WithContext(ctx)
	// 查询并锁定用户券 + 读取券模板
	uc, c, err := userCouponRepo.GetUsableForUpdate(userID, couponID, now)
	if err != nil {
		return nil, nil, 0, err
	}
	if originAmount < c.MinSpendCents {
		return nil, nil, 0, errors.New("amount below coupon threshold")
	}
	newAmount := originAmount
	switch c.Type {
	case model.CouponTypeFullCut:
		newAmount = originAmount - c.AmountCents
	case model.CouponTypeDiscount:
		if c.DiscountRate <= 0 || c.DiscountRate >= 100 {
			return nil, nil, 0, errors.New("invalid discount rate")
		}
		newAmount = originAmount * int64(c.DiscountRate) / 100
	default:
		return nil, nil, 0, errors.New("unsupported coupon type")
	}
	if newAmount < 0 {
		newAmount = 0 // 优惠后不允许负数
	}
	return uc, c, newAmount, nil
}

func (s *CouponService) MarkUsed(ctx context.Context, userCouponID uint, orderID uint) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return s.userCouponRepo.WithContext(ctx).MarkUsed(userCouponID, orderID)
}

func (s *CouponService) ReleaseByOrder(ctx context.Context, orderID uint) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return s.userCouponRepo.WithContext(ctx).ReleaseByOrder(orderID)
}

func (s *CouponService) ListUserCoupons(ctx context.Context, userID uint, status string) ([]MyCoupon, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	var ucs []model.UserCoupon
	q := s.userCouponRepo.WithContext(ctx).DB().Where("user_id = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Order("id desc").Find(&ucs).Error; err != nil {
		return nil, err
	}
	if len(ucs) == 0 {
		return nil, nil
	}
	ids := make([]uint, 0, len(ucs))
	for _, uc := range ucs {
		ids = append(ids, uc.CouponID)
	}
	var cs []model.Coupon
	if err := s.couponRepo.WithContext(ctx).DB().Where("id IN ?", ids).Find(&cs).Error; err != nil {
		return nil, err
	}
	cmap := make(map[uint]model.Coupon, len(cs))
	for _, c := range cs {
		cmap[c.ID] = c
	}
	out := make([]MyCoupon, 0, len(ucs))
	for _, uc := range ucs {
		c := cmap[uc.CouponID]
		out = append(out, MyCoupon{
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
		})
	}
	return out, nil
}

// PurchaseCoupon 模拟购买：直接发一张券给用户（券需标记为 purchasable）。
func (s *CouponService) PurchaseCoupon(ctx context.Context, userID, couponID uint) (*MyCoupon, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	couponRepo := s.couponRepo.WithContext(ctx)
	userCouponRepo := s.userCouponRepo.WithContext(ctx)

	c, err := couponRepo.GetByID(couponID)
	if err != nil {
		return nil, err
	}
	if !c.Purchasable {
		return nil, errors.New("coupon not purchasable")
	}
	now := time.Now()
	uc := model.UserCoupon{
		UserID:       userID,
		CouponID:     couponID,
		Status:       model.CouponStatusAvailable,
		ObtainedFrom: "purchase",
		ValidFrom:    c.ValidFrom,
		ValidTo:      c.ValidTo,
		IssuedAt:     now,
	}
	if err := userCouponRepo.DB().Create(&uc).Error; err != nil {
		return nil, err
	}
	return &MyCoupon{
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
	}, nil
}

// IssueVIPMonthly 按月配额为指定等级的用户发券（幂等：当月超配额不再发）。
func (s *CouponService) IssueVIPMonthly(ctx context.Context, userID uint, level int) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	userCouponRepo := s.userCouponRepo.WithContext(ctx)

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
	existing, err := userCouponRepo.CountByPeriod(userID, "vip_month", start, end)
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
	for i := 0; i < need; i++ {
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
	return userCouponRepo.DB().Create(&ucs).Error
}

func (s *CouponService) ensureTemplate(ctx context.Context, tpl vipCouponTemplate) (*model.Coupon, error) {
	couponRepo := s.couponRepo.WithContext(ctx)
	var c model.Coupon
	err := couponRepo.DB().
		Where("title = ?", tpl.Title).
		First(&c).Error
	if err == nil {
		return &c, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	c = model.Coupon{
		Title:         tpl.Title,
		Type:          tpl.Type,
		AmountCents:   tpl.AmountCents,
		DiscountRate:  tpl.DiscountRate,
		MinSpendCents: tpl.MinSpendCents,
		Purchasable:   false,
		ValidFrom:     time.Now().AddDate(-1, 0, 0),
		ValidTo:       time.Now().AddDate(1, 0, 0),
		Status:        "active",
	}
	if err := couponRepo.DB().Create(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func monthPeriod(now time.Time) (time.Time, time.Time) {
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0)
	return start, end
}
