package main

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/db"
	"SneakerFlash/internal/infra/kafka"
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"
)

func main() {
	config.Init()

	db.Init(config.Conf.Data.Database)
	redis.Init(config.Conf.Data.Redis)

	productRepo := repository.NewProductRepo(db.DB)
	orderRepo := repository.NewOrderRepo(db.DB)

	workerSvc := service.NewWorkerService(db.DB, productRepo, orderRepo)

	kafka.StartConsumer(config.Conf.Data.Kafka, workerSvc.CreateOderFromMessage)
}
