package service

import (
	"SneakerFlash/internal/infra/redis"
	"context"
	"fmt"
)

// setStockCache 覆盖写入商品库存缓存（秒杀读取入口）。
func setStockCache(productID uint, stock int) error {
	ctx := context.Background()
	key := fmt.Sprintf("product:stock:%d", productID)
	return redis.RDB.Set(ctx, key, stock, 0).Err()
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
