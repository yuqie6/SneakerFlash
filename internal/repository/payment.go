package repository

import (
	"SneakerFlash/internal/model"
	"time"

	"gorm.io/gorm"
)

type PaymentRepo struct {
	db *gorm.DB
}

func NewPaymentRepo(db *gorm.DB) *PaymentRepo {
	return &PaymentRepo{
		db: db,
	}
}

// 增加订单
func (r *PaymentRepo) Create(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

// 根据支付号查订单
func (r *PaymentRepo) GetByPaymentID(pid uint) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.Where("payment_id = ?", pid).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// 根据订单号查订单
func (r *PaymentRepo) GetByOrderID(oid uint) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.Where("order_id = ?", oid).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// 根据支付id 获取待支付的订单记录
func (r *PaymentRepo) GetPendingByOrderID(orderID string) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.Where("order_id = ? AND status = ?", orderID, model.PaymentStatusPending).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// 更新订单状态
func (r *PaymentRepo) UpdateStatus(pid uint, status model.PaymentStatus, notifyData string) error {
	updates := map[string]any{
		"status":     status,
		"updated_at": time.Now(),
	}
	if notifyData != "" {
		updates["notify_data"] = notifyData
	}
	return r.db.Model(&model.Payment{}).Where("id = ?", pid).Updates(updates).Error
}

// 根据订单 id 更新状态
func (r *PaymentRepo) UpdateStatusByPaymentID(paymentID string, status model.PaymentStatus, notifyData string) error {
	updates := map[string]any{
		"status":     status,
		"updated_at": time.Now(),
	}
	if notifyData != "" {
		updates["notify_data"] = notifyData
	}
	return r.db.Model(&model.Payment{}).Where("payment_id = ?", paymentID).Updates(updates).Error
}
