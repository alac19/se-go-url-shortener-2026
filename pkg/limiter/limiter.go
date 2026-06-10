// Package limiter provides IP-based rate limiting using the token bucket algorithm.
package limiter

import (
	"sync"

	"golang.org/x/time/rate"
)

// Limiter 定义了限流器的行为接口。
type Limiter interface {
	Allow(ip string) bool
}

// LimiterMap 管理一组按 IP 地址索引的限流器。
// 使用读写锁保证并发安全，为每个 IP 创建独立的令牌桶。
type LimiterMap struct {
	speed    rate.Limit
	bucket   int
	limiters map[string]*rate.Limiter
	mux      sync.RWMutex
}

// NewLimiterMap 创建一个新的限流器管理器。
// 参数 speed 指定每秒生成的令牌数, 参数 bucket 指定令牌桶的最大容量（即允许的突发请求数）。
func NewLimiterMap(speed rate.Limit, bucket int) *LimiterMap {
	return &LimiterMap{speed: speed, bucket: bucket, limiters: make(map[string]*rate.Limiter)}
}

// Allow 判断指定 IP 的请求是否应该被限流（拒绝）。
// 返回 true 表示请求超过速率限制，应被拒绝；返回 false 表示允许通过。
// 如果该 IP 尚未创建限流器，则自动创建一个新的令牌桶实例。
func (lm *LimiterMap) Allow(ip string) bool {
	lm.mux.RLock()
	l, ok := lm.limiters[ip]
	lm.mux.RUnlock()

	if ok {
		return !l.Allow()
	}

	lm.mux.Lock()
	defer lm.mux.Unlock()
	l, ok = lm.limiters[ip]

	if !ok {
		l = rate.NewLimiter(lm.speed, lm.bucket)
		lm.limiters[ip] = l
	}

	return !l.Allow()
}
