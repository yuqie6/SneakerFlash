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
	"strconv"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type ProductService struct {
	repo *repository.ProductRepo
	// 归并重复请求, 防止缓存击穿
	sf *singleflight.Group
}

var (
	ErrProductNotFound  = errors.New("找不到商品信息")
	ErrProductDuplicate = errors.New("商品已存在")
)

func NewProductService(repo *repository.ProductRepo) *ProductService {
	return &ProductService{
		repo: repo,
		sf:   &singleflight.Group{},
	}
}

func (s *ProductService) WithContext(ctx context.Context) *ProductService {
	if ctx == nil {
		return s
	}
	return &ProductService{
		repo: s.repo.WithContext(ctx),
		sf:   s.sf,
	}
}

// CreateProduct 创建商品；若 Redis 预热失败会回滚数据库记录以保持库存一致性。
func (s *ProductService) CreateProduct(product *model.Product) error {
	if err := s.repo.Create(product); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || isMySQLDuplicate(err) {
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

// SyncStockToRedis 将库存同步到 Redis，作为秒杀读写的唯一实时源。
func (s *ProductService) SyncStockToRedis(id uint, stock int) error {
	return setStockCache(id, stock)
}

// ListProducts 分页查询商品列表。
func (s *ProductService) ListProducts(page, pageSize int) ([]model.Product, int64, error) {
	return s.repo.List(page, pageSize)
}

// GetProductByID 优先读缓存，singleflight 防击穿，null 哨兵防穿透，随机 TTL 防雪崩；库存以 Redis 为准。
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
	// 以 redis 实时库存为准，避免详情页显示旧库存
	if stockStr, err := redis.RDB.Get(ctx, fmt.Sprintf("product:stock:%d", product.ID)).Result(); err == nil {
		if v, convErr := strconv.Atoi(stockStr); convErr == nil {
			product.Stock = v
		}
	}
	return product, nil
}

// UpdateProduct 仅允许创建者更新，未命中则视为不存在。
func (s *ProductService) UpdateProduct(userID, id uint, data map[string]any) error {
	if len(data) == 0 {
		return nil
	}
	rows, err := s.repo.UpdateByUser(id, userID, data)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProductNotFound
	}
	return nil
}

// DeleteProduct 删除指定用户的商品。
func (s *ProductService) DeleteProduct(userID, id uint) error {
	p, err := s.repo.GetByIDAndUser(id, userID)
	if err != nil {
		return err
	}
	return s.repo.Delete(p.ID)
}

// ListUserProducts 查询用户发布的商品列表。
func (s *ProductService) ListUserProducts(userID uint, page, size int) ([]model.Product, int64, error) {
	return s.repo.ListByUserID(userID, page, size)
}
