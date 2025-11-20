package main

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/infra/db"
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/server"
)

func main() {
	config.Init()
	db.Init(config.Conf.Data.Database)
	redis.Init(config.Conf.Data.Redis)

	r := server.NewHttpServer()
	r.Run()
}
