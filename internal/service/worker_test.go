package service

import (
	"SneakerFlash/internal/db"
	redisinfra "SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/testutil"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func newWorkerServiceForTest(t *testing.T) (*WorkerService, *model.User, *model.Product) {
	t.Helper()

	testutil.SetupTestConfig()
	db.DB = testutil.NewSQLiteDB(t)
	testutil.SetupTestRedis(t)

	user := &model.User{Username: "worker-user", Password: "hashed"}
	if err := db.DB.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	product := &model.Product{
		UserID:    user.ID,
		Name:      "Jordan 1",
		Price:     1299,
		Stock:     0,
		StartTime: time.Now().Add(-time.Hour),
	}
	if err := db.DB.Create(product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}

	svc := NewWorkerService(db.DB, repository.NewProductRepo(db.DB), repository.NewOrderRepo(db.DB))
	return svc, user, product
}

func TestWorkerService_BatchCreateOrdersFromMessagesRollsBackRedisOnPartialStockFailure(t *testing.T) {
	svc, user, product := newWorkerServiceForTest(t)
	ctx := context.Background()

	stockKey := fmt.Sprintf("product:stock:%d", product.ID)
	userSetKey := fmt.Sprintf("product:users:%d", product.ID)
	if err := redisinfra.RDB.Set(ctx, stockKey, 0, 0).Err(); err != nil {
		t.Fatalf("set stock key: %v", err)
	}
	if err := redisinfra.RDB.SAdd(ctx, userSetKey, user.ID).Err(); err != nil {
		t.Fatalf("seed user set: %v", err)
	}

	message := SeckillMessage{
		UserID:     user.ID,
		ProductID:  product.ID,
		OrderNum:   "ORD-ROLLBACK-001",
		PaymentID:  "PAY-ROLLBACK-001",
		PriceCents: 129900,
		Time:       time.Now(),
	}
	body, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}

	failed, err := svc.BatchCreateOrdersFromMessages([][]byte{body})
	if err != nil {
		t.Fatalf("BatchCreateOrdersFromMessages() error = %v", err)
	}
	if len(failed) != 1 || failed[0] != 0 {
		t.Fatalf("failed indexes = %v, want [0]", failed)
	}

	stock, err := redisinfra.RDB.Get(ctx, stockKey).Int()
	if err != nil {
		t.Fatalf("get stock key: %v", err)
	}
	if stock != 1 {
		t.Fatalf("stock after rollback = %d, want 1", stock)
	}

	isMember, err := redisinfra.RDB.SIsMember(ctx, userSetKey, user.ID).Result()
	if err != nil {
		t.Fatalf("SIsMember() error = %v", err)
	}
	if isMember {
		t.Fatal("user marker should be removed after rollback")
	}

	cache, err := getPendingOrder(ctx, message.OrderNum)
	if err != nil {
		t.Fatalf("getPendingOrder() error = %v", err)
	}
	if cache.Status != PendingStatusFailed || cache.Message != "库存不足" {
		t.Fatalf("pending cache = %+v, want failed stock insufficient", cache)
	}
}
