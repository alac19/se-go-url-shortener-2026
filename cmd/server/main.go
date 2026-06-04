package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

	// 测试路由：GET /ping
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 启动服务（端口 8080）
	r.Run(":8080")
}
