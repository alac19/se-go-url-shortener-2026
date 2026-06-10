package ratelimit

import (
	"github.com/gin-gonic/gin"

	limiter "github.com/alac19/se-go-url-shortener-2026/pkg/limiter"
)

func HandleRateLimit(lm *limiter.LimiterMap) gin.HandlerFunc {
	println("\n进行限流!")

	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()

		if lm.Allow(ip) {
			ctx.JSON(429, gin.H{"error": "too many requests"})

			ctx.Abort()

			return
		}

		ctx.Next()
	}
}
