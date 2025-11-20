package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string  `gorm:"not null"`
	Password string  `gorm:"not null"`
	Balance  float64 `gorm:"not null"`
}
