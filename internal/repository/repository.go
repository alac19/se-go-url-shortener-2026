package repository

import (
	"gorm.io/gorm"

	model "github.com/alac19/se-go-url-shortener-2026/internal/model"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r Repository) Create(lm *model.LinkMap) error {
	// 新增数据插入数据库...
	if err := r.db.Create(lm).Error; err != nil {
		return err
	}

	return nil
}

func (r Repository) UpdateShortCode(id uint64, shortCode string) error {
	// 更新短码字段...
	if err := r.db.Model(&model.LinkMap{}).Where("id = ?", id).Update("short_code", shortCode).Error; err != nil {
		return err
	}

	return nil
}
