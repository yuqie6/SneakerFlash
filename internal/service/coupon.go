package service

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"context"
	"errors"
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

func (s *CouponService) WithContext(ctx context.Context) *CouponService {
	if ctx == nil {
		return s
	}
	ctxDB := s.db.WithContext(ctx)
	return &CouponService{
		db:             ctxDB,
		couponRepo:     s.couponRepo.WithContext(ctx),
		userCouponRepo: s.userCouponRepo.WithContext(ctx),
	}
}

// ApplyCoupon 校验并计算优惠后的金额，返回优惠后金额和需要核销的用户券记录。
func (s *CouponService) ApplyCoupon(userID uint, couponID uint, originAmount int64) (*model.UserCoupon, *model.Coupon, int64, error) {
	now := time.Now()
	// 查询并锁定用户券 + 读取券模板
	uc, c, err := s.userCouponRepo.GetUsableForUpdate(userID, couponID, now)
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

func (s *CouponService) MarkUsed(userCouponID uint, orderID uint) error {
	return s.userCouponRepo.MarkUsed(userCouponID, orderID)
}

func (s *CouponService) ReleaseByOrder(orderID uint) error {
	return s.userCouponRepo.ReleaseByOrder(orderID)
}
