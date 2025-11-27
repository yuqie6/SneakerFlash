package service

import (
	"SneakerFlash/internal/infra/redis"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// pending 状态缓存 TTL，避免长时间占用内存。
const pendingOrderTTL = 10 * time.Minute

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
func setStockCache(productID uint, stock int) error {
	ctx := context.Background()
	key := fmt.Sprintf("product:stock:%d", productID)
	return redis.RDB.Set(ctx, key, stock, 0).Err()
}

func pendingOrderKey(orderNum string) string {
	return fmt.Sprintf("order:pending:%s", orderNum)
}

// setPendingOrder 缓存订单处理状态。
func setPendingOrder(ctx context.Context, payload PendingOrderCache) error {
	if ctx == nil {
		ctx = context.Background()
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
		ctx = context.Background()
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

// markPendingOrderReady 更新缓存为 ready。
func markPendingOrderReady(ctx context.Context, orderNum string, orderID uint, paymentID string) {
	_ = setPendingOrder(ctx, PendingOrderCache{
		OrderNum:  orderNum,
		OrderID:   orderID,
		PaymentID: paymentID,
		Status:    PendingStatusReady,
	})
}

// markPendingOrderFailed 标记处理失败，便于前端提示。
func markPendingOrderFailed(ctx context.Context, orderNum, message string) {
	_ = setPendingOrder(ctx, PendingOrderCache{
		OrderNum: orderNum,
		Status:   PendingStatusFailed,
		Message:  message,
	})
}

// refreshStockCacheAsync 异步刷新库存缓存，失败忽略以避免阻塞主流程。
func refreshStockCacheAsync(productID uint, stock int) {
	go func() {
		_ = setStockCache(productID, stock)
	}()
}

// invalidateProductInfoCache 失效商品详情缓存，促使后续请求回源数据库。
func invalidateProductInfoCache(productID uint) {
	ctx := context.Background()
	key := fmt.Sprintf("product:info:%d", productID)
	redis.RDB.Del(ctx, key)
}
