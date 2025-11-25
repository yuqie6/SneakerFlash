// @title SneakerFlash API
// @version 1.0
// @description SneakerFlash 球鞋秒杀系统接口文档
// @BasePath /api/v1
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"log/slog"
	"os"

	docs "SneakerFlash/docs"

	"SneakerFlash/internal/config"
	"SneakerFlash/internal/db"
	"SneakerFlash/internal/infra/kafka"
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/pkg/logger"
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/server"
)

func main() {
	config.Init()
	logger.InitLogger(config.Conf.Logger, "api")

	db.Init(config.Conf.Data.Database)
	redis.Init(config.Conf.Data.Redis)
	kafka.InitProducer(config.Conf.Data.Kafka)

	db.MakeMigrate()

	if err := utils.InitSnowflake(int64(config.Conf.Server.MachineID)); err != nil {
		slog.Error("初始化雪花算法失败", slog.Any("err", err))
		os.Exit(1)
	}

	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Title = "SneakerFlash API"
	docs.SwaggerInfo.Description = "SneakerFlash 球鞋秒杀系统接口文档"
	docs.SwaggerInfo.Version = "1.0"

	r := server.NewHttpServer()
	if err := r.Run(config.Conf.Server.Port); err != nil {
		slog.Error("启动失败", slog.Any("err", err))
		os.Exit(1)
	}
}
