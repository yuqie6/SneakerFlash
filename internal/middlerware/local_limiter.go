package middlerware

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// LocalLimiter 本地内存限流器，用于在 Redis 之前挡掉大部分超限请求。
// 采用分片 map 减少锁竞争，支持自动过期清理。
type LocalLimiter struct {
	shards    []*limiterShard
	shardMask uint64
	rate      rate.Limit
	burst     int
	ttl       time.Duration
}

type limiterShard struct {
	mu       sync.RWMutex
	limiters map[string]*limiterEntry
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

const (
	defaultShardCount = 64
	defaultTTL        = 10 * time.Minute
	cleanupInterval   = 1 * time.Minute
)

// NewLocalLimiter 创建本地限流器
// r: 每秒令牌数, burst: 桶容量
func NewLocalLimiter(r int, burst int) *LocalLimiter {
	ll := &LocalLimiter{
		shards:    make([]*limiterShard, defaultShardCount),
		shardMask: defaultShardCount - 1,
		rate:      rate.Limit(r),
		burst:     burst,
		ttl:       defaultTTL,
	}

	for i := range ll.shards {
		ll.shards[i] = &limiterShard{
			limiters: make(map[string]*limiterEntry),
		}
	}

	// 启动清理协程
	go ll.cleanup()

	return ll
}

// Allow 检查是否允许通过，非阻塞
func (ll *LocalLimiter) Allow(key string) bool {
	shard := ll.getShard(key)

	shard.mu.RLock()
	entry, exists := shard.limiters[key]
	shard.mu.RUnlock()

	if exists {
		entry.lastSeen = time.Now()
		return entry.limiter.Allow()
	}

	// 不存在则创建
	shard.mu.Lock()
	// 双重检查
	if entry, exists = shard.limiters[key]; exists {
		shard.mu.Unlock()
		entry.lastSeen = time.Now()
		return entry.limiter.Allow()
	}

	limiter := rate.NewLimiter(ll.rate, ll.burst)
	shard.limiters[key] = &limiterEntry{
		limiter:  limiter,
		lastSeen: time.Now(),
	}
	shard.mu.Unlock()

	return limiter.Allow()
}

// getShard 根据 key 哈希获取分片
func (ll *LocalLimiter) getShard(key string) *limiterShard {
	h := fnv64a(key)
	return ll.shards[h&ll.shardMask]
}

// fnv64a 快速哈希
func fnv64a(s string) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	h := uint64(offset64)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime64
	}
	return h
}

// cleanup 定期清理过期的限流器
func (ll *LocalLimiter) cleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		for _, shard := range ll.shards {
			shard.mu.Lock()
			for key, entry := range shard.limiters {
				if now.Sub(entry.lastSeen) > ll.ttl {
					delete(shard.limiters, key)
				}
			}
			shard.mu.Unlock()
		}
	}
}

// ========== 全局限流器实例管理 ==========

var (
	globalLimiters sync.Map // prefix -> *LocalLimiter
)

// GetOrCreateLimiter 获取或创建指定前缀的本地限流器
func GetOrCreateLimiter(prefix string, r, burst int) *LocalLimiter {
	if v, ok := globalLimiters.Load(prefix); ok {
		return v.(*LocalLimiter)
	}

	limiter := NewLocalLimiter(r, burst)
	actual, _ := globalLimiters.LoadOrStore(prefix, limiter)
	return actual.(*LocalLimiter)
}
