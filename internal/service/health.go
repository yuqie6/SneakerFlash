package service

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/db"
	"SneakerFlash/internal/infra/kafka"
	redisinfra "SneakerFlash/internal/infra/redis"
	"context"
	"fmt"
	"strings"
	"time"
)

const healthServiceName = "SneakerFlash"

type HealthService struct{}

type ProbeStatus struct {
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type ReadinessChecks struct {
	Database ProbeStatus `json:"database"`
	Redis    ProbeStatus `json:"redis"`
	Kafka    ProbeStatus `json:"kafka"`
}

type HealthStatus struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
}

type ReadinessStatus struct {
	Status    string          `json:"status"`
	Service   string          `json:"service"`
	Timestamp time.Time       `json:"timestamp"`
	Checks    ReadinessChecks `json:"checks"`
}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) Health(_ context.Context) HealthStatus {
	return HealthStatus{
		Status:    "ok",
		Service:   healthServiceName,
		Timestamp: time.Now().UTC(),
	}
}

func (s *HealthService) Ready(ctx context.Context) (ReadinessStatus, error) {
	status := ReadinessStatus{
		Status:    "ready",
		Service:   healthServiceName,
		Timestamp: time.Now().UTC(),
		Checks: ReadinessChecks{
			Database: ProbeStatus{Status: "up"},
			Redis:    ProbeStatus{Status: "up"},
			Kafka:    ProbeStatus{Status: "up"},
		},
	}

	var notReady []string

	if err := pingDatabase(ctx); err != nil {
		status.Checks.Database = ProbeStatus{Status: "down", Detail: err.Error()}
		notReady = append(notReady, "database")
	}

	if err := pingRedis(ctx); err != nil {
		status.Checks.Redis = ProbeStatus{Status: "down", Detail: err.Error()}
		notReady = append(notReady, "redis")
	}

	if err := kafka.Ping(config.Conf.Data.Kafka.Topic); err != nil {
		status.Checks.Kafka = ProbeStatus{Status: "down", Detail: err.Error()}
		notReady = append(notReady, "kafka")
	}

	if len(notReady) > 0 {
		status.Status = "not_ready"
		return status, fmt.Errorf("components not ready: %s", strings.Join(notReady, ","))
	}

	return status, nil
}

func pingDatabase(parent context.Context) error {
	if db.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}

	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

func pingRedis(parent context.Context) error {
	if redisinfra.RDB == nil {
		return fmt.Errorf("redis not initialized")
	}

	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()

	if err := redisinfra.RDB.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}
