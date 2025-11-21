package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string  `gorm:"type:varchar(50);unique;not null" json:"username"`
	Password string  `gorm:"type:varchar(100);not null" json:"-"`
	Balance  float64 `gorm:"type:decimal(10,2);default:0;not null" json:"balance"`
}

func (User) TableName() string {
	return "users"
}
