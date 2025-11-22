package service

import (
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type ProductService struct {
	repo *repository.ProductRepo
	// 归并重复请求, 防止缓存击穿
	sf singleflight.Group
}

var (
	ErrProductNotFound  = errors.New("找不到商品信息")
	ErrProductDuplicate = errors.New("商品已存在")
)

func NewProductService(repo *repository.ProductRepo) *ProductService {
	return &ProductService{
		repo: repo,
	}
}

// 业务 1: 创建商品
func (s *ProductService) CreateProduct(product *model.Product) error {
	if err := s.repo.Create(product); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrProductDuplicate
		}
		return err
	}

	if err := s.SyncStockToRedis(product.ID, product.Stock); err != nil {
		// 预热失败尝试回滚数据库记录，保持一致性
		_ = s.repo.Delete(product.ID)
		return err
	}
	return nil
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
		// 缓存穿透的基本解法, 不存在的商品我们存储 null 空标识符 然后如果val 是null 就直接返回错误, 不给请求访问数据库
		if val == "null" {
			return nil, ErrProductNotFound
		}

		var p model.Product
		if json.Unmarshal([]byte(val), &p) == nil {
			return &p, nil
		}
	}

	// 缓存未命中, 查数据库
	// 用 singleflight 解决缓存击穿, 阻塞重复请求
	raw, err, _ := s.sf.Do(cacheKey, func() (interface{}, error) {
		// 先再查一次redis, 可能会有别的请求正好查完库写入redis
		if val, err := redis.RDB.Get(ctx, cacheKey).Result(); err == nil {
			if val == "null" {
				return nil, ErrProductNotFound
			}
			var p model.Product
			json.Unmarshal([]byte(val), &p)
			return &p, nil
		}

		// 查数据库
		p, err := s.repo.GetByID(id)
		if err != nil {
			// 查不到, 设置 null
			if errors.Is(err, gorm.ErrRecordNotFound) {
				redis.RDB.Set(ctx, cacheKey, "null", time.Minute*5)
				return nil, err
			}

			// 查数据库出错了, 难以定性, 直接返回 err
			return nil, err
		}

		// 将数据库的数据写回 redis
		// 我们要处理缓存雪崩, 最普遍的方法是设置随机过期时间
		expriation := time.Hour + time.Duration(rand.Intn(1800)*int(time.Second))

		data, _ := json.Marshal(p)
		redis.RDB.Set(ctx, cacheKey, data, expriation)
		return p, nil
	})

	if err != nil {
		return nil, err
	}

	product, ok := raw.(*model.Product)
	if !ok {
		return nil, fmt.Errorf("invalid product data type")
	}
	return product, nil
}
