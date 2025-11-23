package model

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_product_name" json:"name"`
	Price     float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock     int            `gorm:"not null" json:"stock"`
	StartTime time.Time      `gorm:"not null" json:"start_time"`
	Image     string         `gorm:"type:varchar(255)" json:"image"`
}

func (Product) TableName() string {
	return "products"
}
