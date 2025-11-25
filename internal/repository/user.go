package repository

import (
	"SneakerFlash/internal/model"
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 构建用户仓储。
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) DB() *gorm.DB {
	return r.db
}

// WithContext 绑定请求上下文，供用户查询/更新日志关联 request_id。
func (r *UserRepo) WithContext(ctx context.Context) *UserRepo {
	if ctx == nil {
		return r
	}
	return &UserRepo{db: r.db.WithContext(ctx)}
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

// 更新用户基础资料
func (r *UserRepo) UpdateProfile(uid uint, values map[string]any) error {
	if len(values) == 0 {
		return nil
	}
	return r.db.Model(&model.User{}).Where("id = ?", uid).Updates(values).Error
}

// GetByIDForUpdate 查询并加行级锁，避免并发成长值更新丢失。
func (r *UserRepo) GetByIDForUpdate(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateGrowth 同步更新累计实付与成长等级。
func (r *UserRepo) UpdateGrowth(userID uint, totalSpentCents int64, growthLevel int) error {
	updates := map[string]any{
		"total_spent_cents": totalSpentCents,
		"growth_level":      growthLevel,
	}
	return r.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}
