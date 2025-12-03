package service

import (
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/logger"
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	_redis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// WorkerService Kafka 消费者服务，负责秒杀消息落库。
type WorkerService struct {
	db          *gorm.DB
	productRepo *repository.ProductRepo
	orderRepo   *repository.OrderRepo
	paymentRepo *repository.PaymentRepo
}

// NewWorkerService 构建异步消费服务，处理秒杀队列落库。
func NewWorkerService(db *gorm.DB, productRepo *repository.ProductRepo, order *repository.OrderRepo) *WorkerService {
	return &WorkerService{
		db:          db,
		productRepo: productRepo,
		orderRepo:   order,
		paymentRepo: repository.NewPaymentRepo(db),
	}
}

// rollbackRedisStock 回补 Redis 库存并移除用户标记，避免库存锁死。
func rollbackRedisStock(ctx context.Context, productID, userID uint) {
	stockKey := fmt.Sprintf("product:stock:%d", productID)
	userSetKey := fmt.Sprintf("product:users:%d", productID)
	redis.RDB.Incr(ctx, stockKey)
	redis.RDB.SRem(ctx, userSetKey, userID)
}

// ========== 批量插入实现 ==========

// BatchCreateOrdersFromMessages 批量处理秒杀消息：解析 -> 幂等过滤 -> 批量扣库存 -> 批量插入订单/支付单
func (s *WorkerService) BatchCreateOrdersFromMessages(msgBodies [][]byte) error {
	if len(msgBodies) == 0 {
		return nil
	}

	ctx := logger.ContextWithValues(context.Background(), "batch_size", len(msgBodies))
	startTime := time.Now()

	// 1. 解析所有消息
	msgs := make([]*SeckillMessage, 0, len(msgBodies))
	for _, body := range msgBodies {
		var msg SeckillMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			slog.WarnContext(ctx, "消息解析失败，跳过", slog.String("error", err.Error()))
			continue
		}
		msgs = append(msgs, &msg)
	}

	if len(msgs) == 0 {
		return nil
	}

	// 记录处理结果
	results := make([]orderResult, 0, len(msgs))

	// 2. 开启数据库事务
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txOrderRepo := repository.NewOrderRepo(tx)
		txPaymentRepo := repository.NewPaymentRepo(tx)
		txProductRepo := repository.NewProductRepo(tx)

		// 2.1 批量幂等检查 - 按 order_num 查询已存在的订单
		orderNums := make([]string, 0, len(msgs))
		for _, msg := range msgs {
			if msg.OrderNum != "" {
				orderNums = append(orderNums, msg.OrderNum)
			}
		}

		existingOrderMap := make(map[string]*model.Order)
		if len(orderNums) > 0 {
			existingOrders, err := txOrderRepo.GetByOrderNums(ctx, orderNums)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("批量查询订单失败: %w", err)
			}
			for _, order := range existingOrders {
				existingOrderMap[order.OrderNum] = order
			}
		}

		// 2.2 过滤已存在的订单，收集新订单
		newMsgs := make([]*SeckillMessage, 0, len(msgs))
		for _, msg := range msgs {
			if existing, ok := existingOrderMap[msg.OrderNum]; ok {
				// 已存在，直接记录成功
				payment, _ := txPaymentRepo.GetByOrderID(ctx, existing.ID)
				pid := msg.PaymentID
				if payment != nil && payment.PaymentID != "" {
					pid = payment.PaymentID
				}
				results = append(results, orderResult{
					orderNum:  existing.OrderNum,
					orderID:   existing.ID,
					paymentID: pid,
					success:   true,
				})
				continue
			}
			newMsgs = append(newMsgs, msg)
		}

		if len(newMsgs) == 0 {
			slog.InfoContext(ctx, "所有订单已存在，无需处理")
			return nil
		}

		// 2.3 按 productID 分组统计扣库存数量
		stockDeductions := make(map[uint]int64)
		for _, msg := range newMsgs {
			stockDeductions[msg.ProductID]++
		}

		// 2.4 批量扣减库存
		productStocks := make(map[uint]int) // 记录扣减后的库存
		for productID, count := range stockDeductions {
			rowsAffected, err := txProductRepo.ReduceStockDBBatch(ctx, productID, int(count))
			if err != nil {
				return fmt.Errorf("扣减库存失败 productID=%d: %w", productID, err)
			}
			if rowsAffected == 0 {
				// 库存不足，需要标记该商品的所有消息失败
				for _, msg := range newMsgs {
					if msg.ProductID == productID {
						results = append(results, orderResult{
							orderNum: msg.OrderNum,
							success:  false,
							errMsg:   "库存不足",
						})
					}
				}
				// 过滤掉该商品的消息
				filtered := make([]*SeckillMessage, 0, len(newMsgs))
				for _, msg := range newMsgs {
					if msg.ProductID != productID {
						filtered = append(filtered, msg)
					}
				}
				newMsgs = filtered
				continue
			}
			// 获取最新库存用于缓存刷新
			if product, err := txProductRepo.GetByID(ctx, productID); err == nil {
				productStocks[productID] = product.Stock
			}
		}

		if len(newMsgs) == 0 {
			return nil
		}

		// 2.5 构建订单列表
		orders := make([]*model.Order, 0, len(newMsgs))
		for _, msg := range newMsgs {
			orders = append(orders, &model.Order{
				UserID:    msg.UserID,
				ProductID: msg.ProductID,
				OrderNum:  msg.OrderNum,
				Status:    model.OrderStatusUnpaid,
			})
		}

		// 2.6 批量插入订单
		if err := tx.CreateInBatches(orders, 500).Error; err != nil {
			return fmt.Errorf("批量创建订单失败: %w", err)
		}

		// 2.7 构建支付单列表
		payments := make([]*model.Payment, 0, len(newMsgs))
		for i, msg := range newMsgs {
			paymentID := msg.PaymentID
			if paymentID == "" {
				genID, err := utils.GenSnowflakeID()
				if err != nil {
					return fmt.Errorf("生成支付ID失败: %w", err)
				}
				paymentID = genID
			}

			amountCents := msg.PriceCents
			if amountCents <= 0 {
				product, err := txProductRepo.GetByID(ctx, msg.ProductID)
				if err != nil {
					return fmt.Errorf("获取商品价格失败: %w", err)
				}
				amountCents = int64(math.Round(product.Price * 100))
			}

			payments = append(payments, &model.Payment{
				OrderID:     orders[i].ID, // 批量插入后 GORM 会填充 ID
				PaymentID:   paymentID,
				AmountCents: amountCents,
				Status:      model.PaymentStatusPending,
			})
		}

		// 2.8 批量插入支付单
		if err := tx.CreateInBatches(payments, 500).Error; err != nil {
			return fmt.Errorf("批量创建支付单失败: %w", err)
		}

		// 2.9 收集成功结果
		for i, order := range orders {
			results = append(results, orderResult{
				orderNum:  order.OrderNum,
				orderID:   order.ID,
				paymentID: payments[i].PaymentID,
				success:   true,
			})
		}

		// 2.10 异步刷新库存缓存
		for productID, stock := range productStocks {
			refreshStockCacheAsync(productID, stock)
			invalidateProductInfoCache(productID)
		}

		return nil
	})

	// 3. 处理事务结果
	if err != nil {
		slog.ErrorContext(ctx, "批量事务失败", slog.Any("error", err))
		// 事务失败，回滚所有 Redis 库存
		for _, msg := range msgs {
			rollbackRedisStock(ctx, msg.ProductID, msg.UserID)
			markPendingOrderFailed(ctx, msg.OrderNum, err.Error())
		}
		return err
	}

	// 4. 批量更新 Redis pending 状态
	s.batchUpdatePendingStatus(ctx, results)

	elapsed := time.Since(startTime)
	slog.InfoContext(ctx, "批量处理完成",
		slog.Int("total", len(msgs)),
		slog.Int("success", countSuccess(results)),
		slog.Duration("elapsed", elapsed),
	)

	return nil
}

// batchUpdatePendingStatus 使用 Pipeline 批量更新 Redis pending 状态
func (s *WorkerService) batchUpdatePendingStatus(ctx context.Context, results []orderResult) {
	if len(results) == 0 {
		return
	}

	pipe := redis.RDB.Pipeline()
	for _, r := range results {
		var cache PendingOrderCache
		if r.success {
			cache = PendingOrderCache{
				OrderNum:  r.orderNum,
				OrderID:   r.orderID,
				PaymentID: r.paymentID,
				Status:    PendingStatusReady,
			}
		} else {
			cache = PendingOrderCache{
				OrderNum: r.orderNum,
				Status:   PendingStatusFailed,
				Message:  r.errMsg,
			}
		}
		data, _ := json.Marshal(cache)
		key := pendingOrderKey(r.orderNum)
		pipe.Set(ctx, key, data, pendingOrderTTL)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, _redis.Nil) {
		slog.WarnContext(ctx, "批量更新 pending 状态失败", slog.Any("error", err))
	}
}

// orderResult 批量处理结果记录
type orderResult struct {
	orderNum  string
	orderID   uint
	paymentID string
	success   bool
	errMsg    string
}

func countSuccess(results []orderResult) int {
	count := 0
	for _, r := range results {
		if r.success {
			count++
		}
	}
	return count
}
