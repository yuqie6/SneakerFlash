package service

import "SneakerFlash/internal/repository"

type ProductService struct {
	repo *repository.ProductRepo
}

func NewProductService(repo *repository.ProductRepo) *ProductService {
	return &ProductService{
		repo: repo,
	}
}
