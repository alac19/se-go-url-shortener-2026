// Package repository handles database operations.
package repository

import (
	"gorm.io/gorm"

	model "github.com/alac19/se-go-url-shortener-2026/internal/model"
)

// LinkRepository 定义数据访问层接口，用于操作短链接映射表。
type LinkRepository interface {
	Create(lm *model.LinkMap) error
	UpdateShortCode(id uint64, shortCode string) error
	FindLink(lm *model.LinkMap, shortCode string) error
	IncrementClickCount(shortCode string, clickCount int64) error
}

// Repository 基于 GORM 实现 LinkRepository 接口，提供 MySQL 数据库操作。
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建一个 Repository 实例，封装数据库连接。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 插入一条新的短链接记录（short_code 字段暂时为空）。
func (r Repository) Create(lm *model.LinkMap) error {
	return r.db.Create(lm).Error
}

// UpdateShortCode 根据 id 更新短码字段。
func (r Repository) UpdateShortCode(id uint64, shortCode string) error {
	return r.db.Model(&model.LinkMap{}).Where("id = ?", id).Update("short_code", shortCode).Error
}

// FindLink 根据短码查询记录，结果填充到 lm 中。
func (r Repository) FindLink(lm *model.LinkMap, shortCode string) error {
	return r.db.Where("short_code = ?", shortCode).First(lm).Error
}

// IncrementClickCount 为指定短码的链接增加点击次数（原子操作）。
func (r Repository) IncrementClickCount(shortCode string, delta int64) error {
	return r.db.Model(&model.LinkMap{}).Where("short_code = ?", shortCode).Update("click_count", gorm.Expr("click_count + ?", delta)).Error
}
