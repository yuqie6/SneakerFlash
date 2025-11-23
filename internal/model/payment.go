package model

import (
	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type Payment struct {
	gorm.Model
	OrderID     uint          `gorm:"not null;uniqueIndex" json:"order_id"`
	PaymentID   string        `gorm:"type:varchar(64);unique;not null" json:"payment_id"`
	AmountCents int64         `gorm:"not null" json:"amount_cents"`
	Status      PaymentStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	NotifyData  string        `gorm:"type:varchar(20)" json:"notify_data"`
}

func (Payment) TableName() string {
	return "payments"
}
