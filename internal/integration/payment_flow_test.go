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
	"gorm.io/gorm"
)

func TestPaymentCallbackFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		callbackStatus     string
		wantOrderStatus    model.OrderStatus
		wantPaymentStatus  model.PaymentStatus
		wantSpentCents     int64
		withUsedCoupon     bool
		wantCouponReleased bool
	}{
		{
			name:              "paid updates order payment and user growth",
			callbackStatus:    "paid",
			wantOrderStatus:   model.OrderStatusPaid,
			wantPaymentStatus: model.PaymentStatusPaid,
			wantSpentCents:    99900,
		},
		{
			name:               "failed releases coupon",
			callbackStatus:     "failed",
			wantOrderStatus:    model.OrderStatusFailed,
			wantPaymentStatus:  model.PaymentStatusFailed,
			wantSpentCents:     0,
			withUsedCoupon:     true,
			wantCouponReleased: true,
		},
		{
			name:               "refunded releases coupon",
			callbackStatus:     "refunded",
			wantOrderStatus:    model.OrderStatusFailed,
			wantPaymentStatus:  model.PaymentStatusRefunded,
			wantSpentCents:     0,
			withUsedCoupon:     true,
			wantCouponReleased: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdb := setupIntegrationDB(t)
			ctx := context.Background()

			user, _, order, payment := seedPaymentFixtures(t, gdb)
			var userCoupon *model.UserCoupon
			if tt.withUsedCoupon {
				userCoupon = seedUsedCouponForOrder(t, gdb, user.ID, order.ID)
			}

			orderSvc := service.NewOrderService(gdb, repository.NewProductRepo(gdb), repository.NewUserRepo(gdb))
			orderHandler := handler.NewOrderHandler(orderSvc)

			router := gin.New()
			router.POST("/api/v1/payment/callback", orderHandler.PaymentCallback)

			reqBody := []byte(`{"payment_id":"PAY-001","status":"` + tt.callbackStatus + `","notify_data":"mock"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/payment/callback", bytes.NewReader(reqBody))
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
			if resp.Data.Order.Status != int(tt.wantOrderStatus) || resp.Data.Payment.Status != string(tt.wantPaymentStatus) {
				t.Fatalf("response data = %+v", resp.Data)
			}

			orderRepo := repository.NewOrderRepo(gdb)
			updatedOrder, err := orderRepo.GetByID(ctx, order.ID)
			if err != nil {
				t.Fatalf("GetByID(order) error = %v", err)
			}
			if updatedOrder.Status != tt.wantOrderStatus {
				t.Fatalf("order status = %v, want %v", updatedOrder.Status, tt.wantOrderStatus)
			}

			paymentRepo := repository.NewPaymentRepo(gdb)
			updatedPayment, err := paymentRepo.GetByPaymentID(ctx, payment.PaymentID)
			if err != nil {
				t.Fatalf("GetByPaymentID() error = %v", err)
			}
			if updatedPayment.Status != tt.wantPaymentStatus {
				t.Fatalf("payment status = %v, want %v", updatedPayment.Status, tt.wantPaymentStatus)
			}

			userRepo := repository.NewUserRepo(gdb)
			updatedUser, err := userRepo.GetByID(ctx, user.ID)
			if err != nil {
				t.Fatalf("GetByID(user) error = %v", err)
			}
			if updatedUser.TotalSpentCents != tt.wantSpentCents {
				t.Fatalf("user total spent = %d, want %d", updatedUser.TotalSpentCents, tt.wantSpentCents)
			}

			if tt.wantCouponReleased {
				userCouponRepo := repository.NewUserCouponRepo(gdb)
				released, err := userCouponRepo.GetByIDForUpdate(ctx, userCoupon.ID)
				if err != nil {
					t.Fatalf("GetByIDForUpdate(user coupon) error = %v", err)
				}
				if released.Status != model.CouponStatusAvailable || released.OrderID != nil {
					t.Fatalf("released coupon = %+v", released)
				}
			}
		})
	}
}

func seedPaymentFixtures(t *testing.T, gdb *gorm.DB) (*model.User, *model.Product, *model.Order, *model.Payment) {
	t.Helper()

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

	return user, product, order, payment
}

func seedUsedCouponForOrder(t *testing.T, gdb *gorm.DB, userID uint, orderID uint) *model.UserCoupon {
	t.Helper()

	coupon := &model.Coupon{
		Type:          model.CouponTypeFullCut,
		Title:         "支付失败释放券",
		AmountCents:   500,
		MinSpendCents: 1000,
		ValidFrom:     time.Now().Add(-time.Hour),
		ValidTo:       time.Now().Add(time.Hour),
		Status:        model.CouponTemplateStatusActive,
	}
	if err := gdb.Create(coupon).Error; err != nil {
		t.Fatalf("create coupon: %v", err)
	}

	userCoupon := &model.UserCoupon{
		UserID:       userID,
		CouponID:     coupon.ID,
		Status:       model.CouponStatusUsed,
		ObtainedFrom: "purchase",
		ValidFrom:    time.Now().Add(-time.Hour),
		ValidTo:      time.Now().Add(time.Hour),
		OrderID:      &orderID,
		IssuedAt:     time.Now(),
	}
	if err := gdb.Create(userCoupon).Error; err != nil {
		t.Fatalf("create user coupon: %v", err)
	}

	return userCoupon
}
