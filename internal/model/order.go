package model

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus int

const (
	OrderStatusUnpaid OrderStatus = 0
	OrderStatusPaid   OrderStatus = 1
	OrderStatusFailed OrderStatus = 2
)

type Order struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `gorm:"not null;index;idx_user_product,unique" json:"user_id"`
	ProductID uint           `gorm:"not null;index:idx_user_product,unique" json:"product_id"`
	OrderNum  string         `gorm:"type:varchar(32);unique;not null" json:"order_num"`
	Status    OrderStatus    `gorm:"default:0" json:"status"`
}

func (Order) TableName() string {
	return "orders"
}
