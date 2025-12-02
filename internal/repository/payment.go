package repository

import (
	"SneakerFlash/internal/model"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type PaymentRepo struct {
	db *gorm.DB
}

// NewPaymentRepo 构建支付仓储。
func NewPaymentRepo(db *gorm.DB) *PaymentRepo {
	return &PaymentRepo{
		db: db,
	}
}

// 基于 order_id 幂等创建支付单；已存在则直接返回
func (r *PaymentRepo) CreateIfAbsent(ctx context.Context, payment *model.Payment) (*model.Payment, error) {
	db := r.db.WithContext(ctx)
	var existing model.Payment
	err := db.Where("order_id = ?", payment.OrderID).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := db.Create(payment).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			if err := db.Where("order_id = ?", payment.OrderID).First(&existing).Error; err == nil {
				return &existing, nil
			}
		}
		return nil, err
	}
	return payment, nil
}

// 根据支付号查支付单
func (r *PaymentRepo) GetByPaymentID(ctx context.Context, pid string) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.WithContext(ctx).Where("payment_id = ?", pid).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// 根据订单号查支付单（单订单唯一支付单）
func (r *PaymentRepo) GetByOrderID(ctx context.Context, orderID uint) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// UpdateAmountIfPending 更新支付金额，仅在待支付状态下生效。
func (r *PaymentRepo) UpdateAmountIfPending(ctx context.Context, orderID uint, amountCents int64) (int64, error) {
	tx := r.db.WithContext(ctx).Model(&model.Payment{}).
		Where("order_id = ? AND status = ?", orderID, model.PaymentStatusPending).
		Updates(map[string]any{
			"amount_cents": amountCents,
			"updated_at":   time.Now(),
		})
	return tx.RowsAffected, tx.Error
}

// 条件更新支付状态（按支付号+当前状态），用于回调幂等；返回影响行数
func (r *PaymentRepo) UpdateStatusByPaymentIDIfMatch(ctx context.Context, paymentID string, fromStatus model.PaymentStatus, toStatus model.PaymentStatus, notifyData string) (int64, error) {
	updates := map[string]any{
		"status":     toStatus,
		"updated_at": time.Now(),
	}
	if notifyData != "" {
		updates["notify_data"] = notifyData
	}

	tx := r.db.WithContext(ctx).Model(&model.Payment{}).Where("payment_id = ? AND status = ?", paymentID, fromStatus).Updates(updates)
	return tx.RowsAffected, tx.Error
}
