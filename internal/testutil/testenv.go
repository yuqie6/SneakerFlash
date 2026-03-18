package testutil

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/utils"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupTestConfig() {
	config.Conf = config.Config{
		Server: config.ServerConfig{
			Port:      ":8000",
			MachineID: 1,
			UploadDir: "uploads",
		},
		Data: config.DataConfig{
			Kafka: config.KafkaConfig{
				Topic: "seckill-order-test",
			},
		},
		JWT: config.JWTConfig{
			Secret:         "test-secret",
			Expried:        3600,
			RefreshExpried: 7200,
		},
	}
	_ = utils.InitSnowflake(1)
}

func NewSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	err = db.AutoMigrate(
		&model.Order{},
		&model.User{},
		&model.Product{},
		&model.Payment{},
		&model.Coupon{},
		&model.UserCoupon{},
		&model.PaidVIP{},
		&model.OutboxMessage{},
	)
	if err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	return db
}

func SetupTestRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}

	client := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
		DB:   0,
	})

	redis.RDB = client

	t.Cleanup(func() {
		_ = client.Close()
		mr.Close()
	})

	return mr
}

func Ptr[T any](v T) *T {
	return &v
}

func MustTime(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}

	return parsed
}
