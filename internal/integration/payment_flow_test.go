//go:build integration

package integration

import (
	"SneakerFlash/internal/handler"
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestPaymentCallbackFlow_UpdatesOrderAndPayment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := setupIntegrationDB(t)
	ctx := context.Background()

	user := &model.User{Username: "alice", Password: "hashed", GrowthLevel: 1}
	if err := gdb.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	product := &model.Product{
		UserID:    user.ID,
		Name:      "Payment Drop",
		Price:     999,
		Stock:     10,
		StartTime: time.Now().Add(-time.Hour),
	}
	if err := gdb.Create(product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}

	order := &model.Order{
		UserID:    user.ID,
		ProductID: product.ID,
		OrderNum:  "ORD-PAY-001",
		Status:    model.OrderStatusUnpaid,
	}
	if err := gdb.Create(order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	payment := &model.Payment{
		OrderID:     order.ID,
		PaymentID:   "PAY-001",
		AmountCents: 99900,
		Status:      model.PaymentStatusPending,
	}
	if err := gdb.Create(payment).Error; err != nil {
		t.Fatalf("create payment: %v", err)
	}

	orderSvc := service.NewOrderService(gdb, repository.NewProductRepo(gdb), repository.NewUserRepo(gdb))
	orderHandler := handler.NewOrderHandler(orderSvc)

	router := gin.New()
	router.POST("/api/v1/payment/callback", orderHandler.PaymentCallback)

	body := []byte(`{"payment_id":"PAY-001","status":"paid","notify_data":"mock"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payment/callback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Order struct {
				ID     uint `json:"id"`
				Status int  `json:"status"`
			} `json:"order"`
			Payment struct {
				PaymentID string `json:"payment_id"`
				Status    string `json:"status"`
			} `json:"payment"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.Order.Status != int(model.OrderStatusPaid) || resp.Data.Payment.Status != string(model.PaymentStatusPaid) {
		t.Fatalf("response data = %+v", resp.Data)
	}

	orderRepo := repository.NewOrderRepo(gdb)
	updatedOrder, err := orderRepo.GetByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetByID(order) error = %v", err)
	}
	if updatedOrder.Status != model.OrderStatusPaid {
		t.Fatalf("order status = %v, want %v", updatedOrder.Status, model.OrderStatusPaid)
	}

	paymentRepo := repository.NewPaymentRepo(gdb)
	updatedPayment, err := paymentRepo.GetByPaymentID(ctx, payment.PaymentID)
	if err != nil {
		t.Fatalf("GetByPaymentID() error = %v", err)
	}
	if updatedPayment.Status != model.PaymentStatusPaid {
		t.Fatalf("payment status = %v, want %v", updatedPayment.Status, model.PaymentStatusPaid)
	}

	userRepo := repository.NewUserRepo(gdb)
	updatedUser, err := userRepo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID(user) error = %v", err)
	}
	if updatedUser.TotalSpentCents != payment.AmountCents {
		t.Fatalf("user total spent = %d, want %d", updatedUser.TotalSpentCents, payment.AmountCents)
	}
}
