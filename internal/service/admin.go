package service

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"context"
	"fmt"

	"gorm.io/gorm"
)

type AdminStats struct {
	TotalUsers        int64 `json:"total_users"`
	TotalOrders       int64 `json:"total_orders"`
	TotalRevenueCents int64 `json:"total_revenue_cents"`
	TotalProducts     int64 `json:"total_products"`
	PendingOrders     int64 `json:"pending_orders"`
}

type AdminService struct {
	db          *gorm.DB
	userRepo    *repository.UserRepo
	orderRepo   *repository.OrderRepo
	productRepo *repository.ProductRepo
}

func NewAdminService(db *gorm.DB, userRepo *repository.UserRepo, productRepo *repository.ProductRepo) *AdminService {
	return &AdminService{
		db:          db,
		userRepo:    userRepo,
		orderRepo:   repository.NewOrderRepo(db),
		productRepo: productRepo,
	}
}

func (s *AdminService) Stats(ctx context.Context) (*AdminStats, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	totalUsers, err := s.userRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}
	totalOrders, err := s.orderRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}
	totalRevenue, err := s.orderRepo.SumRevenue(ctx)
	if err != nil {
		return nil, err
	}
	totalProducts, err := s.productRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}
	pendingOrders, err := s.orderRepo.CountByStatus(ctx, model.OrderStatusUnpaid)
	if err != nil {
		return nil, err
	}

	return &AdminStats{
		TotalUsers:        totalUsers,
		TotalOrders:       totalOrders,
		TotalRevenueCents: totalRevenue,
		TotalProducts:     totalProducts,
		PendingOrders:     pendingOrders,
	}, nil
}

func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	if ctx == nil {
		return nil, 0, fmt.Errorf("context is nil")
	}
	page, pageSize = normalizePage(page, pageSize)
	return s.userRepo.ListAll(ctx, page, pageSize)
}

func (s *AdminService) ListAllOrders(ctx context.Context, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error) {
	if ctx == nil {
		return nil, 0, fmt.Errorf("context is nil")
	}
	page, pageSize = normalizePage(page, pageSize)
	return s.orderRepo.ListAll(ctx, status, page, pageSize)
}

func (s *AdminService) ListAllProducts(ctx context.Context, page, pageSize int) ([]model.Product, int64, error) {
	if ctx == nil {
		return nil, 0, fmt.Errorf("context is nil")
	}
	page, pageSize = normalizePage(page, pageSize)
	return s.productRepo.ListAll(ctx, page, pageSize)
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}
