package service

import (
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"
)

type WorkerService struct {
	db          *gorm.DB
	productRepo *repository.ProductRepo
	orderRepo   *repository.OrderRepo
}

func NewWorkerService(db *gorm.DB, productRepo *repository.ProductRepo, order *repository.OrderRepo) *WorkerService {
	return &WorkerService{
		db:          db,
		productRepo: productRepo,
		orderRepo:   order,
	}
}

func (s *WorkerService) CreateOderFromMessage(msgBytes []byte) error {
	// 1. 解析消息
	var msg SeckillMessage
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return err
	}

	// 2. 开启数据事务
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 构建支持事务的 repo
		txProductRepo := repository.NewProductRepo(tx)
		txOrderRepo := repository.NewOrderRepo(tx)

		// 扣减数据库库存
		rowsAffected, err := txProductRepo.ReduceStockDB(msg.ProductID)
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			log.Printf("[ERROR]库存扣减失败: UserID = %d ProductID = %d", msg.UserID, msg.ProductID)
			return errors.New("库存不足")
		}

		// 创建订单
		order := &model.Order{
			UserID:    msg.UserID,
			ProductID: msg.ProductID,
			OrderNum:  msg.OrderNum,
			Status:    model.StatusUpaid,
		}

		if err := txOrderRepo.Create(order); err != nil {
			log.Printf("[ERROR]创建订单失败: %v", err)
			return err
		}

		log.Printf("[INFO] 创建订单成功: userID: %d productID: %d", msg.UserID, msg.ProductID)
		return nil
	})
	if err != nil {
		// 事务失败，回滚缓存库存，避免用户库存被锁死
		ctx := context.Background()
		stockKey := fmt.Sprintf("product:stock:%d", msg.ProductID)
		userSetKey := fmt.Sprintf("product:users:%d", msg.ProductID)
		redis.RDB.Incr(ctx, stockKey)
		redis.RDB.SRem(ctx, userSetKey, msg.UserID)
	}
	return err
}
