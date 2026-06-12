// Package ratelimit provides HTTP flow limiting middleware based on the Gin framework.
package ratelimit

import (
	"net/http"

	"github.com/gin-gonic/gin"

	limiter "github.com/alac19/se-go-url-shortener-2026/pkg/limiter"
)

// HandleRateLimit 返回一个 Gin 中间件处理函数，用于对请求进行限流。
// 参数 lm 是实现了 limiter.Limiter 接口的限流器实例。
// 如果 lm.Allow(ip) 返回 true（表示应限流），则中间件返回 429 状态码并终止请求；
// 否则调用 ctx.Next() 继续处理后续的 HTTP 处理函数。
func HandleRateLimit(lm limiter.Limiter) gin.HandlerFunc {
	println("\n进行限流!")

	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()

		if lm.Allow(ip) {
			ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})

			ctx.Abort()

			return
		}

		ctx.Next()
	}
}
