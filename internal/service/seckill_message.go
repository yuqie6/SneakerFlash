package service

import "time"

// SeckillMessage 描述秒杀队列消息，入口与 worker 共用，避免消息格式漂移。
type SeckillMessage struct {
	UserID     uint      `json:"user_id"`
	ProductID  uint      `json:"product_id"`
	OrderNum   string    `json:"order_num"`
	PaymentID  string    `json:"payment_id"`
	PriceCents int64     `json:"price_cents"`
	Time       time.Time `json:"time"`
}
