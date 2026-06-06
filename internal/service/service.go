package service

import (
	model "github.com/alac19/se-go-url-shortener-2026/internal/model"
	repository "github.com/alac19/se-go-url-shortener-2026/internal/repository"
	base62 "github.com/alac19/se-go-url-shortener-2026/pkg"
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) Service {
	return Service{repo: repo}
}

func (s Service) CreateShortLink(longURL string) (string, error) {
	// 获取 id
	// 调用 repository 层
	lm := &model.LinkMap{LongURL: longURL}

	if err := s.repo.Create(lm); err != nil {
		return "", err
	}

	// 掉用 base62 包算出 id 编码
	shortCode := base62.IntToBase62(lm.ID)

	// 更新数据库
	if err := s.repo.UpdateShortCode(lm.ID, shortCode); err != nil {
		return "", err
	}

	return shortCode, nil
}

func (s Service) Redirect(shortCode string) string {
	// 重定向业务逻辑...

	return "http://example.com"
}
