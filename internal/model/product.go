package model

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"uniqueIndex:idx_user_name_deleted" json:"-"`
	UserID    uint           `gorm:"not null;uniqueIndex:idx_user_name_deleted" json:"user_id"`
	Name      string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_user_name_deleted" json:"name"`
	Price     float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock     int            `gorm:"not null" json:"stock"`
	StartTime time.Time      `gorm:"not null" json:"start_time"`
	EndTime   *time.Time     `json:"end_time"` // 可选，NULL 表示永不过期
	Image     string         `gorm:"type:varchar(255)" json:"image"`
}

func (Product) TableName() string {
	return "products"
}
