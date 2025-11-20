package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name      string  `gorm:"not null"`
	Price     float64 `gorm:"not null"`
	Stock     int
	StartTime time.Time
}
