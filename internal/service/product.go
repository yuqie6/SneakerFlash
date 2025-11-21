package service

import (
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type ProductService struct {
	repo *repository.ProductRepo
}

func NewProductService(repo *repository.ProductRepo) *ProductService {
	return &ProductService{
		repo: repo,
	}
}

// 业务 1: 创建商品
func (s *ProductService) CreateProduct(product *model.Product) error {
	if err := s.repo.Create(product); err != nil {
		return err
	}

	return s.SyncStockToRedis(product.ID, product.Stock)
}

// 业务 2: 库存预热(核心)
func (s *ProductService) SyncStockToRedis(id uint, stock int) error {
	ctx := context.Background()
	key := fmt.Sprintf("product:stock:%d", id)

	// 直接覆盖
	return redis.RDB.Set(ctx, key, stock, 0).Err() // 永不结束
}

// 业务 3: 获取列表
func (s *ProductService) ListProducts(page, pageSize int) ([]model.Product, int64, error) {
	return s.repo.List(page, pageSize)
}

// 业务 4: 获取商品详情
// 查缓存 -> 命中返回 -> 未命中查库 -> 写缓存 -> 返回
func (s *ProductService) GetProductByID(id uint) (*model.Product, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("product:info:%d", id)

	// 查 redis
	val, err := redis.RDB.Get(ctx, cacheKey).Result()
	if err == nil {
		// 命中缓存了
		var p model.Product
		if json.Unmarshal([]byte(val), &p) == nil {
			return &p, nil
		}
	}

	// 缓存未命中, 查数据库
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 写回 redis
	data, _ := json.Marshal(p)
	redis.RDB.Set(ctx, cacheKey, data, time.Hour)

	return p, nil
}
