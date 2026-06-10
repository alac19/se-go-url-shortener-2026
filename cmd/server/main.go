package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	// 导入自定义模块包
	handler "github.com/alac19/se-go-url-shortener-2026/internal/handler"
	ratelimit "github.com/alac19/se-go-url-shortener-2026/internal/middleware"
	repository "github.com/alac19/se-go-url-shortener-2026/internal/repository"
	cache "github.com/alac19/se-go-url-shortener-2026/internal/repository/cache"
	service "github.com/alac19/se-go-url-shortener-2026/internal/service"
	limiter "github.com/alac19/se-go-url-shortener-2026/pkg/limiter"
)

var db *gorm.DB

func main() {
	fmt.Println("项目开发阶段启动！")

	// 连接 MySQL
	dsn := "root:Alac197@@tcp(127.0.0.1:3306)/shortlink_db?charset=utf8mb4&parseTime=True&loc=Local"
	var err error

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	fmt.Println("MySQL 连接成功")

	// 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // 启动的 Redis 容器
		Password: "",               // 没设密码就留空
		DB:       0,                // 使用默认数据库
	})

	// 测试连接
	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("连接 Redis 失败: %v", err)
	}

	fmt.Println("Redis 连接成功")

	// 初始化 Gin
	r := gin.Default()

	// // 测试路由：GET /ping
	// r.GET("/ping", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{"message": "pong"})
	// })
	// fmt.Println("初始化服务框架已通过测试！")
	// fmt.Println("进行 MVP 开发学习...")

	repo := repository.NewRepository(db)

	redis := &cache.Redis{Rdb: rdb}

	service := service.NewService(repo, redis)

	// if POST
	lm := limiter.NewLimiterMap(rate.Every(12*time.Second), 5)

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
	r.Run(":8080")
}
