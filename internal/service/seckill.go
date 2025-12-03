package service

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/infra/kafka"
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/repository"
	"context"
	"encoding/json"
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
}

func NewSeckillService(db *gorm.DB, productRepo *repository.ProductRepo) *SeckillService {
	return &SeckillService{
		db:          db,
		productRepo: productRepo,
	}
}

var (
	ErrSeckillRepeat   = errors.New("您已经抢购过该商品")
	ErrSeckillFull     = errors.New("手慢无, 商品已经售罄")
	ErrSeckillBusy     = errors.New("系统繁忙, 请稍后重试")
	ErrSeckillNotStart = errors.New("活动尚未开始")
)

type SeckillResult struct {
	OrderID   uint   `json:"order_id,omitempty"`
	OrderNum  string `json:"order_num"`
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
}

// Seckill 秒杀扣减库存并投递消息，由 worker 落库；Redis 原子扣减保护库存，投递失败回滚库存。
func (s *SeckillService) Seckill(ctx context.Context, userID, productID uint) (*SeckillResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	// 0. 校验商品存在与开始时间
	product, err := s.productRepo.GetByID(ctx, productID)
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

	// 4. 抢到了, 生成订单号/支付号，准备消息
	orderNum, err := utils.GenSnowflakeID()
	if err != nil {
		return nil, ErrSeckillBusy
	}
	paymentID, err := utils.GenSnowflakeID()
	if err != nil {
		return nil, ErrSeckillBusy
	}
	priceCents := int64(math.Round(product.Price * 100))
	if priceCents <= 0 {
		return nil, ErrSeckillBusy
	}

	msg := SeckillMessage{
		UserID:     userID,
		ProductID:  productID,
		OrderNum:   orderNum,
		PaymentID:  paymentID,
		PriceCents: priceCents,
		Time:       time.Now(),
	}

	msgBytes, _ := json.Marshal(msg)
	if err := kafka.Send(config.Conf.Data.Kafka.Topic, string(msgBytes)); err != nil {
		// 投递失败回滚 Redis 库存/用户标记，避免锁死
		redis.RDB.Incr(ctx, stockKey)
		redis.RDB.SRem(ctx, userSetKey, userID)
		return nil, ErrSeckillBusy
	}

	// 预写 pending 状态，便于前端轮询
	_ = setPendingOrder(ctx, PendingOrderCache{
		OrderNum:   orderNum,
		PaymentID:  paymentID,
		ProductID:  productID,
		UserID:     userID,
		PriceCents: priceCents,
		Status:     PendingStatusPending,
	})

	// 异步刷新商品缓存（worker pool + 去重）
	invalidateProductInfoCache(productID)

	return &SeckillResult{
		OrderNum:  orderNum,
		PaymentID: paymentID,
		Status:    string(PendingStatusPending),
	}, nil
}
