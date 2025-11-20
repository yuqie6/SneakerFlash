package main

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/infra/db"
	"SneakerFlash/internal/infra/kafka"
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/server"
)

func main() {
	config.Init()
	db.Init(config.Conf.Data.Database)
	redis.Init(config.Conf.Data.Redis)
	kafka.InitProducer(config.Conf.Data.Kafka)

	r := server.NewHttpServer()
	r.Run(config.Conf.Server.Port)
}
