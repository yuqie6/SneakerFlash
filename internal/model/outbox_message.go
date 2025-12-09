package model

import (
	"time"
)

// OutboxStatus 本地消息表状态
type OutboxStatus int

const (
	OutboxStatusPending OutboxStatus = 0 // 待发送
	OutboxStatusSent    OutboxStatus = 1 // 已发送
	OutboxStatusFailed  OutboxStatus = 2 // 发送失败（达到最大重试）
)

// OutboxMessage 本地消息表模型，用于实现 Transactional Outbox Pattern
type OutboxMessage struct {
	ID         uint         `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
	Topic      string       `gorm:"type:varchar(128);not null;index" json:"topic"`
	Payload    string       `gorm:"type:text;not null" json:"payload"` // JSON 消息体
	Status     OutboxStatus `gorm:"default:0;index" json:"status"`
	RetryCount int          `gorm:"default:0" json:"retry_count"`
	LastError  string       `gorm:"type:varchar(512)" json:"last_error"`
	SentAt     *time.Time   `gorm:"index" json:"sent_at"` // 发送成功时间
}

func (OutboxMessage) TableName() string {
	return "outbox_messages"
}
