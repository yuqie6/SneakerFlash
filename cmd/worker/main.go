package main

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/db"
	"SneakerFlash/internal/infra/kafka"
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"
	"log"
)

func main() {
	config.Init()

	db.Init(config.Conf.Data.Database)
	redis.Init(config.Conf.Data.Redis)
	if err := utils.InitSnowflake(int64(config.Conf.Server.MachineID)); err != nil {
		log.Fatalf("[ERROR] 初始化雪花算法失败: %v", err)
	}

	productRepo := repository.NewProductRepo(db.DB)
	orderRepo := repository.NewOrderRepo(db.DB)

	workerSvc := service.NewWorkerService(db.DB, productRepo, orderRepo)

	kafka.StartConsumer(config.Conf.Data.Kafka, workerSvc.CreateOderFromMessage)
}
