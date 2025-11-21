package main

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/db"
	"SneakerFlash/internal/infra/redis"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config.Init()

	db.Init(config.Conf.Data.Database)

	redis.Init(config.Conf.Data.Redis)

	// 启动消费者逻辑
	log.Println("worker 启动中, 正在监听 kafka")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("worker 退出")
}
