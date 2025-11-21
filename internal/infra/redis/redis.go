package redis

import (
	"SneakerFlash/internal/config"
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func Init(cfg config.RedisConfig) {
	RDB = redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdle,

		DialTimeout:  time.Duration(cfg.ConnTimeout) * time.Second,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
	})

	// 启动前 ping 测试
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := RDB.Ping(ctx).Result(); err != nil {
		log.Fatalf("连接 redis 失败: %s", err)
	}

	log.Println("redis 初始化成功")
}
