package models

import (
	"gorm.io/gorm"
)

type OrderStatus int

const (
	StatusUpaid  OrderStatus = 0
	statusPaid   OrderStatus = 1
	statusFailed OrderStatus = 2
)

type Order struct {
	gorm.Model
	UserID    uint        `gorm:"not null;index;idx_user_product,unique" json:"user_id"`
	ProductID uint        `gorm:"not null;index:idx_user_product,unique" json:"product_id"`
	OrderNum  string      `gorm:"type:varchar(32);unique;not null" json:"order_num"`
	Status    OrderStatus `gorm:"default:0" json:"status"`
}

func (Order) TableName() string {
	return "orders"
}
