package cron

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/infra/kafka"
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

// OutboxCron 本地消息表补偿定时任务
type OutboxCron struct {
	outboxRepo *repository.OutboxRepo
	maxRetries int
	cfg        config.KafkaConfig
	stopCh     chan struct{}
}

// NewOutboxCron 创建补偿定时任务
func NewOutboxCron(db *gorm.DB, cfg config.KafkaConfig) *OutboxCron {
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	return &OutboxCron{
		outboxRepo: repository.NewOutboxRepo(db),
		maxRetries: maxRetries,
		cfg:        cfg,
		stopCh:     make(chan struct{}),
	}
}

// Start 启动补偿任务
func (c *OutboxCron) Start() {
	scanInterval := c.cfg.OutboxScanInterval
	if scanInterval <= 0 {
		scanInterval = 30
	}

	ticker := time.NewTicker(time.Duration(scanInterval) * time.Second)
	slog.Info("Outbox 补偿任务启动", slog.Int("scan_interval_sec", scanInterval))

	go func() {
		for {
			select {
			case <-ticker.C:
				c.compensate()
				c.cleanupOldMessages()
			case <-c.stopCh:
				ticker.Stop()
				slog.Info("Outbox 补偿任务停止")
				return
			}
		}
	}()
}

// Stop 停止补偿任务
func (c *OutboxCron) Stop() {
	close(c.stopCh)
}

// compensate 补偿未发送的消息
func (c *OutboxCron) compensate() {
	ctx := context.Background()
	timeout := time.Duration(c.cfg.OutboxTimeout) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	msgs, err := c.outboxRepo.GetPendingMessages(ctx, timeout, 100)
	if err != nil {
		slog.Error("获取待补偿消息失败", slog.Any("error", err))
		return
	}

	if len(msgs) == 0 {
		return
	}

	slog.Info("发现待补偿消息", slog.Int("count", len(msgs)))

	for _, msg := range msgs {
		c.processMessage(ctx, msg)
	}
}

// processMessage 处理单条补偿消息
func (c *OutboxCron) processMessage(ctx context.Context, msg *model.OutboxMessage) {
	// 检查是否达到最大重试次数
	if msg.RetryCount >= c.maxRetries {
		slog.Warn("消息达到最大重试次数，标记为失败",
			slog.Uint64("msg_id", uint64(msg.ID)),
			slog.Int("retry_count", msg.RetryCount))

		if err := c.outboxRepo.MarkFailed(ctx, msg.ID, "达到最大重试次数"); err != nil {
			slog.Error("标记消息失败状态失败", slog.Uint64("msg_id", uint64(msg.ID)), slog.Any("error", err))
		}

		// 投递到死信队列
		if c.cfg.DLQTopic != "" {
			dlqMsg := kafka.NewDLQMessage(msg.Topic, []byte(msg.Payload), msg.RetryCount, "达到最大重试次数")
			if err := kafka.SendToDLQ(c.cfg.DLQTopic, dlqMsg); err != nil {
				slog.Error("投递死信队列失败", slog.Uint64("msg_id", uint64(msg.ID)), slog.Any("error", err))
			}
		}
		return
	}

	// 尝试发送消息
	slog.Info("重试发送消息",
		slog.Uint64("msg_id", uint64(msg.ID)),
		slog.Int("retry_count", msg.RetryCount))

	if err := kafka.Send(msg.Topic, msg.Payload); err != nil {
		slog.Warn("消息发送失败，稍后重试",
			slog.Uint64("msg_id", uint64(msg.ID)),
			slog.Any("error", err))

		if incrErr := c.outboxRepo.IncrRetry(ctx, msg.ID, err.Error()); incrErr != nil {
			slog.Error("增加重试次数失败", slog.Uint64("msg_id", uint64(msg.ID)), slog.Any("error", incrErr))
		}
		return
	}

	// 发送成功，标记为已发送
	if err := c.outboxRepo.MarkSent(ctx, msg.ID); err != nil {
		slog.Error("标记消息发送成功失败", slog.Uint64("msg_id", uint64(msg.ID)), slog.Any("error", err))
		return
	}

	slog.Info("消息补偿发送成功", slog.Uint64("msg_id", uint64(msg.ID)))
}

// cleanupOldMessages 清理已发送成功的旧消息
func (c *OutboxCron) cleanupOldMessages() {
	ctx := context.Background()
	rowsAffected, err := c.outboxRepo.CleanupOldMessages(ctx, 7) // 保留 7 天
	if err != nil {
		slog.Error("清理旧消息失败", slog.Any("error", err))
		return
	}
	if rowsAffected > 0 {
		slog.Info("清理旧消息完成", slog.Int64("deleted", rowsAffected))
	}
}
