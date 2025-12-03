package main

import (
	"log/slog"
	"os"

	"SneakerFlash/internal/config"
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
	if err := utils.InitSnowflake(int64(config.Conf.Server.MachineID)); err != nil {
		slog.Error("初始化雪花算法失败", slog.Any("err", err))
		os.Exit(1)
	}

	productRepo := repository.NewProductRepo(db.DB)
	orderRepo := repository.NewOrderRepo(db.DB)

	workerSvc := service.NewWorkerService(db.DB, productRepo, orderRepo)

	// 使用批量消费模式，大幅提升 TPS
	kafka.StartBatchConsumer(config.Conf.Data.Kafka, workerSvc.BatchCreateOrdersFromMessages)
}
