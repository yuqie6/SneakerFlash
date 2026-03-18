//go:build integration

package integration

import (
	"SneakerFlash/internal/handler"
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/middlerware"
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestAdminFlow_StatsAndResources(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := setupIntegrationDB(t)
	adminUser, user := seedAdminFixtures(t, gdb)
	token := mustAdminToken(t, adminUser)

	adminSvc := service.NewAdminService(gdb, repository.NewUserRepo(gdb), repository.NewProductRepo(gdb))
	riskSvc := service.NewRiskService(nil)
	couponSvc := service.NewCouponService(gdb)
	adminHandler := handler.NewAdminHandler(adminSvc, riskSvc, couponSvc)

	router := gin.New()
	admin := router.Group("/api/v1/admin")
	admin.Use(middlerware.JWTauth(), middlerware.AdminAuth())
	{
		admin.GET("/stats", adminHandler.Stats)
		admin.GET("/users", adminHandler.ListUsers)
		admin.GET("/orders", adminHandler.ListOrders)
		admin.GET("/products", adminHandler.ListProducts)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("stats status = %d body=%s", rec.Code, rec.Body.String())
	}

	var statsResp struct {
		Code int                `json:"code"`
		Data service.AdminStats `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &statsResp); err != nil {
		t.Fatalf("decode stats: %v", err)
	}
	if statsResp.Data.TotalUsers != 2 || statsResp.Data.TotalOrders != 2 || statsResp.Data.TotalRevenueCents != 129900 || statsResp.Data.TotalProducts != 2 || statsResp.Data.PendingOrders != 1 {
		t.Fatalf("unexpected stats: %+v", statsResp.Data)
	}

	usersReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?page=1&page_size=20", nil)
	usersReq.Header.Set("Authorization", "Bearer "+token)
	usersRec := httptest.NewRecorder()
	router.ServeHTTP(usersRec, usersReq)
	if usersRec.Code != http.StatusOK {
		t.Fatalf("users status = %d body=%s", usersRec.Code, usersRec.Body.String())
	}

	var usersResp struct {
		Code int `json:"code"`
		Data struct {
			List []model.User `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(usersRec.Body.Bytes(), &usersResp); err != nil {
		t.Fatalf("decode users: %v", err)
	}
	if len(usersResp.Data.List) != 2 {
		t.Fatalf("users len = %d, want 2", len(usersResp.Data.List))
	}
	foundRole := false
	for _, item := range usersResp.Data.List {
		if item.ID == user.ID && item.Role == model.UserRoleUser {
			foundRole = true
		}
	}
	if !foundRole {
		t.Fatalf("user role not returned: %+v", usersResp.Data.List)
	}

	ordersReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/orders?status=1", nil)
	ordersReq.Header.Set("Authorization", "Bearer "+token)
	ordersRec := httptest.NewRecorder()
	router.ServeHTTP(ordersRec, ordersReq)
	if ordersRec.Code != http.StatusOK {
		t.Fatalf("orders status = %d body=%s", ordersRec.Code, ordersRec.Body.String())
	}

	var ordersResp struct {
		Data struct {
			List []model.Order `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(ordersRec.Body.Bytes(), &ordersResp); err != nil {
		t.Fatalf("decode orders: %v", err)
	}
	if len(ordersResp.Data.List) != 1 || ordersResp.Data.List[0].Status != model.OrderStatusPaid {
		t.Fatalf("unexpected orders: %+v", ordersResp.Data.List)
	}

	productsReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/products", nil)
	productsReq.Header.Set("Authorization", "Bearer "+token)
	productsRec := httptest.NewRecorder()
	router.ServeHTTP(productsRec, productsReq)
	if productsRec.Code != http.StatusOK {
		t.Fatalf("products status = %d body=%s", productsRec.Code, productsRec.Body.String())
	}

	var productsResp struct {
		Data struct {
			List []model.Product `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(productsRec.Body.Bytes(), &productsResp); err != nil {
		t.Fatalf("decode products: %v", err)
	}
	if len(productsResp.Data.List) != 2 {
		t.Fatalf("products len = %d, want 2", len(productsResp.Data.List))
	}

	userToken := mustUserToken(t, user)
	forbiddenReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/stats", nil)
	forbiddenReq.Header.Set("Authorization", "Bearer "+userToken)
	forbiddenRec := httptest.NewRecorder()
	router.ServeHTTP(forbiddenRec, forbiddenReq)
	if forbiddenRec.Code != http.StatusForbidden {
		t.Fatalf("forbidden status = %d, want 403 body=%s", forbiddenRec.Code, forbiddenRec.Body.String())
	}
}

func TestAdminFlow_CouponsAndRisk(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := setupIntegrationDB(t)
	adminUser, _ := seedAdminFixtures(t, gdb)
	token := mustAdminToken(t, adminUser)

	adminSvc := service.NewAdminService(gdb, repository.NewUserRepo(gdb), repository.NewProductRepo(gdb))
	riskSvc := service.NewRiskService(redis.RDB)
	couponSvc := service.NewCouponService(gdb)
	adminHandler := handler.NewAdminHandler(adminSvc, riskSvc, couponSvc)

	router := gin.New()
	admin := router.Group("/api/v1/admin")
	admin.Use(middlerware.JWTauth(), middlerware.AdminAuth())
	{
		admin.GET("/coupons", adminHandler.ListCoupons)
		admin.POST("/coupons", adminHandler.CreateCoupon)
		admin.PUT("/coupons/:id", adminHandler.UpdateCoupon)
		admin.DELETE("/coupons/:id", adminHandler.DeleteCoupon)
		admin.GET("/risk/blacklist", adminHandler.ListBlacklist)
		admin.POST("/risk/blacklist", adminHandler.AddBlacklist)
		admin.DELETE("/risk/blacklist", adminHandler.RemoveBlacklist)
		admin.GET("/risk/graylist", adminHandler.ListGraylist)
		admin.POST("/risk/graylist", adminHandler.AddGraylist)
		admin.DELETE("/risk/graylist", adminHandler.RemoveGraylist)
	}

	createBody := []byte(`{"type":"full_cut","title":"Admin 券","description":"后台创建","amount_cents":800,"discount_rate":0,"min_spend_cents":3000,"valid_from":"2026-03-18T10:00","valid_to":"2026-03-31T10:00","purchasable":true,"price_cents":100,"status":"active"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/coupons", bytes.NewReader(createBody))
	createReq.Header.Set("Authorization", "Bearer "+token)
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create coupon status = %d body=%s", createRec.Code, createRec.Body.String())
	}

	var createResp struct {
		Data model.Coupon `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create coupon: %v", err)
	}
	if createResp.Data.ID == 0 || createResp.Data.Title != "Admin 券" {
		t.Fatalf("unexpected created coupon: %+v", createResp.Data)
	}

	updateBody := []byte(`{"title":"Admin 券更新","type":"discount","discount_rate":85,"amount_cents":0,"min_spend_cents":0,"valid_from":"2026-03-19T10:00","valid_to":"2026-03-30T10:00","purchasable":false,"price_cents":0,"status":"inactive"}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/admin/coupons/"+strconv.Itoa(int(createResp.Data.ID)), bytes.NewReader(updateBody))
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateReq.Header.Set("Content-Type", "application/json")
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update coupon status = %d body=%s", updateRec.Code, updateRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/coupons", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list coupons status = %d body=%s", listRec.Code, listRec.Body.String())
	}

	var listResp struct {
		Data struct {
			List []model.Coupon `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list coupons: %v", err)
	}
	if len(listResp.Data.List) != 1 || listResp.Data.List[0].Type != model.CouponTypeDiscount || listResp.Data.List[0].Status != model.CouponTemplateStatusInactive {
		t.Fatalf("unexpected coupons: %+v", listResp.Data.List)
	}

	addBlackReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/risk/blacklist", bytes.NewReader([]byte(`{"type":"ip","value":"127.0.0.1"}`)))
	addBlackReq.Header.Set("Authorization", "Bearer "+token)
	addBlackReq.Header.Set("Content-Type", "application/json")
	addBlackRec := httptest.NewRecorder()
	router.ServeHTTP(addBlackRec, addBlackReq)
	if addBlackRec.Code != http.StatusOK {
		t.Fatalf("add blacklist status = %d body=%s", addBlackRec.Code, addBlackRec.Body.String())
	}

	addGrayReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/risk/graylist", bytes.NewReader([]byte(`{"type":"user","value":"42"}`)))
	addGrayReq.Header.Set("Authorization", "Bearer "+token)
	addGrayReq.Header.Set("Content-Type", "application/json")
	addGrayRec := httptest.NewRecorder()
	router.ServeHTTP(addGrayRec, addGrayReq)
	if addGrayRec.Code != http.StatusOK {
		t.Fatalf("add graylist status = %d body=%s", addGrayRec.Code, addGrayRec.Body.String())
	}

	blackReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/risk/blacklist", nil)
	blackReq.Header.Set("Authorization", "Bearer "+token)
	blackRec := httptest.NewRecorder()
	router.ServeHTTP(blackRec, blackReq)
	if blackRec.Code != http.StatusOK {
		t.Fatalf("list blacklist status = %d body=%s", blackRec.Code, blackRec.Body.String())
	}

	var blackResp struct {
		Data struct {
			IP []string `json:"ip"`
		} `json:"data"`
	}
	if err := json.Unmarshal(blackRec.Body.Bytes(), &blackResp); err != nil {
		t.Fatalf("decode blacklist: %v", err)
	}
	if len(blackResp.Data.IP) != 1 || blackResp.Data.IP[0] != "127.0.0.1" {
		t.Fatalf("unexpected blacklist: %+v", blackResp.Data.IP)
	}

	removeBlackReq := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/risk/blacklist", bytes.NewReader([]byte(`{"type":"ip","value":"127.0.0.1"}`)))
	removeBlackReq.Header.Set("Authorization", "Bearer "+token)
	removeBlackReq.Header.Set("Content-Type", "application/json")
	removeBlackRec := httptest.NewRecorder()
	router.ServeHTTP(removeBlackRec, removeBlackReq)
	if removeBlackRec.Code != http.StatusOK {
		t.Fatalf("remove blacklist status = %d body=%s", removeBlackRec.Code, removeBlackRec.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/coupons/"+strconv.Itoa(int(createResp.Data.ID)), nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusOK {
		t.Fatalf("delete coupon status = %d body=%s", deleteRec.Code, deleteRec.Body.String())
	}

	ctx := context.Background()
	if _, err := repository.NewCouponRepo(gdb).GetByID(ctx, createResp.Data.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("coupon still exists, err=%v", err)
	}
}

func seedAdminFixtures(t *testing.T, gdb *gorm.DB) (*model.User, *model.User) {
	t.Helper()

	adminUser := &model.User{Username: "admin", Password: "hashed", Role: model.UserRoleAdmin, GrowthLevel: 4}
	if err := gdb.Create(adminUser).Error; err != nil {
		t.Fatalf("create admin user: %v", err)
	}
	user := &model.User{Username: "alice", Password: "hashed", Role: model.UserRoleUser, GrowthLevel: 2}
	if err := gdb.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	productA := &model.Product{UserID: adminUser.ID, Name: "Admin Product", Price: 1299, Stock: 8, StartTime: time.Now().Add(-time.Hour)}
	productB := &model.Product{UserID: user.ID, Name: "User Product", Price: 699, Stock: 12, StartTime: time.Now().Add(-time.Hour)}
	if err := gdb.Create(productA).Error; err != nil {
		t.Fatalf("create productA: %v", err)
	}
	if err := gdb.Create(productB).Error; err != nil {
		t.Fatalf("create productB: %v", err)
	}

	paidOrder := &model.Order{UserID: user.ID, ProductID: productA.ID, OrderNum: "ORD-ADMIN-001", Status: model.OrderStatusPaid}
	pendingOrder := &model.Order{UserID: adminUser.ID, ProductID: productB.ID, OrderNum: "ORD-ADMIN-002", Status: model.OrderStatusUnpaid}
	if err := gdb.Create(paidOrder).Error; err != nil {
		t.Fatalf("create paidOrder: %v", err)
	}
	if err := gdb.Create(pendingOrder).Error; err != nil {
		t.Fatalf("create pendingOrder: %v", err)
	}

	paymentA := &model.Payment{OrderID: paidOrder.ID, PaymentID: "PAY-ADMIN-001", AmountCents: 129900, Status: model.PaymentStatusPaid}
	paymentB := &model.Payment{OrderID: pendingOrder.ID, PaymentID: "PAY-ADMIN-002", AmountCents: 69900, Status: model.PaymentStatusPending}
	if err := gdb.Create(paymentA).Error; err != nil {
		t.Fatalf("create paymentA: %v", err)
	}
	if err := gdb.Create(paymentB).Error; err != nil {
		t.Fatalf("create paymentB: %v", err)
	}

	return adminUser, user
}

func mustAdminToken(t *testing.T, user *model.User) string {
	t.Helper()
	token, _, err := utils.GenerateTokens(user.ID, user.Username, user.Role)
	if err != nil {
		t.Fatalf("generate admin token: %v", err)
	}
	return token
}

func mustUserToken(t *testing.T, user *model.User) string {
	t.Helper()
	token, _, err := utils.GenerateTokens(user.ID, user.Username, user.Role)
	if err != nil {
		t.Fatalf("generate user token: %v", err)
	}
	return token
}
