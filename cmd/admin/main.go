package main

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/db"
	"SneakerFlash/internal/pkg/logger"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"
	"context"
	"flag"
	"log/slog"
	"os"
)

func main() {
	username := flag.String("username", "", "需要提权为管理员的用户名")
	flag.Parse()

	if *username == "" {
		slog.Error("缺少必填参数", slog.String("flag", "username"))
		os.Exit(1)
	}

	config.Init()
	logger.InitLogger(config.Conf.Logger, "admin-cli")
	db.Init(config.Conf.Data.Database)
	db.MakeMigrate()

	userSvc := service.NewUserService(repository.NewUserRepo(db.DB))
	if err := userSvc.PromoteToAdmin(context.Background(), *username); err != nil {
		slog.Error("提权失败", slog.String("username", *username), slog.Any("err", err))
		os.Exit(1)
	}

	slog.Info("提权成功", slog.String("username", *username))
}
