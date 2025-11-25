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

// WithContext 绑定请求上下文，确保支付操作日志能关联到请求链路。
func (r *PaymentRepo) WithContext(ctx context.Context) *PaymentRepo {
	if ctx == nil {
		return r
	}
	return &PaymentRepo{db: r.db.WithContext(ctx)}
}

// 基于 order_id 幂等创建支付单；已存在则直接返回
func (r *PaymentRepo) CreateIfAbsent(payment *model.Payment) (*model.Payment, error) {
	var existing model.Payment
	err := r.db.Where("order_id = ?", payment.OrderID).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := r.db.Create(payment).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			if err := r.db.Where("order_id = ?", payment.OrderID).First(&existing).Error; err == nil {
				return &existing, nil
			}
		}
		return nil, err
	}
	return payment, nil
}

// 根据支付号查支付单
func (r *PaymentRepo) GetByPaymentID(pid string) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.Where("payment_id = ?", pid).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// 根据订单号查支付单（单订单唯一支付单）
func (r *PaymentRepo) GetByOrderID(orderID uint) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// 条件更新支付状态（按支付号+当前状态），用于回调幂等；返回影响行数
func (r *PaymentRepo) UpdateStatusByPaymentIDIfMatch(paymentID string, fromStatus model.PaymentStatus, toStatus model.PaymentStatus, notifyData string) (int64, error) {
	updates := map[string]any{
		"status":     toStatus,
		"updated_at": time.Now(),
	}
	if notifyData != "" {
		updates["notify_data"] = notifyData
	}

	tx := r.db.Model(&model.Payment{}).Where("payment_id = ? AND status = ?", paymentID, fromStatus).Updates(updates)
	return tx.RowsAffected, tx.Error
}
