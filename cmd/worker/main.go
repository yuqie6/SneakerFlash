package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"SneakerFlash/internal/config"
	"SneakerFlash/internal/cron"
	"SneakerFlash/internal/db"
	"SneakerFlash/internal/infra/kafka"
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/pkg/logger"
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"
)

func main() {
	config.Init()
	logger.InitLogger(config.Conf.Logger, "worker")

	db.Init(config.Conf.Data.Database)
	redis.Init(config.Conf.Data.Redis)
	kafka.InitProducer(config.Conf.Data.Kafka) // 初始化 Kafka 生产者（用于 Outbox 补偿和 DLQ）

	if err := utils.InitSnowflake(int64(config.Conf.Server.MachineID)); err != nil {
		slog.Error("初始化雪花算法失败", slog.Any("err", err))
		os.Exit(1)
	}

	productRepo := repository.NewProductRepo(db.DB)
	orderRepo := repository.NewOrderRepo(db.DB)

	workerSvc := service.NewWorkerService(db.DB, productRepo, orderRepo)
	orderCancelCron := cron.NewOrderCancelCron(db.DB)

	// 启动 VIP 月度发券定时任务
	vipCron := cron.NewVIPCouponCron(db.DB)
	vipCron.Start()
	defer vipCron.Stop()

	// 启动 Outbox 补偿定时任务
	outboxCron := cron.NewOutboxCron(db.DB, config.Conf.Data.Kafka)
	outboxCron.Start()
	defer outboxCron.Stop()
	orderCancelCron.Start()
	defer orderCancelCron.Stop()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := kafka.StartBatchConsumer(ctx, config.Conf.Data.Kafka, workerSvc.BatchCreateOrdersFromMessages); err != nil {
		slog.Error("worker 消费异常退出", slog.Any("err", err))
		os.Exit(1)
	}

	if err := kafka.CloseProducer(); err != nil {
		slog.Warn("关闭 Kafka 失败", slog.Any("err", err))
	}
	if err := redis.Close(); err != nil {
		slog.Warn("关闭 Redis 失败", slog.Any("err", err))
	}
	if err := db.Close(); err != nil {
		slog.Warn("关闭数据库失败", slog.Any("err", err))
	}
}
