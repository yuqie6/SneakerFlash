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

func (r *PaidVIPRepo) WithContext(ctx context.Context) *PaidVIPRepo {
	if ctx == nil {
		return r
	}
	return &PaidVIPRepo{db: r.db.WithContext(ctx)}
}

// GetByUser 查询当前付费 VIP 记录。
func (r *PaidVIPRepo) GetByUser(userID uint) (*model.PaidVIP, error) {
	var pv model.PaidVIP
	if err := r.db.Where("user_id = ?", userID).First(&pv).Error; err != nil {
		return nil, err
	}
	return &pv, nil
}

// Upsert 覆盖/新增付费 VIP。
func (r *PaidVIPRepo) Upsert(userID uint, level int, start, end time.Time) error {
	pv := model.PaidVIP{
		UserID:    userID,
		Level:     level,
		StartedAt: start,
		ExpiredAt: end,
	}
	return r.db.
		Where("user_id = ?", userID).
		Assign(pv).
		FirstOrCreate(&pv).Error
}
