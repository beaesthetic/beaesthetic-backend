package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Cache struct {
	redis redis.UniversalClient
	log   *zap.Logger
}

func New(redis redis.UniversalClient, log *zap.Logger) *Cache {
	if log == nil {
		log = zap.NewNop()
	}
	return &Cache{redis: redis, log: log}
}

func GetOrLoad[T any](ctx context.Context, cache *Cache, cacheName string, ttl time.Duration, keys []any, loader func() (T, error)) (T, error) {
	if cache == nil {
		return loader()
	}
	cacheKey := buildKey(cacheName, keys...)
	if value, ok := get[T](ctx, cache, cacheKey, ttl); ok {
		return value, nil
	}
	value, err := loader()
	if err != nil {
		return value, err
	}
	set(ctx, cache, cacheKey, value, ttl)
	return value, nil
}

func LoadAndInvalidate[T any](ctx context.Context, cache *Cache, cacheName string, keys []any, loader func() (T, error)) (T, error) {
	value, err := loader()
	if err != nil {
		return value, err
	}
	if cache != nil {
		cache.InvalidateKey(ctx, cacheName, keys...)
	}
	return value, nil
}

func (c *Cache) InvalidateKey(ctx context.Context, cacheName string, keys ...any) {
	cacheKey := buildKey(cacheName, keys...)
	if err := c.redis.Del(ctx, cacheKey).Err(); err != nil {
		c.log.Warn("failed to invalidate customer cache", zap.Error(err), zap.String("key", cacheKey))
	}
}

func get[T any](ctx context.Context, cache *Cache, key string, ttl time.Duration) (T, bool) {
	var zero T
	value, err := cache.redis.GetEx(ctx, key, ttl).Result()
	if err != nil {
		if err != redis.Nil {
			cache.log.Warn("failed to read customer cache", zap.Error(err), zap.String("key", key))
		}
		return zero, false
	}
	if err := json.Unmarshal([]byte(value), &zero); err != nil {
		cache.log.Warn("failed to decode customer cache", zap.Error(err), zap.String("key", key))
		return zero, false
	}
	return zero, true
}

func set(ctx context.Context, cache *Cache, key string, value any, ttl time.Duration) {
	bytes, err := json.Marshal(value)
	if err != nil {
		cache.log.Warn("failed to encode customer cache", zap.Error(err), zap.String("key", key))
		return
	}
	if err := cache.redis.Set(ctx, key, string(bytes), ttl).Err(); err != nil {
		cache.log.Warn("failed to write customer cache", zap.Error(err), zap.String("key", key))
	}
}

func buildKey(cacheName string, keys ...any) string {
	parts := []string{cacheName}
	for _, key := range keys {
		if key == nil {
			continue
		}
		value := strings.TrimSpace(fmt.Sprint(key))
		if value == "" {
			continue
		}
		parts = append(parts, value)
	}
	return strings.Join(parts, ":")
}
