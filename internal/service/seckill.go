package service

import (
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/repository"
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	_redis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// lua 脚本: 原子检查库存, 扣减, 记录用户
// key1 商品库存
// key2 商品购买用户
// argv1 用户 id
var seckillScript = _redis.NewScript(`
	-- 1. 检查用户是否已经抢购过
	if redis.call("SISMEMBER", KEYS[2], ARGV[1]) == 1 then
		return -1 -- 重复抢购
	end

	-- 2. 检查库存是否充足
	local stock = tonumber(redis.call("GET", KEYS[1]))
	if stock == nil or stock <= 0 then
		return 0 -- 库存不足
	end

	-- 3. 扣减库存
	redis.call("DECR", KEYS[1])
	
	-- 4. 记录该用户已经抢购
	redis.call("SADD", KEYS[2], ARGV[1])
	return 1
`)

type SeckillService struct {
	db          *gorm.DB
	productRepo *repository.ProductRepo
	orderRepo   *repository.OrderRepo
	paymentRepo *repository.PaymentRepo
}

func NewSeckillService(db *gorm.DB, productRepo *repository.ProductRepo) *SeckillService {
	return &SeckillService{
		db:          db,
		productRepo: productRepo,
		orderRepo:   repository.NewOrderRepo(db),
		paymentRepo: repository.NewPaymentRepo(db),
	}
}

// WithContext 绑定请求上下文，确保事务和仓储日志携带 request_id。
func (s *SeckillService) WithContext(ctx context.Context) *SeckillService {
	if ctx == nil {
		return s
	}
	ctxDB := s.db.WithContext(ctx)
	return &SeckillService{
		db:          ctxDB,
		productRepo: s.productRepo.WithContext(ctx),
		orderRepo:   s.orderRepo.WithContext(ctx),
		paymentRepo: s.paymentRepo.WithContext(ctx),
	}
}

var (
	ErrSeckillRepeat   = errors.New("您已经抢购过该商品")
	ErrSeckillFull     = errors.New("手慢无, 商品已经售罄")
	ErrSeckillBusy     = errors.New("系统繁忙, 请稍后重试")
	ErrSeckillNotStart = errors.New("活动尚未开始")
)

type SeckillResult struct {
	OrderID   uint   `json:"order_id"`
	OrderNum  string `json:"order_num"`
	PaymentID string `json:"payment_id"`
}

// Seckill 秒杀扣减库存并同步创建订单+支付单；Redis 原子扣减保护库存，事务失败会回滚缓存库存。
func (s *SeckillService) Seckill(userID, productID uint) (*SeckillResult, error) {
	ctx := context.Background()
	if s.db != nil && s.db.Statement != nil && s.db.Statement.Context != nil {
		ctx = s.db.Statement.Context
	}

	// 0. 校验商品存在与开始时间
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	if time.Now().Before(product.StartTime) {
		return nil, ErrSeckillNotStart
	}

	// 1. 准备 redis key
	stockKey := fmt.Sprintf("product:stock:%d", productID)
	userSetKey := fmt.Sprintf("product:users:%d", productID)

	// 2. 执行 lua 脚本
	res, err := seckillScript.Run(ctx, redis.RDB, []string{stockKey, userSetKey}, userID).Int()
	if err != nil {
		return nil, err
	}

	// 3. 处理 lua 结果
	switch res {
	case -1:
		return nil, ErrSeckillRepeat
	case 0:
		return nil, ErrSeckillFull
	}

	// 4. 抢到了, 创建订单与支付单
	orderNum, err := utils.GenSnowflakeID()
	if err != nil {
		return nil, ErrSeckillBusy
	}

	var result SeckillResult
	latestStock := -1

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		txProductRepo := repository.NewProductRepo(tx)
		txOrderRepo := repository.NewOrderRepo(tx)
		txPaymentRepo := repository.NewPaymentRepo(tx)

		// 幂等：若已存在订单则直接返回
		if existing, err := txOrderRepo.GetByUserAndProduct(userID, productID); err == nil && existing != nil {
			payment, _ := txPaymentRepo.GetByOrderID(existing.ID)
			paymentID := ""
			if payment != nil {
				paymentID = payment.PaymentID
			}
			result = SeckillResult{
				OrderID:   existing.ID,
				OrderNum:  existing.OrderNum,
				PaymentID: paymentID,
			}
			return nil
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		rowsAffected, err := txProductRepo.ReduceStockDB(productID)
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return ErrSeckillFull
		}

		amountCents := int64(math.Round(product.Price * 100))
		if amountCents <= 0 {
			return ErrSeckillBusy
		}

		order := &model.Order{
			UserID:    userID,
			ProductID: productID,
			OrderNum:  orderNum,
			Status:    model.OrderStatusUnpaid,
		}
		if err := txOrderRepo.Create(order); err != nil {
			return err
		}

		paymentID, err := utils.GenSnowflakeID()
		if err != nil {
			return err
		}
		payment := &model.Payment{
			OrderID:     order.ID,
			PaymentID:   paymentID,
			AmountCents: amountCents,
			Status:      model.PaymentStatusPending,
		}
		payment, err = txPaymentRepo.CreateIfAbsent(payment)
		if err != nil {
			return err
		}

		if updatedProduct, err := txProductRepo.GetByID(productID); err == nil {
			latestStock = updatedProduct.Stock
		}

		result = SeckillResult{
			OrderID:   order.ID,
			OrderNum:  order.OrderNum,
			PaymentID: payment.PaymentID,
		}
		return nil
	})

	if txErr != nil {
		// 事务失败，回滚缓存库存，避免用户库存被锁死
		redis.RDB.Incr(ctx, stockKey)
		redis.RDB.SRem(ctx, userSetKey, userID)

		if errors.Is(txErr, ErrSeckillFull) {
			return nil, ErrSeckillFull
		}
		return nil, ErrSeckillBusy
	}

	// 异步刷新缓存库存与商品详情
	if latestStock >= 0 {
		refreshStockCacheAsync(productID, latestStock)
	}
	go invalidateProductInfoCache(productID)

	return &result, nil
}
