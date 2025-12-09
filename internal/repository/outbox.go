package repository

import (
	"SneakerFlash/internal/model"
	"context"
	"time"

	"gorm.io/gorm"
)

// OutboxRepo 本地消息表仓储
type OutboxRepo struct {
	db *gorm.DB
}

// NewOutboxRepo 创建 OutboxRepo 实例
func NewOutboxRepo(db *gorm.DB) *OutboxRepo {
	return &OutboxRepo{db: db}
}

// Create 创建待发送消息
func (r *OutboxRepo) Create(ctx context.Context, msg *model.OutboxMessage) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

// CreateInTx 在事务中创建待发送消息
func (r *OutboxRepo) CreateInTx(ctx context.Context, tx *gorm.DB, msg *model.OutboxMessage) error {
	return tx.WithContext(ctx).Create(msg).Error
}

// MarkSent 标记为已发送
func (r *OutboxRepo) MarkSent(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.OutboxMessage{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":  model.OutboxStatusSent,
		"sent_at": now,
	}).Error
}

// MarkFailed 标记为发送失败
func (r *OutboxRepo) MarkFailed(ctx context.Context, id uint, errMsg string) error {
	return r.db.WithContext(ctx).Model(&model.OutboxMessage{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     model.OutboxStatusFailed,
		"last_error": errMsg,
	}).Error
}

// GetPendingMessages 获取超时未发送的消息（用于补偿）
func (r *OutboxRepo) GetPendingMessages(ctx context.Context, timeout time.Duration, limit int) ([]*model.OutboxMessage, error) {
	var msgs []*model.OutboxMessage
	cutoff := time.Now().Add(-timeout)
	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", model.OutboxStatusPending, cutoff).
		Order("created_at ASC").
		Limit(limit).
		Find(&msgs).Error
	return msgs, err
}

// IncrRetry 增加重试次数并记录错误
func (r *OutboxRepo) IncrRetry(ctx context.Context, id uint, lastErr string) error {
	return r.db.WithContext(ctx).Model(&model.OutboxMessage{}).Where("id = ?", id).Updates(map[string]interface{}{
		"retry_count": gorm.Expr("retry_count + ?", 1),
		"last_error":  lastErr,
	}).Error
}

// CleanupOldMessages 清理已发送超过指定天数的消息
func (r *OutboxRepo) CleanupOldMessages(ctx context.Context, days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days)
	result := r.db.WithContext(ctx).
		Where("status = ? AND sent_at < ?", model.OutboxStatusSent, cutoff).
		Delete(&model.OutboxMessage{})
	return result.RowsAffected, result.Error
}

// GetByID 根据 ID 获取消息
func (r *OutboxRepo) GetByID(ctx context.Context, id uint) (*model.OutboxMessage, error) {
	var msg model.OutboxMessage
	err := r.db.WithContext(ctx).First(&msg, id).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
