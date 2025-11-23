package service

import (
	"SneakerFlash/internal/infra/redis"
	"context"
	"fmt"
)

// setStockCache 覆盖写入商品库存缓存（秒杀用）
func setStockCache(productID uint, stock int) error {
	ctx := context.Background()
	key := fmt.Sprintf("product:stock:%d", productID)
	return redis.RDB.Set(ctx, key, stock, 0).Err()
}

// refreshStockCacheAsync 异步刷新库存缓存，忽略错误
func refreshStockCacheAsync(productID uint, stock int) {
	go func() {
		_ = setStockCache(productID, stock)
	}()
}

// invalidateProductInfoCache 让详情重新回源数据库
func invalidateProductInfoCache(productID uint) {
	ctx := context.Background()
	key := fmt.Sprintf("product:info:%d", productID)
	redis.RDB.Del(ctx, key)
}
