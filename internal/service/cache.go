package service

import (
	"SneakerFlash/internal/infra/redis"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// pending 状态缓存 TTL，避免长时间占用内存。
const pendingOrderTTL = 10 * time.Minute

// 缓存 worker pool 配置
const (
	cacheWorkerCount  = 10    // worker 数量
	cacheTaskChanSize = 10000 // channel 缓冲大小
)

var (
	// 缓存失效任务 channel
	cacheInvalidateChan = make(chan uint, cacheTaskChanSize)
	// 去重：正在处理或待处理的 productID
	pendingInvalidate sync.Map
	// 确保 invalidate worker 只启动一次
	cacheInvalidateWorkerOnce sync.Once

	// 库存刷新任务 channel
	stockRefreshChan = make(chan uint, cacheTaskChanSize)
	// 存储每个 productID 最新的 stock 值
	pendingStockRefresh sync.Map
	// 确保 stock refresh worker 只启动一次
	stockRefreshWorkerOnce sync.Once
)

// initCacheInvalidateWorkers 启动缓存失效 worker pool
func initCacheInvalidateWorkers() {
	for range cacheWorkerCount {
		go cacheInvalidateWorker()
	}
}

// cacheInvalidateWorker 消费缓存失效任务
func cacheInvalidateWorker() {
	for productID := range cacheInvalidateChan {
		ctx := context.Background()
		key := fmt.Sprintf("product:info:%d", productID)
		redis.RDB.Del(ctx, key)
		// 删除完成，允许后续同 ID 任务进入
		pendingInvalidate.Delete(productID)
	}
}

// initStockRefreshWorkers 启动库存刷新 worker pool
func initStockRefreshWorkers() {
	for range cacheWorkerCount {
		go stockRefreshWorker()
	}
}

// stockRefreshWorker 消费库存刷新任务
func stockRefreshWorker() {
	for productID := range stockRefreshChan {
		// 取出最新的 stock 值
		val, ok := pendingStockRefresh.LoadAndDelete(productID)
		if !ok {
			continue
		}
		stock := val.(int)
		_ = setStockCache(context.Background(), productID, stock)
	}
}

type PendingOrderStatus string

const (
	PendingStatusPending PendingOrderStatus = "pending"
	PendingStatusReady   PendingOrderStatus = "ready"
	PendingStatusFailed  PendingOrderStatus = "failed"
)

// PendingOrderCache 记录入口排队结果，便于前端轮询拿到 order_id/payment_id。
type PendingOrderCache struct {
	OrderNum   string             `json:"order_num"`
	OrderID    uint               `json:"order_id,omitempty"`
	PaymentID  string             `json:"payment_id"`
	ProductID  uint               `json:"product_id,omitempty"`
	UserID     uint               `json:"user_id,omitempty"`
	PriceCents int64              `json:"price_cents,omitempty"`
	Status     PendingOrderStatus `json:"status"`
	Message    string             `json:"message,omitempty"`
}

// setStockCache 覆盖写入商品库存缓存（秒杀读取入口）。
func setStockCache(ctx context.Context, productID uint, stock int) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	key := fmt.Sprintf("product:stock:%d", productID)
	return redis.RDB.Set(ctx, key, stock, 0).Err()
}

func pendingOrderKey(orderNum string) string {
	return fmt.Sprintf("order:pending:%s", orderNum)
}

// setPendingOrder 缓存订单处理状态。
func setPendingOrder(ctx context.Context, payload PendingOrderCache) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	key := pendingOrderKey(payload.OrderNum)
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return redis.RDB.Set(ctx, key, data, pendingOrderTTL).Err()
}

// getPendingOrder 读取订单处理状态缓存。
func getPendingOrder(ctx context.Context, orderNum string) (*PendingOrderCache, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	key := pendingOrderKey(orderNum)
	res, err := redis.RDB.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var cache PendingOrderCache
	if uErr := json.Unmarshal([]byte(res), &cache); uErr != nil {
		return nil, uErr
	}
	return &cache, nil
}

// markPendingOrderFailed 标记处理失败，便于前端提示。
func markPendingOrderFailed(ctx context.Context, orderNum, message string) {
	_ = setPendingOrder(ctx, PendingOrderCache{
		OrderNum: orderNum,
		Status:   PendingStatusFailed,
		Message:  message,
	})
}

// refreshStockCacheAsync 异步刷新库存缓存，使用 worker pool + 最新值覆盖。
func refreshStockCacheAsync(productID uint, stock int) {
	// 延迟初始化 worker pool
	stockRefreshWorkerOnce.Do(initStockRefreshWorkers)

	// 存储最新的 stock 值（覆盖旧值）
	_, alreadyPending := pendingStockRefresh.Swap(productID, stock)

	// 如果已经在队列中，不需要重复发送
	if alreadyPending {
		return
	}

	// 非阻塞发送
	select {
	case stockRefreshChan <- productID:
	default:
		pendingStockRefresh.Delete(productID) // 没进队列，清除标记
	}
}

// invalidateProductInfoCache 异步失效商品详情缓存，使用 worker pool + 去重。
func invalidateProductInfoCache(productID uint) {
	// 延迟初始化 worker pool
	cacheInvalidateWorkerOnce.Do(initCacheInvalidateWorkers)

	// 去重：如果该 productID 已在队列中，直接跳过
	if _, exists := pendingInvalidate.LoadOrStore(productID, struct{}{}); exists {
		return
	}

	// 非阻塞发送，channel 满了就丢弃（worker 成功后还会再刷新）
	select {
	case cacheInvalidateChan <- productID:
	default:
		pendingInvalidate.Delete(productID) // 没进队列，清除标记
	}
}
