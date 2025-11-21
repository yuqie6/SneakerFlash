package repository

import (
	"SneakerFlash/internal/model"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

// 创建一个用户
func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// 通过用户名得到用户信息
func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// 根据 id 得到用户信息
func (r *UserRepo) GetByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// 更新用户密码
func (r *UserRepo) UpdatePassword(uid uint, newPassword string) error {
	return r.db.Model(&model.User{}).Where("id = ?", uid).Update("password", newPassword).Error
}
