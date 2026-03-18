package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	redisinfra "SneakerFlash/internal/infra/redis"
	goredis "github.com/redis/go-redis/v9"
)

const retryCounterTTL = 24 * time.Hour

type retryStore interface {
	Incr(ctx context.Context, key string) (int, error)
	Delete(ctx context.Context, key string) error
}

func newRetryStore() retryStore {
	memory := newMemoryRetryStore()
	if redisinfra.RDB == nil {
		return memory
	}
	return &redisRetryStore{
		client:   redisinfra.RDB,
		fallback: memory,
	}
}

type memoryRetryStore struct {
	mu     sync.Mutex
	counts map[string]int
}

func newMemoryRetryStore() *memoryRetryStore {
	return &memoryRetryStore{
		counts: make(map[string]int),
	}
}

func (s *memoryRetryStore) Incr(_ context.Context, key string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counts[key]++
	return s.counts[key], nil
}

func (s *memoryRetryStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.counts, key)
	return nil
}

func (s *memoryRetryStore) Count(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.counts[key]
}

type redisRetryStore struct {
	client   redisStringCounter
	fallback retryStore
}

type redisStringCounter interface {
	Incr(context.Context, string) *goredis.IntCmd
	Expire(context.Context, string, time.Duration) *goredis.BoolCmd
	Del(context.Context, ...string) *goredis.IntCmd
}

func (s *redisRetryStore) Incr(ctx context.Context, key string) (int, error) {
	count64, err := s.client.Incr(ctx, retryCounterKey(key)).Result()
	if err != nil {
		return s.fallback.Incr(ctx, key)
	}
	if err := s.client.Expire(ctx, retryCounterKey(key), retryCounterTTL).Err(); err != nil {
		return int(count64), fmt.Errorf("set retry ttl: %w", err)
	}
	return int(count64), nil
}

func (s *redisRetryStore) Delete(ctx context.Context, key string) error {
	if _, err := s.client.Del(ctx, retryCounterKey(key)).Result(); err != nil {
		return s.fallback.Delete(ctx, key)
	}
	return s.fallback.Delete(ctx, key)
}

func retryCounterKey(key string) string {
	return "kafka:consume:retry:" + key
}
