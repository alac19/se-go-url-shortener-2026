package service

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Service struct {
	db  *gorm.DB
	rdb *redis.Client
}

// func Do(operation string) string {
// 	switch {
// 	case operation == "GET":
// 		println("Service 层处理 GET 业务逻辑")
// 		return "GET"
// 	case operation == "POST":
// 		println("Service 层处理 POST 业务逻辑")
// 		return "POST"
// 	}

// 	return ""
// }

func NewService(db *gorm.DB, rdb *redis.Client) Service {
	return Service{db: db, rdb: rdb}
}

func (s Service) CreatShortLink(longURL string) string {
	// 生成短链接业务逻辑...

	return "abc123"
}

func (s Service) Redirect(shortCode string) string {
	// 重定向业务逻辑...

	return "http://example.com"
}
