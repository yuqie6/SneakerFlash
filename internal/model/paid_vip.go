package model

import "time"

// PaidVIP 记录付费 VIP 的等级与有效期。
type PaidVIP struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex" json:"user_id"`
	Level     int       `gorm:"not null" json:"level"`
	StartedAt time.Time `gorm:"not null" json:"started_at"`
	ExpiredAt time.Time `gorm:"not null" json:"expired_at"`
}

func (PaidVIP) TableName() string {
	return "paid_vips"
}
