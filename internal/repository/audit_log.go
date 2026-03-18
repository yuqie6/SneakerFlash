package repository

import (
	"SneakerFlash/internal/model"
	"context"

	"gorm.io/gorm"
)

type AuditLogRepo struct {
	db *gorm.DB
}

func NewAuditLogRepo(db *gorm.DB) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

func (r *AuditLogRepo) Create(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *AuditLogRepo) List(ctx context.Context, filter AuditLogFilter) ([]model.AuditLog, int64, error) {
	var (
		list  []model.AuditLog
		total int64
	)

	query := r.db.WithContext(ctx).Model(&model.AuditLog{})
	if filter.ActorName != "" {
		query = query.Where("actor_name = ?", filter.ActorName)
	}
	if filter.Resource != "" {
		query = query.Where("resource = ?", filter.Resource)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order("id desc").Offset(offset).Limit(filter.PageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

type AuditLogFilter struct {
	ActorName string
	Resource  string
	Action    string
	Page      int
	PageSize  int
}
