package service

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type PaidPlan struct {
	PlanID       int
	Level        int
	DurationDays int
	PriceCents   int64
}

// 简单预置两个付费 VIP 套餐，可按需扩展。
var paidPlans = map[int]PaidPlan{
	1: {PlanID: 1, Level: 3, DurationDays: 30, PriceCents: 3000}, // L3 30 天
	2: {PlanID: 2, Level: 4, DurationDays: 90, PriceCents: 8000}, // L4 90 天
}

type VIPProfile struct {
	TotalSpentCents int64     `json:"total_spent_cents"`
	GrowthLevel     int       `json:"growth_level"`
	PaidLevel       int       `json:"paid_level"`
	PaidExpiredAt   time.Time `json:"paid_expired_at"`
	EffectiveLevel  int       `json:"effective_level"`
}

type VIPService struct {
	db          *gorm.DB
	userRepo    *repository.UserRepo
	paidVIPRepo *repository.PaidVIPRepo
}

func NewVIPService(db *gorm.DB, userRepo *repository.UserRepo) *VIPService {
	return &VIPService{
		db:          db,
		userRepo:    userRepo,
		paidVIPRepo: repository.NewPaidVIPRepo(db),
	}
}

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
	return s.Profile(ctx, userID)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
