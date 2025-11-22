package model

import (
	"time"

	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusFaild    PaymentStatus = "faild"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type Payment struct {
	gorm.Model
	OrderID    uint          `gorm:"not null;uniqueIdex" json:"order_id"`
	PaymentID  string        `gorm:"type:varchar(64);unique;not null" json:"payment_id"`
	Amount     float64       `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status     PaymentStatus `gorm:"varchar(20);default:'pending'" json:"status"`
	NotifyData string        `gorm:"type:varchar(20)" json:"notify_data"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

func (Payment) TableName() string {
	return "payments"
}
