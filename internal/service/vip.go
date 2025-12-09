package service

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PaidPlan 付费 VIP 套餐配置
type PaidPlan struct {
	PlanID       int
	Level        int   // VIP 等级
	DurationDays int   // 有效天数
	PriceCents   int64 // 价格（分）
}

// 简单预置两个付费 VIP 套餐，可按需扩展。
var paidPlans = map[int]PaidPlan{
	1: {PlanID: 1, Level: 3, DurationDays: 30, PriceCents: 3000}, // L3 30 天
	2: {PlanID: 2, Level: 4, DurationDays: 90, PriceCents: 8000}, // L4 90 天
}

// VIPProfile 用户 VIP 状态视图
type VIPProfile struct {
	TotalSpentCents int64     `json:"total_spent_cents"` // 累计消费（分）
	GrowthLevel     int       `json:"growth_level"`      // 成长等级（消费累计）
	PaidLevel       int       `json:"paid_level"`        // 付费等级
	PaidExpiredAt   time.Time `json:"paid_expired_at"`   // 付费到期时间
	EffectiveLevel  int       `json:"effective_level"`   // 生效等级 = max(成长, 付费)
}

// VIPService VIP 服务，处理等级查询和付费开通。
type VIPService struct {
	db          *gorm.DB
	userRepo    *repository.UserRepo
	paidVIPRepo *repository.PaidVIPRepo
	couponSvc   *CouponService
}

func NewVIPService(db *gorm.DB, userRepo *repository.UserRepo, couponSvc *CouponService) *VIPService {
	return &VIPService{
		db:          db,
		userRepo:    userRepo,
		paidVIPRepo: repository.NewPaidVIPRepo(db),
		couponSvc:   couponSvc,
	}
}

// Profile 查询用户 VIP 状态，合并成长等级与付费等级。
func (s *VIPService) Profile(ctx context.Context, userID uint) (*VIPProfile, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var paid *model.PaidVIP
	if pv, err := s.paidVIPRepo.GetByUser(ctx, userID); err == nil {
		paid = pv
	}

	profile := &VIPProfile{
		TotalSpentCents: user.TotalSpentCents,
		GrowthLevel:     user.GrowthLevel,
	}
	if paid != nil && paid.ExpiredAt.After(time.Now()) {
		profile.PaidLevel = paid.Level
		profile.PaidExpiredAt = paid.ExpiredAt
	}
	profile.EffectiveLevel = max(profile.GrowthLevel, profile.PaidLevel)
	return profile, nil
}

// PurchasePaidVIP 激活付费 VIP（模拟购买成功），当前直接落库，可结合支付单扩展。
func (s *VIPService) PurchasePaidVIP(ctx context.Context, userID uint, planID int) (*VIPProfile, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	plan, ok := paidPlans[planID]
	if !ok {
		return nil, fmt.Errorf("未知付费VIP套餐")
	}
	start := time.Now()
	end := start.Add(time.Duration(plan.DurationDays) * 24 * time.Hour)
	if err := s.paidVIPRepo.Upsert(ctx, userID, plan.Level, start, end); err != nil {
		return nil, err
	}
	// 购买成功后立即发放当月 VIP 优惠券
	if s.couponSvc != nil {
		_ = s.couponSvc.IssueVIPMonthly(ctx, userID, plan.Level)
	}
	return s.Profile(ctx, userID)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
