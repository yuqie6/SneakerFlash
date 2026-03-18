package service

import (
	"SneakerFlash/internal/db"
	redisinfra "SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/testutil"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func newSeckillServiceForTest(t *testing.T) (*SeckillService, *model.Product) {
	t.Helper()

	testutil.SetupTestConfig()
	db.DB = testutil.NewSQLiteDB(t)
	testutil.SetupTestRedis(t)

	product := &model.Product{
		UserID:    1,
		Name:      "Jordan Test",
		Price:     999,
		Stock:     5,
		StartTime: time.Now().Add(-time.Hour),
	}
	if err := db.DB.Create(product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}

	svc := NewSeckillService(db.DB, repository.NewProductRepo(db.DB))
	return svc, product
}

func TestSeckillService_Seckill(t *testing.T) {
	svc, product := newSeckillServiceForTest(t)
	ctx := context.Background()

	originalSend := sendKafkaMessage
	originalGen := genSeckillID
	t.Cleanup(func() {
		sendKafkaMessage = originalSend
		genSeckillID = originalGen
	})

	idSequence := []string{"ORD-100", "PAY-100", "ORD-200", "PAY-200"}
	genSeckillID = func() (string, error) {
		if len(idSequence) == 0 {
			return "", fmt.Errorf("no more ids")
		}
		next := idSequence[0]
		idSequence = idSequence[1:]
		return next, nil
	}
	sendKafkaMessage = func(topic, message string) error { return errors.New("skip send in unit test") }

	t.Run("not started", func(t *testing.T) {
		futureProduct := &model.Product{
			UserID:    2,
			Name:      "Future Drop",
			Price:     1299,
			Stock:     3,
			StartTime: time.Now().Add(time.Hour),
		}
		if err := db.DB.Create(futureProduct).Error; err != nil {
			t.Fatalf("create future product: %v", err)
		}

		if _, err := svc.Seckill(ctx, 1, futureProduct.ID); !errors.Is(err, ErrSeckillNotStart) {
			t.Fatalf("Seckill(not started) error = %v, want %v", err, ErrSeckillNotStart)
		}
	})

	t.Run("sold out", func(t *testing.T) {
		if err := redisinfra.RDB.Set(ctx, fmt.Sprintf("product:stock:%d", product.ID), 0, 0).Err(); err != nil {
			t.Fatalf("set stock cache: %v", err)
		}

		if _, err := svc.Seckill(ctx, 2, product.ID); !errors.Is(err, ErrSeckillFull) {
			t.Fatalf("Seckill(sold out) error = %v, want %v", err, ErrSeckillFull)
		}
	})

	t.Run("repeat purchase", func(t *testing.T) {
		stockKey := fmt.Sprintf("product:stock:%d", product.ID)
		userSetKey := fmt.Sprintf("product:users:%d", product.ID)
		if err := redisinfra.RDB.Set(ctx, stockKey, 5, 0).Err(); err != nil {
			t.Fatalf("set stock cache: %v", err)
		}
		if err := redisinfra.RDB.SAdd(ctx, userSetKey, 3).Err(); err != nil {
			t.Fatalf("set user cache: %v", err)
		}

		if _, err := svc.Seckill(ctx, 3, product.ID); !errors.Is(err, ErrSeckillRepeat) {
			t.Fatalf("Seckill(repeat) error = %v, want %v", err, ErrSeckillRepeat)
		}
	})

	t.Run("success writes pending cache", func(t *testing.T) {
		stockKey := fmt.Sprintf("product:stock:%d", product.ID)
		userSetKey := fmt.Sprintf("product:users:%d", product.ID)
		if err := redisinfra.RDB.Set(ctx, stockKey, 5, 0).Err(); err != nil {
			t.Fatalf("set stock cache: %v", err)
		}
		if err := redisinfra.RDB.Del(ctx, userSetKey).Err(); err != nil {
			t.Fatalf("clear user cache: %v", err)
		}

		got, err := svc.Seckill(ctx, 9, product.ID)
		if err != nil {
			t.Fatalf("Seckill() error = %v", err)
		}
		if got.Status != string(PendingStatusPending) || got.OrderNum == "" || got.PaymentID == "" {
			t.Fatalf("Seckill() = %+v, want pending result", got)
		}

		cache, err := getPendingOrder(ctx, got.OrderNum)
		if err != nil {
			t.Fatalf("getPendingOrder() error = %v", err)
		}
		if cache.Status != PendingStatusPending || cache.UserID != 9 || cache.ProductID != product.ID {
			t.Fatalf("pending cache = %+v", cache)
		}

		remaining, err := redisinfra.RDB.Get(ctx, stockKey).Int()
		if err != nil {
			t.Fatalf("get stock cache: %v", err)
		}
		if remaining != 4 {
			t.Fatalf("remaining stock = %d, want 4", remaining)
		}
	})
}
