package service

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/infra/kafka"
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_redis "github.com/redis/go-redis/v9"
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

type SeckillService struct{}

func NewSeckillService() *SeckillService {
	return &SeckillService{}
}

var (
	ErrSeckillRepeat = errors.New("您已经抢购过该商品")
	ErrSeckillFull   = errors.New("手慢无, 商品已经售罄")
	ErrSeckillBusy   = errors.New("系统繁忙, 请稍后重试")
)

// kafka 消息结构体
type SeckillMessage struct {
	UserID    uint      `json:"user_id"`
	ProductID uint      `json:"product_id"`
	OrderNum  string    `json:"order_num"` // 订单号
	Time      time.Time `json:"time"`
}

// 秒杀具体逻辑
func (s *SeckillService) Seckill(userID, productID uint) (string, error) {
	ctx := context.Background()

	// 1. 准备 redis key
	stockKey := fmt.Sprintf("product:stock:%d", productID)
	userSetKey := fmt.Sprintf("product:users:%d", productID)

	// 2. 执行 lua 脚本
	res, err := seckillScript.Run(ctx, redis.RDB, []string{stockKey, userSetKey}, userID).Int()
	if err != nil {
		return "", err
	}

	// 3. 处理 lua 结果
	switch res {
	case -1:
		return "", ErrSeckillRepeat
	case 0:
		return "", ErrSeckillFull
	}

	// 4. 抢到了, 需要给 kafka 消息, 创建订单

	orderNum, err := utils.GenSnowflakeID()
	if err != nil {
		return "", ErrSeckillBusy
	}

	msg := SeckillMessage{
		UserID:    userID,
		ProductID: productID,
		OrderNum:  orderNum,
		Time:      time.Now(),
	}

	msgBytes, _ := json.Marshal(msg)

	// 5. 投递给kafka
	err = kafka.Send(config.Conf.Data.Kafka.Topic, string(msgBytes))
	if err != nil {
		// 如果 kafka 发送失败, 必须回滚 redis 库存, 暂时简单的手动添加库存, 然后删掉用户的缓存
		redis.RDB.Incr(ctx, stockKey)
		redis.RDB.SRem(ctx, userSetKey, userID)

		return "", ErrSeckillBusy
	}
	return orderNum, nil
}
