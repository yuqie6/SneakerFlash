package repository

import (
	"SneakerFlash/internal/model"
	"context"
	"time"

	"gorm.io/gorm"
)

type PaidVIPRepo struct {
	db *gorm.DB
}

func NewPaidVIPRepo(db *gorm.DB) *PaidVIPRepo {
	return &PaidVIPRepo{db: db}
}

// GetByUser 查询当前付费 VIP 记录。
func (r *PaidVIPRepo) GetByUser(ctx context.Context, userID uint) (*model.PaidVIP, error) {
	var pv model.PaidVIP
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&pv).Error; err != nil {
		return nil, err
	}
	return &pv, nil
}

// Upsert 覆盖/新增付费 VIP。
func (r *PaidVIPRepo) Upsert(ctx context.Context, userID uint, level int, start, end time.Time) error {
	pv := model.PaidVIP{
		UserID:    userID,
		Level:     level,
		StartedAt: start,
		ExpiredAt: end,
	}
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Assign(pv).
		FirstOrCreate(&pv).Error
}
