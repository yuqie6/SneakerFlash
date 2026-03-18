package service

import (
	"SneakerFlash/internal/db"
	"SneakerFlash/internal/testutil"
	"context"
	"fmt"
	"testing"
)

func TestHealthService_Health(t *testing.T) {
	testutil.SetupTestConfig()
	svc := NewHealthService()

	got := svc.Health(context.Background())
	if got.Status != "ok" {
		t.Fatalf("Health().Status = %q, want ok", got.Status)
	}
	if got.Service != "SneakerFlash" {
		t.Fatalf("Health().Service = %q, want SneakerFlash", got.Service)
	}
	if got.Timestamp.IsZero() {
		t.Fatal("Health().Timestamp is zero")
	}
}

func TestHealthService_Ready(t *testing.T) {
	testutil.SetupTestConfig()
	db.DB = testutil.NewSQLiteDB(t)
	testutil.SetupTestRedis(t)

	originalKafkaPing := kafkaPing
	t.Cleanup(func() {
		kafkaPing = originalKafkaPing
	})

	t.Run("all dependencies ready", func(t *testing.T) {
		kafkaPing = func(topic string) error { return nil }

		status, err := NewHealthService().Ready(context.Background())
		if err != nil {
			t.Fatalf("Ready() error = %v", err)
		}
		if status.Status != "ready" {
			t.Fatalf("Ready().Status = %q, want ready", status.Status)
		}
		if status.Checks.Database.Status != "up" || status.Checks.Redis.Status != "up" || status.Checks.Kafka.Status != "up" {
			t.Fatalf("unexpected readiness checks: %+v", status.Checks)
		}
	})

	t.Run("kafka not ready", func(t *testing.T) {
		kafkaPing = func(topic string) error { return fmt.Errorf("metadata unavailable") }

		status, err := NewHealthService().Ready(context.Background())
		if err == nil {
			t.Fatal("Ready() error = nil, want not ready error")
		}
		if status.Status != "not_ready" {
			t.Fatalf("Ready().Status = %q, want not_ready", status.Status)
		}
		if status.Checks.Kafka.Status != "down" {
			t.Fatalf("Kafka check status = %q, want down", status.Checks.Kafka.Status)
		}
	})
}
