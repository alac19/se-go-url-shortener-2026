// Package repository handles database operations.
package repository

import (
	"gorm.io/gorm"

	model "github.com/alac19/se-go-url-shortener-2026/internal/model"
)

type LinkRepository interface {
	Create(lm *model.LinkMap) error
	UpdateShortCode(id uint64, shortCode string) error
	FindLink(lm *model.LinkMap, shortCode string) error
}

type Repository struct {
	db *gorm.DB
}

// NewRepository 创建一个 Repository 实例，封装数据库连接。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r Repository) Create(lm *model.LinkMap) error {
	return r.db.Create(lm).Error
}

func (r Repository) UpdateShortCode(id uint64, shortCode string) error {
	return r.db.Model(&model.LinkMap{}).Where("id = ?", id).Update("short_code", shortCode).Error
}

func (r Repository) FindLink(lm *model.LinkMap, shortCode string) error {
	return r.db.Where("short_code = ?", shortCode).First(lm).Error
}
