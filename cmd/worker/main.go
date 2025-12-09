package main

import (
	"log/slog"
	"os"

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

	// 启动 VIP 月度发券定时任务
	vipCron := cron.NewVIPCouponCron(db.DB)
	vipCron.Start()
	defer vipCron.Stop()

	// 启动 Outbox 补偿定时任务
	outboxCron := cron.NewOutboxCron(db.DB, config.Conf.Data.Kafka)
	outboxCron.Start()
	defer outboxCron.Stop()

	// 使用批量消费模式，大幅提升 TPS
	kafka.StartBatchConsumer(config.Conf.Data.Kafka, workerSvc.BatchCreateOrdersFromMessages)
}
