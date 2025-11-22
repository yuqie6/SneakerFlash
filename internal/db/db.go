package db

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/model"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	var gormLogger logger.Interface
	if cfg.LogLever == 4 {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// 打开连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	log.Println("数据库连接成功")

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取底层db失败: %v", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
	log.Println("连接池设置成功")

	DB = db
}

func MakeMigrate() {
	if DB == nil {
		log.Fatal("数据库没有初始化")
	}

	log.Println("正在迁移数据库")

	err := DB.AutoMigrate(
		&model.Order{},
		&model.User{},
		&model.Product{},
		&model.Payment{},
	)

	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	log.Println("数据库迁移成功")
}
