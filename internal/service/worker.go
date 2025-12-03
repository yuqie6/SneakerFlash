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

	"gorm.io/gorm"
)

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

// CreateOderFromMessage 消费秒杀消息：幂等校验 -> 扣减库存 -> 创建订单/支付单 -> 失效缓存，事务失败时回滚 Redis 库存。
func (s *WorkerService) CreateOderFromMessage(msgBytes []byte) error {
	// 1. 解析消息
	var msg SeckillMessage
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return err
	}

	logCtx := logger.ContextWithValues(
		context.Background(),
		"user_id", msg.UserID,
		"product_id", msg.ProductID,
		"order_num", msg.OrderNum,
	)

	var latestStock = -1
	var readyOrderNum string
	var readyOrderID uint
	var readyPaymentID string

	// 2. 开启数据事务
	err := s.db.WithContext(logCtx).Transaction(func(tx *gorm.DB) error {
		// 构建支持事务的 repo
		txProductRepo := repository.NewProductRepo(tx)
		txOrderRepo := repository.NewOrderRepo(tx)
		txPaymentRepo := repository.NewPaymentRepo(tx)

		// 幂等：优先按订单号，再按 user+product
		if msg.OrderNum != "" {
			if existing, err := txOrderRepo.GetByOrderNum(logCtx, msg.OrderNum); err == nil && existing != nil {
				payment, _ := txPaymentRepo.GetByOrderID(logCtx, existing.ID)
				pid := msg.PaymentID
				if payment != nil && payment.PaymentID != "" {
					pid = payment.PaymentID
				}
				readyOrderNum, readyOrderID, readyPaymentID = existing.OrderNum, existing.ID, pid
				return nil
			} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		if existing, err := txOrderRepo.GetByUserAndProduct(logCtx, msg.UserID, msg.ProductID); err == nil && existing != nil {
			payment, _ := txPaymentRepo.GetByOrderID(logCtx, existing.ID)
			pid := msg.PaymentID
			if payment != nil && payment.PaymentID != "" {
				pid = payment.PaymentID
			}
			readyOrderNum, readyOrderID, readyPaymentID = existing.OrderNum, existing.ID, pid
			return nil
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// 扣减数据库库存
		rowsAffected, err := txProductRepo.ReduceStockDB(logCtx, msg.ProductID)
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			slog.WarnContext(logCtx, "库存扣减失败")
			return errors.New("库存不足")
		}

		amountCents := msg.PriceCents
		if amountCents <= 0 {
			product, pErr := txProductRepo.GetByID(logCtx, msg.ProductID)
			if pErr != nil {
				return pErr
			}
			amountCents = int64(math.Round(product.Price * 100))
		}
		if amountCents <= 0 {
			return errors.New("invalid price")
		}

		order := &model.Order{
			UserID:    msg.UserID,
			ProductID: msg.ProductID,
			OrderNum:  msg.OrderNum,
			Status:    model.OrderStatusUnpaid,
		}

		if err := txOrderRepo.Create(logCtx, order); err != nil {
			slog.ErrorContext(logCtx, "创建订单失败", slog.Any("err", err))
			return err
		}

		paymentID := msg.PaymentID
		if paymentID == "" {
			genID, genErr := utils.GenSnowflakeID()
			if genErr != nil {
				return genErr
			}
			paymentID = genID
		}
		payment := &model.Payment{
			OrderID:     order.ID,
			PaymentID:   paymentID,
			AmountCents: amountCents,
			Status:      model.PaymentStatusPending,
		}
		if _, err := txPaymentRepo.CreateIfAbsent(logCtx, payment); err != nil {
			return err
		}

		if updatedProduct, err := txProductRepo.GetByID(logCtx, msg.ProductID); err == nil {
			latestStock = updatedProduct.Stock
		}

		readyOrderNum, readyOrderID, readyPaymentID = order.OrderNum, order.ID, payment.PaymentID
		slog.InfoContext(logCtx, "创建订单成功")
		return nil
	})
	if err != nil {
		rollbackRedisStock(logCtx, msg.ProductID, msg.UserID)
		markPendingOrderFailed(logCtx, msg.OrderNum, err.Error())
		return err
	}

	if readyOrderNum != "" {
		markPendingOrderReady(logCtx, readyOrderNum, readyOrderID, readyPaymentID)
	}

	// 事务成功后刷新缓存/失效详情（worker pool）
	if latestStock >= 0 {
		refreshStockCacheAsync(msg.ProductID, latestStock)
	}
	invalidateProductInfoCache(msg.ProductID)
	return nil
}

// rollbackRedisStock 回补 Redis 库存并移除用户标记，避免库存锁死。
func rollbackRedisStock(ctx context.Context, productID, userID uint) {
	stockKey := fmt.Sprintf("product:stock:%d", productID)
	userSetKey := fmt.Sprintf("product:users:%d", productID)
	redis.RDB.Incr(ctx, stockKey)
	redis.RDB.SRem(ctx, userSetKey, userID)
}
