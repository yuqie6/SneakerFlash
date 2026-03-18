package service

import (
	"SneakerFlash/internal/db"
	redisinfra "SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/testutil"
	"context"
	"testing"
	"time"

	"gorm.io/gorm"
)

func newOrderServiceForTest(t *testing.T) (*OrderService, *gormFixtures) {
	t.Helper()

	testutil.SetupTestConfig()
	db.DB = testutil.NewSQLiteDB(t)
	testutil.SetupTestRedis(t)

	fixtures := seedOrderFixtures(t, db.DB)
	svc := NewOrderService(db.DB, repository.NewProductRepo(db.DB), repository.NewUserRepo(db.DB))
	return svc, fixtures
}

type gormFixtures struct {
	user    *model.User
	product *model.Product
	order   *model.Order
	payment *model.Payment
}

func seedOrderFixtures(t *testing.T, gdb *gorm.DB) *gormFixtures {
	t.Helper()

	now := time.Now()
	user := &model.User{Username: "alice", Password: "hashed"}
	product := &model.Product{
		UserID:    1,
		Name:      "AJ 1",
		Price:     1299,
		Stock:     10,
		StartTime: now.Add(-time.Hour),
	}
	order := &model.Order{
		UserID:    1,
		ProductID: 1,
		OrderNum:  "ORD-001",
		Status:    model.OrderStatusUnpaid,
	}
	payment := &model.Payment{
		OrderID:     1,
		PaymentID:   "PAY-001",
		AmountCents: 129900,
		Status:      model.PaymentStatusPending,
	}

	if err := gdb.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	product.UserID = user.ID
	if err := gdb.Create(product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	order.UserID = user.ID
	order.ProductID = product.ID
	if err := gdb.Create(order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	payment.OrderID = order.ID
	if err := gdb.Create(payment).Error; err != nil {
		t.Fatalf("create payment: %v", err)
	}

	return &gormFixtures{
		user:    user,
		product: product,
		order:   order,
		payment: payment,
	}
}

func TestOrderService_PollOrder(t *testing.T) {
	svc, fixtures := newOrderServiceForTest(t)
	ctx := context.Background()

	t.Run("pending from cache", func(t *testing.T) {
		err := setPendingOrder(ctx, PendingOrderCache{
			OrderNum:  "PENDING-001",
			PaymentID: "PAY-PENDING",
			Status:    PendingStatusPending,
		})
		if err != nil {
			t.Fatalf("setPendingOrder() error = %v", err)
		}

		got, err := svc.PollOrder(ctx, fixtures.user.ID, "PENDING-001")
		if err != nil {
			t.Fatalf("PollOrder() error = %v", err)
		}
		if got.Status != PendingStatusPending || got.PaymentID != "PAY-PENDING" {
			t.Fatalf("PollOrder() = %+v, want pending with payment id", got)
		}
	})

	t.Run("ready from cache", func(t *testing.T) {
		err := setPendingOrder(ctx, PendingOrderCache{
			OrderNum:  fixtures.order.OrderNum,
			OrderID:   fixtures.order.ID,
			PaymentID: fixtures.payment.PaymentID,
			Status:    PendingStatusReady,
		})
		if err != nil {
			t.Fatalf("setPendingOrder() error = %v", err)
		}

		got, err := svc.PollOrder(ctx, fixtures.user.ID, fixtures.order.OrderNum)
		if err != nil {
			t.Fatalf("PollOrder() error = %v", err)
		}
		if got.Status != PendingStatusReady || got.Order == nil || got.Order.Order.ID != fixtures.order.ID {
			t.Fatalf("PollOrder() = %+v, want ready with order", got)
		}
	})

	t.Run("fallback to database", func(t *testing.T) {
		if err := redisinfra.RDB.Del(ctx, pendingOrderKey(fixtures.order.OrderNum)).Err(); err != nil {
			t.Fatalf("delete pending cache: %v", err)
		}

		got, err := svc.PollOrder(ctx, fixtures.user.ID, fixtures.order.OrderNum)
		if err != nil {
			t.Fatalf("PollOrder() error = %v", err)
		}
		if got.Status != PendingStatusReady || got.Order == nil {
			t.Fatalf("PollOrder() = %+v, want ready from database", got)
		}
	})
}

func TestOrderService_HandlePaymentResultPaidIsIdempotent(t *testing.T) {
	svc, fixtures := newOrderServiceForTest(t)
	ctx := context.Background()

	got, err := svc.HandlePaymentResult(ctx, fixtures.payment.PaymentID, model.PaymentStatusPaid, "mock")
	if err != nil {
		t.Fatalf("HandlePaymentResult() error = %v", err)
	}
	if got.Order.Status != model.OrderStatusPaid {
		t.Fatalf("order status = %v, want %v", got.Order.Status, model.OrderStatusPaid)
	}
	if got.Payment.Status != model.PaymentStatusPaid {
		t.Fatalf("payment status = %v, want %v", got.Payment.Status, model.PaymentStatusPaid)
	}

	gotAgain, err := svc.HandlePaymentResult(ctx, fixtures.payment.PaymentID, model.PaymentStatusPaid, "mock")
	if err != nil {
		t.Fatalf("HandlePaymentResult() second error = %v", err)
	}
	if gotAgain.Order.Status != model.OrderStatusPaid || gotAgain.Payment.Status != model.PaymentStatusPaid {
		t.Fatalf("second HandlePaymentResult() = %+v, want paid state preserved", gotAgain)
	}
}

func TestOrderService_CancelExpiredOrders(t *testing.T) {
	svc, fixtures := newOrderServiceForTest(t)
	ctx := context.Background()

	if err := db.DB.Model(&model.Order{}).Where("id = ?", fixtures.order.ID).Update("created_at", time.Now().Add(-20*time.Minute)).Error; err != nil {
		t.Fatalf("update order created_at: %v", err)
	}
	if err := setStockCache(ctx, fixtures.product.ID, fixtures.product.Stock); err != nil {
		t.Fatalf("set stock cache: %v", err)
	}
	if err := redisinfra.RDB.SAdd(ctx, "product:users:1", fixtures.user.ID).Err(); err != nil {
		t.Fatalf("seed user marker: %v", err)
	}

	cancelled, err := svc.CancelExpiredOrders(ctx, 15*time.Minute, 10)
	if err != nil {
		t.Fatalf("CancelExpiredOrders() error = %v", err)
	}
	if cancelled != 1 {
		t.Fatalf("cancelled = %d, want 1", cancelled)
	}

	order, err := repository.NewOrderRepo(db.DB).GetByID(ctx, fixtures.order.ID)
	if err != nil {
		t.Fatalf("load order: %v", err)
	}
	if order.Status != model.OrderStatusCancelled {
		t.Fatalf("order status = %v, want %v", order.Status, model.OrderStatusCancelled)
	}

	payment, err := repository.NewPaymentRepo(db.DB).GetByOrderID(ctx, fixtures.order.ID)
	if err != nil {
		t.Fatalf("load payment: %v", err)
	}
	if payment.Status != model.PaymentStatusFailed {
		t.Fatalf("payment status = %v, want %v", payment.Status, model.PaymentStatusFailed)
	}

	product, err := repository.NewProductRepo(db.DB).GetByID(ctx, fixtures.product.ID)
	if err != nil {
		t.Fatalf("load product: %v", err)
	}
	if product.Stock != fixtures.product.Stock+1 {
		t.Fatalf("product stock = %d, want %d", product.Stock, fixtures.product.Stock+1)
	}
}

func TestOrderService_CancelExpiredOrdersSkipsPaidPayment(t *testing.T) {
	svc, fixtures := newOrderServiceForTest(t)
	ctx := context.Background()

	if err := db.DB.Model(&model.Order{}).Where("id = ?", fixtures.order.ID).Update("created_at", time.Now().Add(-20*time.Minute)).Error; err != nil {
		t.Fatalf("update order created_at: %v", err)
	}
	if err := db.DB.Model(&model.Payment{}).Where("id = ?", fixtures.payment.ID).Update("status", model.PaymentStatusPaid).Error; err != nil {
		t.Fatalf("update payment status: %v", err)
	}
	if err := setStockCache(ctx, fixtures.product.ID, fixtures.product.Stock); err != nil {
		t.Fatalf("set stock cache: %v", err)
	}

	cancelled, err := svc.CancelExpiredOrders(ctx, 15*time.Minute, 10)
	if err != nil {
		t.Fatalf("CancelExpiredOrders() error = %v", err)
	}
	if cancelled != 0 {
		t.Fatalf("cancelled = %d, want 0", cancelled)
	}

	order, err := repository.NewOrderRepo(db.DB).GetByID(ctx, fixtures.order.ID)
	if err != nil {
		t.Fatalf("load order: %v", err)
	}
	if order.Status != model.OrderStatusUnpaid {
		t.Fatalf("order status = %v, want %v", order.Status, model.OrderStatusUnpaid)
	}

	product, err := repository.NewProductRepo(db.DB).GetByID(ctx, fixtures.product.ID)
	if err != nil {
		t.Fatalf("load product: %v", err)
	}
	if product.Stock != fixtures.product.Stock {
		t.Fatalf("product stock = %d, want %d", product.Stock, fixtures.product.Stock)
	}
}
