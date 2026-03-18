package cron

import (
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

const (
	defaultOrderCancelInterval = 30 * time.Second
	defaultOrderTimeout        = 15 * time.Minute
)

type OrderCancelCron struct {
	orderSvc *service.OrderService
	stopCh   chan struct{}
}

func NewOrderCancelCron(db *gorm.DB) *OrderCancelCron {
	return &OrderCancelCron{
		orderSvc: service.NewOrderService(db, repository.NewProductRepo(db), repository.NewUserRepo(db)),
		stopCh:   make(chan struct{}),
	}
}

func (c *OrderCancelCron) Start() {
	ticker := time.NewTicker(defaultOrderCancelInterval)
	slog.Info("未支付订单自动取消任务已启动",
		slog.Duration("interval", defaultOrderCancelInterval),
		slog.Duration("timeout", defaultOrderTimeout),
	)

	go func() {
		for {
			select {
			case <-ticker.C:
				cancelled, err := c.orderSvc.CancelExpiredOrders(context.Background(), defaultOrderTimeout, 100)
				if err != nil {
					slog.Error("未支付订单自动取消失败", slog.Any("err", err))
					continue
				}
				if cancelled > 0 {
					slog.Info("未支付订单自动取消完成", slog.Int("cancelled", cancelled))
				}
			case <-c.stopCh:
				ticker.Stop()
				slog.Info("未支付订单自动取消任务停止")
				return
			}
		}
	}()
}

func (c *OrderCancelCron) Stop() {
	close(c.stopCh)
}
