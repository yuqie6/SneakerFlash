//go:build integration

package integration

import (
	redisinfra "SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestWorkerFlow_CreateOrderAndPendingReady(t *testing.T) {
	gdb := setupIntegrationDB(t)
	ctx := context.Background()

	product := &model.Product{
		UserID:    1,
		Name:      "Worker Drop",
		Price:     999,
		Stock:     5,
		StartTime: time.Now().Add(-time.Hour),
	}
	if err := gdb.Create(product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}

	workerSvc := service.NewWorkerService(gdb, repository.NewProductRepo(gdb), repository.NewOrderRepo(gdb))

	msg := service.SeckillMessage{
		UserID:     7,
		ProductID:  product.ID,
		OrderNum:   "ORD-INTEGRATION-1",
		PaymentID:  "PAY-INTEGRATION-1",
		PriceCents: 99900,
		Time:       time.Now(),
	}
	body, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal msg: %v", err)
	}

	stockKey := fmt.Sprintf("product:stock:%d", product.ID)
	userSetKey := fmt.Sprintf("product:users:%d", product.ID)
	if err := redisinfra.RDB.Set(ctx, stockKey, product.Stock, 0).Err(); err != nil {
		t.Fatalf("set stock cache error = %v", err)
	}
	if err := redisinfra.RDB.SAdd(ctx, userSetKey, msg.UserID).Err(); err != nil {
		t.Fatalf("set user set error = %v", err)
	}

	failed, err := workerSvc.BatchCreateOrdersFromMessages([][]byte{body})
	if err != nil {
		t.Fatalf("BatchCreateOrdersFromMessages() error = %v", err)
	}
	if len(failed) != 0 {
		t.Fatalf("failed indexes = %v, want none", failed)
	}

	orderRepo := repository.NewOrderRepo(gdb)
	order, err := orderRepo.GetByOrderNum(ctx, msg.OrderNum)
	if err != nil {
		t.Fatalf("GetByOrderNum() error = %v", err)
	}
	if order.Status != model.OrderStatusUnpaid {
		t.Fatalf("order status = %v, want unpaid", order.Status)
	}

	paymentRepo := repository.NewPaymentRepo(gdb)
	payment, err := paymentRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetByOrderID() error = %v", err)
	}
	if payment.PaymentID != msg.PaymentID {
		t.Fatalf("payment id = %q, want %q", payment.PaymentID, msg.PaymentID)
	}

	cache, err := redisinfra.RDB.Get(ctx, fmt.Sprintf("order:pending:%s", msg.OrderNum)).Result()
	if err != nil {
		t.Fatalf("get pending cache error = %v", err)
	}
	var pending service.PendingOrderCache
	if err := json.Unmarshal([]byte(cache), &pending); err != nil {
		t.Fatalf("decode pending cache error = %v", err)
	}
	if pending.Status != service.PendingStatusReady || pending.OrderID != order.ID {
		t.Fatalf("pending cache = %+v", pending)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		remaining, err := redisinfra.RDB.Get(ctx, stockKey).Int()
		if err != nil {
			t.Fatalf("get stock cache error = %v", err)
		}
		if remaining == product.Stock-1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("stock cache = %d, want %d", remaining, product.Stock-1)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
