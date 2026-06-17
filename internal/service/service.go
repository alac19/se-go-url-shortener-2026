// Package service implements business logic of shortlink.
package service

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"

	model "github.com/alac19/se-go-url-shortener-2026/internal/model"
	repository "github.com/alac19/se-go-url-shortener-2026/internal/repository"
	cache "github.com/alac19/se-go-url-shortener-2026/internal/repository/cache"
	base62 "github.com/alac19/se-go-url-shortener-2026/pkg/base62"
	urlcheck "github.com/alac19/se-go-url-shortener-2026/pkg/urlcheck"
)

var (
	// ErrInValidURL 表示长链接格式无效。
	ErrInValidURL = errors.New("invalid url format")

	// ErrURLNotReachable 表示长链接经过重试后仍不可访问。
	ErrURLNotReachable = errors.New("url not reachable after retry")
)

// Service 是短链接服务的核心业务逻辑层，负责短链生成、重定向、统计计数及异步写入。
// 它依赖 repository（数据库操作）和 cache（Redis 缓存），并通过 domain 拼接完整短链接，
// scanCount 用于控制 SCAN 命令每次返回的 key 数量。
type Service struct {
	repo      repository.LinkRepository
	cache     cache.Cache
	domain    string
	scanCount int64
}

// NewService 创建一个 Service 实例，注入 repository（数据库操作）和 cache（Redis 缓存）依赖。
func NewService(repo repository.LinkRepository, cache cache.Cache, domain string, scanCount int64) Service {
	return Service{repo: repo, cache: cache, domain: domain, scanCount: scanCount}
}

// CreateShortLink 根据原始长链接生成短码。
// 先插入数据库获得自增 ID → 转换为 base62 短码 → 更新记录中的短码字段。
// 返回短码（不含域名），若出错返回错误。
func (s Service) CreateShortLink(longURL string) (string, error) {
	if err := urlcheck.IsValidURL(longURL); err != nil {
		return "", ErrInValidURL
	}

	if err := urlcheck.IsURLReachableWithRetry(longURL); err != nil {
		slog.Error("Network error", "error", err)
		return "", ErrURLNotReachable
	}

	lm := &model.LinkMap{LongURL: longURL}

	if err := s.repo.Create(lm); err != nil {
		return "", err
	}

	shortCode := base62.IntToBase62(lm.ID)

	// 更新数据库
	if err := s.repo.UpdateShortCode(lm.ID, shortCode); err != nil {
		return "", err
	}

	return s.domain + shortCode, nil
}

// Redirect 根据短码查询原始长链接。
// 优先从 Redis 缓存读取，命中则直接返回；若缓存未命中或 Redis 故障则降级查询数据库，
// 查询成功后回写缓存（设置 1 小时过期）。若短码不存在或数据库查询失败，返回错误。
func (s Service) Redirect(shortCode string) (string, error) {
	lm := &model.LinkMap{}
	var longURL string
	ctx := context.Background()
	cacheKey := "shortlink:" + shortCode

	// 查 Redis
	val, err := s.cache.Get(ctx, cacheKey)

	// 缓存命中
	if err == nil {
		longURL = val
	} else {
		// Redis 出错，降级
		if err != redis.Nil {
			slog.Error("Redis error", "error", err)
		}

		// 查数据库
		if err := s.repo.FindLink(lm, shortCode); err != nil {
			return "", err
		}

		longURL = lm.LongURL

		if err := s.cache.Set(ctx, cacheKey, lm.LongURL, 0); err != nil {
			slog.Error("Redis error", "error", err)
		}
	}

	statsKey := "stats:" + shortCode

	if _, err := s.cache.Incr(ctx, statsKey); err != nil {
		slog.Error("Redis error", "error", err)
	}

	return longURL, nil
}

// FlushStats 将 Redis 中暂存的点击统计计数批量写入 MySQL。
// 该方法使用 SCAN 命令安全地遍历所有以 "stats:" 为前缀的键，
// 对每个键获取计数值并累加到对应短链的 click_count 字段中，成功写入后删除该 Redis 键。
// 若某键处理失败（如 MySQL 更新错误），则保留该键，等待下一次扫描重试。
func (s Service) FlushStats() {
	ctx := context.Background()
	var cursor uint64

	for {
		var keys []string
		var err error

		keys, cursor, err = s.cache.Scan(ctx, cursor, "stats:*", s.scanCount)

		if err != nil {
			slog.Error("SCAN error", "error", err)
			break
		}
		for _, key := range keys {
			code := strings.TrimPrefix(key, "stats:")

			countStr, err := s.cache.Get(ctx, key)

			if err != nil {
				if err != redis.Nil {
					slog.Error("Get error", "key", key, "error", err)
				}

				continue
			}

			count, err := strconv.ParseInt(countStr, 10, 64)

			if err != nil {
				slog.Error("ParseInt error", "value", countStr, "error", err)
				continue
			}

			if count <= 0 {
				continue
			}

			if err := s.repo.IncrementClickCount(code, count); err != nil {
				continue
			}

			if _, err = s.cache.Del(ctx, key); err != nil {
				slog.Error("Del error", "key", key, "error", err)
				continue
			}
		}
		if cursor == 0 {
			break
		}
	}
}
