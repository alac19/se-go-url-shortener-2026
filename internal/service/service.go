// Package service implements business logic of shortlink.
package service

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	model "github.com/alac19/se-go-url-shortener-2026/internal/model"
	repository "github.com/alac19/se-go-url-shortener-2026/internal/repository"
	cache "github.com/alac19/se-go-url-shortener-2026/internal/repository/cache"
	base62 "github.com/alac19/se-go-url-shortener-2026/pkg"
)

type Service struct {
	repo  *repository.Repository
	redis *cache.Redis
}

// NewService 创建一个 Service 实例，注入 repository（数据库操作）和 cache（Redis 缓存）依赖。
func NewService(repo *repository.Repository, redis *cache.Redis) Service {
	return Service{repo: repo, redis: redis}
}

// CreateShortLink 根据原始长链接生成短码。
// 先插入数据库获得自增 ID → 转换为 base62 短码 → 更新记录中的短码字段。
// 返回短码（不含域名），若出错返回错误。
func (s Service) CreateShortLink(longURL string) (string, error) {
	lm := &model.LinkMap{LongURL: longURL}

	if err := s.repo.Create(lm); err != nil {
		return "", err
	}

	shortCode := base62.IntToBase62(lm.ID)

	// 更新数据库
	if err := s.repo.UpdateShortCode(lm.ID, shortCode); err != nil {
		return "", err
	}

	return shortCode, nil
}

// Redirect 根据短码查询原始长链接。
// 优先从 Redis 缓存读取，命中则直接返回；若缓存未命中或 Redis 故障则降级查询数据库，
// 查询成功后回写缓存（设置 1 小时过期）。若短码不存在或数据库查询失败，返回错误。
func (s Service) Redirect(shortCode string) (string, error) {
	lm := &model.LinkMap{}
	ctx := context.Background()
	cacheKey := "shortlink:" + shortCode

	// 查 Redis
	val, err := s.redis.Get(ctx, cacheKey)

	// 缓存命中
	if err == nil {
		return val, nil
	}

	// Redis 出错，降级
	if err != redis.Nil {
		log.Printf("Redis error: %v", err)
	}

	// 查数据库
	if err := s.repo.FindLink(lm, shortCode); err != nil {
		return "", err
	}

	if err := s.redis.Set(ctx, cacheKey, lm.LongURL, time.Hour); err != nil {
		log.Printf("Redis error: %v", err)
	}

	statsKey := "stats:" + shortCode

	if _, err := s.redis.Incr(ctx, statsKey); err != nil {
		log.Printf("Redis error: %v", err)
	}

	return lm.LongURL, nil
}
