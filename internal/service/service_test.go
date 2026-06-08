package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	model "github.com/alac19/se-go-url-shortener-2026/internal/model"
)

type mockRepo struct {
	// 记录调用参数（可用于断言）
	CalledCreateArg   *model.LinkMap
	CalledUpdateID    uint64
	CalledUpdateCode  string
	CalledFindLinkArg *model.LinkMap
	CalledFindCode    string

	// 预设返回值
	CreateErr      error
	UpdateErr      error
	FindLinkResult *model.LinkMap
	FindLinkErr    error
}

type mockCache struct {
	CalledCtx           context.Context
	CalledSetKey        string
	CalledSetValue      interface{}
	CalledSetExpiration time.Duration
	CalledGetkey        string
	CalledIncrKey       string

	SetErr     error
	GetResult  string
	GetErr     error
	IncrResult int64
	IncrErr    error
}

func (mp *mockRepo) Create(lm *model.LinkMap) error {
	mp.CalledCreateArg = lm
	lm.ID = 1

	return mp.CreateErr // 返回预设的错误
}

func (mp *mockRepo) UpdateShortCode(id uint64, shortCode string) error {
	mp.CalledUpdateID = id
	mp.CalledUpdateCode = shortCode

	return mp.UpdateErr
}

func (mp *mockRepo) FindLink(lm *model.LinkMap, shortCode string) error {
	mp.CalledFindLinkArg = lm
	mp.CalledFindCode = shortCode

	if mp.FindLinkResult != nil {
		*lm = *mp.FindLinkResult
	}

	return mp.FindLinkErr
}

func (mc *mockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	mc.CalledCtx = ctx
	mc.CalledSetKey = key
	mc.CalledSetValue = value
	mc.CalledSetExpiration = expiration

	return mc.SetErr
}

func (mc *mockCache) Get(ctx context.Context, key string) (string, error) {
	mc.CalledCtx = ctx
	mc.CalledGetkey = key

	return mc.GetResult, mc.GetErr
}

func (mc *mockCache) Incr(ctx context.Context, key string) (int64, error) {
	mc.CalledCtx = ctx
	mc.CalledIncrKey = key

	return mc.IncrResult, mc.IncrErr
}

func TestCreateShortLink(t *testing.T) {
	tests := []struct {
		name          string
		longURL       string
		mockCreateErr error
		mockUpdateErr error
		wantShortCode string
		wantErr       bool
	}{
		{"正常生成", "https://example.com", nil, nil, "1", false},
		{"Create失败", "https://example.com", errors.New("db error"), nil, "", true},
		{"Update失败", "https://example.com", nil, errors.New("update error"), "", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mp := &mockRepo{
				CreateErr: test.mockCreateErr,
				UpdateErr: test.mockUpdateErr,
			}
			mc := &mockCache{}

			s := NewService(mp, mc)

			got, err := s.CreateShortLink(test.longURL)

			if (err != nil) != test.wantErr {
				t.Errorf("error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if got != test.wantShortCode {
				t.Errorf("CreateShortLink() = %v, want %v", got, test.wantShortCode)
			}
		})
	}
}

func TestRedirect(t *testing.T) {
	tests := []struct {
		name               string
		shortCode          string
		mockFindLinkResult *model.LinkMap
		mockFindErr        error
		mockSetErr         error
		mockGetResult      string
		mockGetErr         error
		mockIncrErr        error
		wantLongURL        string
		wantErr            bool
	}{
		{"缓存命中", "1", &model.LinkMap{ID: 1, LongURL: "https://example.com", ShortCode: "1"}, nil, nil, "https://example.com", nil, nil, "https://example.com", false},
		{"缓存出错降级成功", "1", &model.LinkMap{ID: 1, LongURL: "https://example.com", ShortCode: "1"}, nil, nil, "", errors.New("redis error"), nil, "https://example.com", false},
		{"缓存出错降级成功, 但 Set 失败", "1", &model.LinkMap{ID: 1, LongURL: "https://example.com", ShortCode: "1"}, nil, errors.New("set error"), "", errors.New("redis error"), nil, "https://example.com", false},
		{"缓存出错降级成功, 但 Incr 失败", "1", &model.LinkMap{ID: 1, LongURL: "https://example.com", ShortCode: "1"}, nil, errors.New("set error"), "", errors.New("redis error"), errors.New("incr error"), "https://example.com", false},
		{"缓存出错降级失败", "1", nil, errors.New("find error"), nil, "", errors.New("redis error"), nil, "", true},
		{"缓存为空, 全部成功", "1", &model.LinkMap{ID: 1, LongURL: "https://example.com", ShortCode: "1"}, nil, nil, "https://example.com", redis.Nil, nil, "https://example.com", false},
		{"缓存为空, FindLink 失败", "1", nil, errors.New("find error"), nil, "", redis.Nil, nil, "", true},
		{"缓存为空, Set 失败", "1", &model.LinkMap{ID: 1, LongURL: "https://example.com", ShortCode: "1"}, nil, errors.New("set error"), "", redis.Nil, nil, "https://example.com", false},
		{"缓存为空, Incr 失败", "1", &model.LinkMap{ID: 1, LongURL: "https://example.com", ShortCode: "1"}, nil, nil, "", redis.Nil, errors.New("incr error"), "https://example.com", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mp := &mockRepo{
				FindLinkResult: test.mockFindLinkResult,
				FindLinkErr:    test.mockFindErr,
			}
			mc := &mockCache{
				GetResult: test.mockGetResult,
				GetErr:    test.mockGetErr,
				SetErr:    test.mockSetErr,
				IncrErr:   test.mockIncrErr,
			}

			s := NewService(mp, mc)

			got, err := s.Redirect(test.shortCode)

			if (err != nil) != test.wantErr {
				t.Errorf("error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if got != test.wantLongURL {
				t.Errorf("Redirect() = %v, want %v", got, test.wantLongURL)
			}
		})
	}
}
