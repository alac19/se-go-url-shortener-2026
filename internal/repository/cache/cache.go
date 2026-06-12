// Package cache provides Redis cache interface.
package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var defaultTTL time.Duration

// Cache 定义缓存操作接口，支持设置键值对（带过期时间）和获取值。
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Incr(ctx context.Context, key string) (int64, error)
	Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error)
	Del(ctx context.Context, key string) (int64, error)
}

type Redis struct {
	Rdb *redis.Client
}

func (r Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if expiration == 0 {
		_, err := r.Rdb.Set(ctx, key, value, defaultTTL).Result()

		return err
	} else {
		_, err := r.Rdb.Set(ctx, key, value, expiration).Result()

		return err
	}
}

func (r Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.Rdb.Get(ctx, key).Result()

	return val, err
}

func (r Redis) Incr(ctx context.Context, key string) (int64, error) {
	count, err := r.Rdb.Incr(ctx, key).Result()

	return count, err
}

func (r Redis) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return r.Rdb.Scan(ctx, cursor, match, count).Result()
}

func (r Redis) Del(ctx context.Context, key string) (int64, error) {
	return r.Rdb.Del(ctx, key).Result()
}

func Configure(ttl int) {
	defaultTTL = time.Duration(ttl) * time.Second
}
