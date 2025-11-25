package db

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/logger"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// 全局 Database 变量, 只和repository层交互
var DB *gorm.DB

// 初始化数据库连接
func Init(cfg config.DatabaseConfig) {
	// 动态构建 dsn
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBname)

	// 配置日志模式
	logLevel := gormlogger.Error
	switch cfg.LogLever {
	case 1:
		logLevel = gormlogger.Silent
	case 2:
		logLevel = gormlogger.Error
	case 3:
		logLevel = gormlogger.Warn
	case 4:
		logLevel = gormlogger.Info
	default:
		logLevel = gormlogger.Error
	}
	slowThreshold := time.Duration(cfg.SlowThresholdMs) * time.Millisecond
	gormLogger := logger.NewGormLogger(logLevel, slowThreshold)

	// 打开连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		slog.Error("数据库连接失败", slog.Any("err", err))
		panic(err)
	}
	slog.Info("数据库连接成功")

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("获取底层db失败", slog.Any("err", err))
		panic(err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
	slog.Info("连接池设置成功",
		slog.Int("max_idle", cfg.MaxIdle),
		slog.Int("max_open", cfg.MaxOpen),
		slog.Int("max_lifetime", cfg.MaxLifetime),
	)

	DB = db
}

func MakeMigrate() {
	if DB == nil {
		slog.Error("数据库没有初始化")
		panic("db not initialized")
	}

	slog.Info("正在迁移数据库")

	err := DB.AutoMigrate(
		&model.Order{},
		&model.User{},
		&model.Product{},
		&model.Payment{},
	)

	if err != nil {
		slog.Error("数据库迁移失败", slog.Any("err", err))
		panic(err)
	}

	slog.Info("数据库迁移成功")
}
