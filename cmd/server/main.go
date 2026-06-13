package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	// 导入自定义模块包
	config "github.com/alac19/se-go-url-shortener-2026/internal/config"
	handler "github.com/alac19/se-go-url-shortener-2026/internal/handler"
	ratelimit "github.com/alac19/se-go-url-shortener-2026/internal/middleware"
	repository "github.com/alac19/se-go-url-shortener-2026/internal/repository"
	cache "github.com/alac19/se-go-url-shortener-2026/internal/repository/cache"
	service "github.com/alac19/se-go-url-shortener-2026/internal/service"
	limiter "github.com/alac19/se-go-url-shortener-2026/pkg/limiter"
	logger "github.com/alac19/se-go-url-shortener-2026/pkg/logger"
	urlcheck "github.com/alac19/se-go-url-shortener-2026/pkg/urlcheck"
)

var db *gorm.DB

func main() {
	fmt.Println("项目开发阶段启动")

	// 加载配置
	cfg, err := config.LoadConfig("configs/config.toml")

	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}

	if err := logger.Init(cfg.Log.Level, cfg.Log.FilePath); err != nil {
		slog.Error("初始化日志失败", "error", err)
		os.Exit(1)
	}
	slog.Info("日志系统初始化成功", "level", cfg.Log.Level, "file", cfg.Log.FilePath)

	// 连接 MySQL
	dsn := cfg.MySQL.DSN

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		slog.Error("连接数据库失败", "error", err)
		os.Exit(1)
	}

	slog.Info("MySQL 连接成功")

	// 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,     // 启动的 Redis 容器
		Password: cfg.Redis.Password, // 没设密码就留空
		DB:       cfg.Redis.DB,       // 使用默认数据库
	})

	// 测试连接
	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("连接 Redis 失败", "error", err)
		os.Exit(1)
	}

	slog.Info("Redis 连接成功")

	// 初始化 Gin
	r := gin.Default()

	// // 测试路由：GET /ping
	// r.GET("/ping", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{"message": "pong"})
	// })
	// fmt.Println("初始化服务框架已通过测试！")
	// fmt.Println("进行 MVP 开发学习...")

	repo := repository.NewRepository(db)

	cache.Configure(cfg.Cache.TTLSeconds)
	redis := &cache.Redis{Rdb: rdb}
	urlcheck.Configure(cfg.URLCheck.TimeoutSeconds, cfg.URLCheck.MaxRetries, cfg.URLCheck.RetryDelaySeconds)

	service := service.NewService(repo, redis, cfg.Server.Domain, int64(cfg.AsyncFlush.ScanCount))

	// 异步写入
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.AsyncFlush.IntervalSeconds) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			service.FlushStats()
		}
	}()

	// if POST
	lm := limiter.NewLimiterMap(rate.Every(time.Duration(cfg.Ratelimit.EverySeconds)*time.Second), cfg.Ratelimit.Burst)

	md1 := ratelimit.HandleRateLimit(lm)

	hd1 := handler.HandleCreateShortLink(service)

	fmt.Printf("已处理 handler 层 (返回 %v)、service 层、repository 层！\n", hd1)

	r.POST("/api/links", md1, hd1)

	fmt.Println("路由注册成功！")

	// if GET
	hd2 := handler.HandleRedirect(service)

	fmt.Printf("已处理 handler 层 (返回 %v)、service 层、repository 层！\n", hd2)

	r.GET("/:code", hd2)

	fmt.Println("路由注册成功！")

	// 启动服务（端口 8080）
	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		slog.Error("启动服务器失败", "error", err)
		os.Exit(1)
	}
}
