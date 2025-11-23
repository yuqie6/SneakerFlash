package service

import (
	"SneakerFlash/internal/infra/redis"
	"context"
	"fmt"
)

// syncStockCache 覆盖写入商品库存缓存（秒杀用）
func syncStockCache(productID uint, stock int) {
	ctx := context.Background()
	key := fmt.Sprintf("product:stock:%d", productID)
	_ = redis.RDB.Set(ctx, key, stock, 0).Err()
}

// invalidateProductInfoCache 让详情重新回源数据库
func invalidateProductInfoCache(productID uint) {
	ctx := context.Background()
	key := fmt.Sprintf("product:info:%d", productID)
	redis.RDB.Del(ctx, key)
}
