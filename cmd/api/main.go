package main

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/infra/db"
	"SneakerFlash/internal/server"
)

func main() {
	config.Init()
	db.Init(config.Conf.Data.Database)

	r := server.NewHttpServer()
	r.Run()
}
