package model

import "time"

type AuditLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	ActorID      uint      `gorm:"index;not null" json:"actor_id"`
	ActorName    string    `gorm:"type:varchar(64);not null" json:"actor_name"`
	ActorRole    string    `gorm:"type:varchar(32);not null" json:"actor_role"`
	Resource     string    `gorm:"type:varchar(32);index;not null" json:"resource"`
	Action       string    `gorm:"type:varchar(32);index;not null" json:"action"`
	ResourceID   string    `gorm:"type:varchar(64);default:''" json:"resource_id"`
	RequestID    string    `gorm:"type:varchar(64);default:''" json:"request_id"`
	RequestPath  string    `gorm:"type:varchar(255);default:''" json:"request_path"`
	RequestIP    string    `gorm:"type:varchar(64);default:''" json:"request_ip"`
	RequestBody  string    `gorm:"type:text" json:"request_body"`
	Result       string    `gorm:"type:varchar(16);not null" json:"result"`
	ErrorMessage string    `gorm:"type:varchar(255);default:''" json:"error_message"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
