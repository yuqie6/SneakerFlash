package service

import "SneakerFlash/internal/repository"

type OrderService struct {
	repo *repository.OrderRepo
}

func NewOrderService(repo *repository.OrderRepo) *OrderService {
	return &OrderService{
		repo: repo,
	}
}
