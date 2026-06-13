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

// Redis 实现 Cache 接口，封装了 go-redis 客户端。
type Redis struct {
	Rdb *redis.Client
}

// Set 存储键值对到 Redis。如果 expiration == 0，则使用默认 TTL（通过 Configure 设置）。
func (r Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if expiration == 0 {
		_, err := r.Rdb.Set(ctx, key, value, defaultTTL).Result()

		return err
	} else {
		_, err := r.Rdb.Set(ctx, key, value, expiration).Result()

		return err
	}
}

// Get 从 Redis 中获取指定 key 的值。
func (r Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.Rdb.Get(ctx, key).Result()

	return val, err
}

// Incr 将指定 key 的整数值原子性自增 1，返回自增后的结果。
func (r Redis) Incr(ctx context.Context, key string) (int64, error) {
	count, err := r.Rdb.Incr(ctx, key).Result()

	return count, err
}

// Scan 遍历匹配 pattern 的 key，使用游标分批返回，避免阻塞 Redis。
func (r Redis) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return r.Rdb.Scan(ctx, cursor, match, count).Result()
}

// Del 从 Redis 中删除指定的 key，返回删除的 key 数量。
func (r Redis) Del(ctx context.Context, key string) (int64, error) {
	return r.Rdb.Del(ctx, key).Result()
}

// Configure 设置默认的缓存过期时间（秒）。
func Configure(ttl int) {
	defaultTTL = time.Duration(ttl) * time.Second
}
