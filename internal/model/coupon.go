package model

import (
	"time"
)

type CouponType string

const (
	CouponTypeFullCut  CouponType = "full_cut" // 满减
	CouponTypeDiscount CouponType = "discount" // 折扣
)

type CouponStatus string

const (
	CouponStatusAvailable CouponStatus = "available"
	CouponStatusUsed      CouponStatus = "used"
	CouponStatusExpired   CouponStatus = "expired"
)

// 券模板状态
const (
	CouponTemplateStatusActive   = "active"
	CouponTemplateStatusInactive = "inactive"
)

type Coupon struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Type          CouponType `gorm:"type:varchar(20);not null" json:"type"`
	Title         string     `gorm:"type:varchar(100);not null" json:"title"`
	Description   string     `gorm:"type:varchar(255);default:''" json:"description"`
	AmountCents   int64      `gorm:"default:0;not null" json:"amount_cents"`    // 满减金额（分）
	DiscountRate  int        `gorm:"default:0;not null" json:"discount_rate"`   // 折扣百分比，90 表示 9 折
	MinSpendCents int64      `gorm:"default:0;not null" json:"min_spend_cents"` // 使用门槛（分）
	ValidFrom     time.Time  `json:"valid_from"`
	ValidTo       time.Time  `json:"valid_to"`
	Purchasable   bool       `gorm:"default:false" json:"purchasable"`      // 是否可购买
	PriceCents    int64      `gorm:"default:0;not null" json:"price_cents"` // 购买价格（分）
	Status        string     `gorm:"type:varchar(20);default:'active'" json:"status"`
}

func (Coupon) TableName() string {
	return "coupons"
}

type UserCoupon struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	UserID       uint         `gorm:"not null;index" json:"user_id"`
	CouponID     uint         `gorm:"not null;index" json:"coupon_id"`
	Status       CouponStatus `gorm:"type:varchar(20);not null;default:'available';index" json:"status"`
	ObtainedFrom string       `gorm:"type:varchar(50);default:''" json:"obtained_from"`
	ValidFrom    time.Time    `json:"valid_from"`
	ValidTo      time.Time    `json:"valid_to"`
	OrderID      *uint        `gorm:"index" json:"order_id,omitempty"`
	IssuedAt     time.Time    `json:"issued_at"`
}

func (UserCoupon) TableName() string {
	return "user_coupons"
}
