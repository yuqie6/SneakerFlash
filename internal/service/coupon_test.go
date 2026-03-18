package service

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/testutil"
	"context"
	"errors"
	"testing"
	"time"
)

func newCouponServiceForTest(t *testing.T) (*CouponService, context.Context) {
	t.Helper()

	testutil.SetupTestConfig()
	db := testutil.NewSQLiteDB(t)
	return NewCouponService(db), context.Background()
}

func TestCouponService_ApplyCoupon(t *testing.T) {
	svc, ctx := newCouponServiceForTest(t)
	now := time.Now()

	fullCut := &model.Coupon{
		Type:          model.CouponTypeFullCut,
		Title:         "满减券",
		AmountCents:   500,
		MinSpendCents: 3000,
		ValidFrom:     now.Add(-time.Hour),
		ValidTo:       now.Add(time.Hour),
		Status:        model.CouponTemplateStatusActive,
	}
	discount := &model.Coupon{
		Type:          model.CouponTypeDiscount,
		Title:         "九折券",
		DiscountRate:  90,
		MinSpendCents: 0,
		ValidFrom:     now.Add(-time.Hour),
		ValidTo:       now.Add(time.Hour),
		Status:        model.CouponTemplateStatusActive,
	}

	if err := svc.couponRepo.Create(ctx, fullCut); err != nil {
		t.Fatalf("create full cut coupon: %v", err)
	}
	if err := svc.couponRepo.Create(ctx, discount); err != nil {
		t.Fatalf("create discount coupon: %v", err)
	}

	fullCutUC := &model.UserCoupon{
		UserID:       1,
		CouponID:     fullCut.ID,
		Status:       model.CouponStatusAvailable,
		ObtainedFrom: "purchase",
		ValidFrom:    now.Add(-time.Hour),
		ValidTo:      now.Add(time.Hour),
		IssuedAt:     now,
	}
	discountUC := &model.UserCoupon{
		UserID:       1,
		CouponID:     discount.ID,
		Status:       model.CouponStatusAvailable,
		ObtainedFrom: "purchase",
		ValidFrom:    now.Add(-time.Hour),
		ValidTo:      now.Add(time.Hour),
		IssuedAt:     now,
	}
	expiredUC := &model.UserCoupon{
		UserID:       1,
		CouponID:     fullCut.ID,
		Status:       model.CouponStatusAvailable,
		ObtainedFrom: "purchase",
		ValidFrom:    now.Add(-2 * time.Hour),
		ValidTo:      now.Add(-time.Hour),
		IssuedAt:     now.Add(-2 * time.Hour),
	}

	for _, uc := range []*model.UserCoupon{fullCutUC, discountUC, expiredUC} {
		if err := svc.userCouponRepo.Create(ctx, uc); err != nil {
			t.Fatalf("create user coupon: %v", err)
		}
	}

	t.Run("full cut success", func(t *testing.T) {
		uc, coupon, amount, err := svc.ApplyCoupon(ctx, 1, fullCutUC.ID, 5000)
		if err != nil {
			t.Fatalf("ApplyCoupon() error = %v", err)
		}
		if uc.ID != fullCutUC.ID || coupon.ID != fullCut.ID || amount != 4500 {
			t.Fatalf("ApplyCoupon() = uc=%d coupon=%d amount=%d", uc.ID, coupon.ID, amount)
		}
	})

	t.Run("discount success", func(t *testing.T) {
		_, _, amount, err := svc.ApplyCoupon(ctx, 1, discountUC.ID, 5000)
		if err != nil {
			t.Fatalf("ApplyCoupon() error = %v", err)
		}
		if amount != 4500 {
			t.Fatalf("discounted amount = %d, want 4500", amount)
		}
	})

	t.Run("below threshold", func(t *testing.T) {
		_, _, _, err := svc.ApplyCoupon(ctx, 1, fullCutUC.ID, 2000)
		if !errors.Is(err, ErrCouponBelowThreshold) {
			t.Fatalf("ApplyCoupon() error = %v, want %v", err, ErrCouponBelowThreshold)
		}
	})

	t.Run("expired", func(t *testing.T) {
		_, _, _, err := svc.ApplyCoupon(ctx, 1, expiredUC.ID, 5000)
		if !errors.Is(err, ErrCouponExpired) {
			t.Fatalf("ApplyCoupon() error = %v, want %v", err, ErrCouponExpired)
		}
	})
}

func TestCouponService_MarkExpiredCoupons(t *testing.T) {
	svc, ctx := newCouponServiceForTest(t)
	now := time.Now()

	coupon := &model.Coupon{
		Type:          model.CouponTypeFullCut,
		Title:         "过期券",
		AmountCents:   500,
		MinSpendCents: 0,
		ValidFrom:     now.Add(-24 * time.Hour),
		ValidTo:       now.Add(24 * time.Hour),
		Status:        model.CouponTemplateStatusActive,
	}
	if err := svc.couponRepo.Create(ctx, coupon); err != nil {
		t.Fatalf("create coupon: %v", err)
	}

	uc := &model.UserCoupon{
		UserID:       1,
		CouponID:     coupon.ID,
		Status:       model.CouponStatusAvailable,
		ObtainedFrom: "purchase",
		ValidFrom:    now.Add(-48 * time.Hour),
		ValidTo:      now.Add(-24 * time.Hour),
		IssuedAt:     now.Add(-48 * time.Hour),
	}
	if err := svc.userCouponRepo.Create(ctx, uc); err != nil {
		t.Fatalf("create user coupon: %v", err)
	}

	affected, err := svc.MarkExpiredCoupons(ctx)
	if err != nil {
		t.Fatalf("MarkExpiredCoupons() error = %v", err)
	}
	if affected != 1 {
		t.Fatalf("affected = %d, want 1", affected)
	}

	list, _, err := svc.ListUserCoupons(ctx, 1, string(model.CouponStatusExpired), 1, 10)
	if err != nil {
		t.Fatalf("ListUserCoupons() error = %v", err)
	}
	if len(list) != 1 || list[0].Status != model.CouponStatusExpired {
		t.Fatalf("expired coupons = %+v", list)
	}
}

func TestCouponService_DeleteTemplateInUse(t *testing.T) {
	svc, ctx := newCouponServiceForTest(t)
	now := time.Now()

	coupon := &model.Coupon{
		Type:          model.CouponTypeFullCut,
		Title:         "不可删除券",
		AmountCents:   500,
		MinSpendCents: 1000,
		ValidFrom:     now.Add(-time.Hour),
		ValidTo:       now.Add(time.Hour),
		Status:        model.CouponTemplateStatusActive,
	}
	if err := svc.couponRepo.Create(ctx, coupon); err != nil {
		t.Fatalf("create coupon: %v", err)
	}

	uc := &model.UserCoupon{
		UserID:       1,
		CouponID:     coupon.ID,
		Status:       model.CouponStatusAvailable,
		ObtainedFrom: "purchase",
		ValidFrom:    now.Add(-time.Hour),
		ValidTo:      now.Add(time.Hour),
		IssuedAt:     now,
	}
	if err := svc.userCouponRepo.Create(ctx, uc); err != nil {
		t.Fatalf("create user coupon: %v", err)
	}

	if err := svc.DeleteTemplate(ctx, coupon.ID); !errors.Is(err, ErrCouponTemplateInUse) {
		t.Fatalf("DeleteTemplate() error = %v, want %v", err, ErrCouponTemplateInUse)
	}
}

func TestCouponService_CreateTemplateRejectsInvalidStatus(t *testing.T) {
	svc, ctx := newCouponServiceForTest(t)
	now := time.Now()

	_, err := svc.CreateTemplate(ctx, CouponTemplateInput{
		Type:          model.CouponTypeFullCut,
		Title:         "状态非法",
		AmountCents:   500,
		MinSpendCents: 1000,
		ValidFrom:     now.Add(-time.Hour),
		ValidTo:       now.Add(time.Hour),
		Status:        "enabled",
	})
	if !errors.Is(err, ErrCouponTemplateStatus) {
		t.Fatalf("CreateTemplate() error = %v, want %v", err, ErrCouponTemplateStatus)
	}
}
