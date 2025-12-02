package repository

import (
	"SneakerFlash/internal/model"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CouponRepo struct {
	db *gorm.DB
}

func NewCouponRepo(db *gorm.DB) *CouponRepo {
	return &CouponRepo{db: db}
}

func (r *CouponRepo) DB() *gorm.DB {
	return r.db
}

func (r *CouponRepo) GetByID(ctx context.Context, id uint) (*model.Coupon, error) {
	var c model.Coupon
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

type UserCouponRepo struct {
	db *gorm.DB
}

func NewUserCouponRepo(db *gorm.DB) *UserCouponRepo {
	return &UserCouponRepo{db: db}
}

func (r *UserCouponRepo) DB() *gorm.DB {
	return r.db
}

// GetUsableForUpdate 查询可用券并加锁。
func (r *UserCouponRepo) GetUsableForUpdate(ctx context.Context, userID, userCouponID uint, now time.Time) (*model.UserCoupon, *model.Coupon, error) {
	db := r.db.WithContext(ctx)
	// 行级锁避免并发核销同一张券
	var uc model.UserCoupon
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND user_id = ?", userCouponID, userID).
		First(&uc).Error; err != nil {
		return nil, nil, err
	}
	if uc.Status != model.CouponStatusAvailable {
		return nil, nil, errors.New("coupon not available")
	}
	if now.Before(uc.ValidFrom) || now.After(uc.ValidTo) {
		return nil, nil, errors.New("coupon expired")
	}
	var c model.Coupon
	// 读券模板信息（金额/折扣/门槛）
	if err := db.First(&c, uc.CouponID).Error; err != nil {
		return nil, nil, err
	}
	return &uc, &c, nil
}

// MarkUsed 标记卷为已使用
func (r *UserCouponRepo) MarkUsed(ctx context.Context, userCouponID uint, orderID uint) error {
	return r.db.WithContext(ctx).Model(&model.UserCoupon{}).
		Where("id = ?", userCouponID).
		Updates(map[string]any{
			"status":   model.CouponStatusUsed,
			"order_id": orderID,
		}).Error
}

// ReleaseByOrder 将订单占用的券恢复可用（用于支付失败）。
func (r *UserCouponRepo) ReleaseByOrder(ctx context.Context, orderID uint) error {
	return r.db.WithContext(ctx).Model(&model.UserCoupon{}).
		Where("order_id = ? AND status = ?", orderID, model.CouponStatusUsed).
		Updates(map[string]any{
			"status":   model.CouponStatusAvailable,
			"order_id": nil,
		}).Error
}

// GetByOrderID 查询绑定到订单的券记录。
func (r *UserCouponRepo) GetByOrderID(ctx context.Context, orderID uint) (*model.UserCoupon, error) {
	var uc model.UserCoupon
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&uc).Error; err != nil {
		return nil, err
	}
	return &uc, nil
}

// CountByPeriod 统计某个来源在周期内已发放数量。
func (r *UserCouponRepo) CountByPeriod(ctx context.Context, userID uint, obtainedFrom string, start, end time.Time) (int64, error) {
	var cnt int64
	err := r.db.WithContext(ctx).Model(&model.UserCoupon{}).
		Where("user_id = ? AND obtained_from = ? AND valid_from >= ? AND valid_from < ?", userID, obtainedFrom, start, end).
		Count(&cnt).Error
	return cnt, err
}

// MarkExpiredBatch 批量将已过期但 status 仍为 available 的券标记为 expired。
func (r *UserCouponRepo) MarkExpiredBatch(ctx context.Context, now time.Time) (int64, error) {
	tx := r.db.WithContext(ctx).Model(&model.UserCoupon{}).
		Where("status = ? AND valid_to < ?", model.CouponStatusAvailable, now).
		Update("status", model.CouponStatusExpired)
	return tx.RowsAffected, tx.Error
}
