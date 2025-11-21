package main

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/db"
	"SneakerFlash/internal/infra/kafka"
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/server"
	"log"
)

func main() {
	config.Init()

	db.Init(config.Conf.Data.Database)
	redis.Init(config.Conf.Data.Redis)
	kafka.InitProducer(config.Conf.Data.Kafka)

	db.MakeMigrate()

	r := server.NewHttpServer()
	if err := r.Run(config.Conf.Server.Port); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
