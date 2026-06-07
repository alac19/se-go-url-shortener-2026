// Package cache provides Redis cache interface.
package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache 定义缓存操作接口，支持设置键值对（带过期时间）和获取值。
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type Redis struct {
	Rdb *redis.Client
}

func (r Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	_, err := r.Rdb.Set(ctx, key, value, expiration).Result()

	return err
}

func (r Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.Rdb.Get(ctx, key).Result()

	return val, err
}
