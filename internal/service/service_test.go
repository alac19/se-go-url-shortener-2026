package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	model "github.com/alac19/se-go-url-shortener-2026/internal/model"
	urlcheck "github.com/alac19/se-go-url-shortener-2026/pkg/urlcheck"
)

type mockRepo struct {
	// 记录调用参数（可用于断言）
	CalledCreateArg   *model.LinkMap
	CalledUpdateID    uint64
	CalledUpdateCode  string
	CalledFindLinkArg *model.LinkMap
	CalledFindCode    string
	CalledStatsCode   string
	CallDelta         int64

	// 预设返回值
	CreateErr       error
	UpdateErr       error
	FindLinkResult  *model.LinkMap
	FindLinkErr     error
	IncrementErr    error
	IncrementErrMap map[string]error
}

type mockCache struct {
	CalledCtx           context.Context
	CalledSetKey        string
	CalledSetValue      interface{}
	CalledSetExpiration time.Duration
	CalledGetkey        string
	CalledIncrKey       string
	CalledCursor        uint64
	CalledMatch         string
	CalledCount         int64
	CalledDelKey        string

	SetErr           error
	GetResult        string
	GetErr           error
	IncrResult       int64
	IncrErr          error
	ScanResultKeys   []string
	ScanResultCursor uint64
	ScanErr          error
	DelResult        int64
	DelErr           error
	GetResultMap     map[string]string
	GetErrMap        map[string]error
	DelResultMap     map[string]int64
	DelErrMap        map[string]error
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

func (mp *mockRepo) IncrementClickCount(shortCode string, delta int64) error {
	mp.CalledStatsCode = shortCode
	mp.CallDelta = delta

	if err, ok := mp.IncrementErrMap[shortCode]; ok {
		return err
	}

	return mp.IncrementErr
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

	if val, ok := mc.GetResultMap[key]; ok {
		return val, mc.GetErrMap[key]
	}

	return mc.GetResult, mc.GetErr
}

func (mc *mockCache) Incr(ctx context.Context, key string) (int64, error) {
	mc.CalledCtx = ctx
	mc.CalledIncrKey = key

	return mc.IncrResult, mc.IncrErr
}

func (mc *mockCache) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	mc.CalledCtx = ctx
	mc.CalledCursor = cursor
	mc.CalledMatch = match
	mc.CalledCount = count

	return mc.ScanResultKeys, mc.ScanResultCursor, mc.ScanErr
}

func (mc *mockCache) Del(ctx context.Context, key string) (int64, error) {
	mc.CalledCtx = ctx
	mc.CalledDelKey = key

	if err, ok := mc.DelErrMap[key]; ok {
		return mc.DelResult, err
	}

	return mc.DelResult, mc.DelErr
}

func TestCreateShortLink(t *testing.T) {
	// 配置 urlcheck 包，让真实网络请求能够正常进行（https://example.com 总是可达，且测试很快）
	urlcheck.Configure(3, 3, 1)

	tests := []struct {
		name          string
		longURL       string
		mockCreateErr error
		mockUpdateErr error
		wantShortCode string
		wantErr       bool
	}{
		{"正常生成", "https://example.com", nil, nil, "http://localhost:8080/1", false},
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

			s := NewService(mp, mc, "http://localhost:8080/", 0)

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
		mockIncrResult     int64
		mockIncrErr        error
		wantLongURL        string
		wantErr            bool
	}{
		{"缓存命中", "1", &model.LinkMap{ID: 1, LongURL: "https://example.com", ShortCode: "1"}, nil, nil, "https://example.com", nil, 1, nil, "https://example.com", false},
		{"缓存出错降级成功", "1", &model.LinkMap{ID: 1, LongURL: "https://example.com", ShortCode: "1"}, nil, nil, "", errors.New("redis error"), 1, nil, "https://example.com", false},
		{"缓存出错降级成功, 但 Set 失败", "1", &model.LinkMap{ID: 1, LongURL: "https://example.com", ShortCode: "1"}, nil, errors.New("set error"), "", errors.New("redis error"), 1, nil, "https://example.com", false},
		{"缓存出错降级成功, 但 Incr 失败", "1", &model.LinkMap{ID: 1, LongURL: "https://example.com", ShortCode: "1"}, nil, errors.New("set error"), "", errors.New("redis error"), 0, errors.New("incr error"), "https://example.com", false},
		{"缓存出错降级失败", "1", nil, errors.New("find error"), nil, "", errors.New("redis error"), 0, nil, "", true},
		{"缓存为空, 全部成功", "1", &model.LinkMap{ID: 1, LongURL: "https://example.com", ShortCode: "1"}, nil, nil, "https://example.com", redis.Nil, 1, nil, "https://example.com", false},
		{"缓存为空, FindLink 失败", "1", nil, errors.New("find error"), nil, "", redis.Nil, 0, nil, "", true},
		{"缓存为空, Set 失败", "1", &model.LinkMap{ID: 1, LongURL: "https://example.com", ShortCode: "1"}, nil, errors.New("set error"), "", redis.Nil, 1, nil, "https://example.com", false},
		{"缓存为空, Incr 失败", "1", &model.LinkMap{ID: 1, LongURL: "https://example.com", ShortCode: "1"}, nil, nil, "", redis.Nil, 0, errors.New("incr error"), "https://example.com", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mp := &mockRepo{
				FindLinkResult: test.mockFindLinkResult,
				FindLinkErr:    test.mockFindErr,
			}
			mc := &mockCache{
				GetResult:  test.mockGetResult,
				GetErr:     test.mockGetErr,
				SetErr:     test.mockSetErr,
				IncrResult: test.mockIncrResult,
				IncrErr:    test.mockIncrErr,
			}

			s := NewService(mp, mc, "", 0)

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

func TestFlushStats(t *testing.T) {
	tests := []struct {
		name                 string
		mockScanResultKeys   []string
		mockScanResultCursor uint64
		mockScanErr          error
		mockGetResultMap     map[string]string
		mockGetErrMap        map[string]error
		mockIncrementErrMap  map[string]error
		mockDelErrMap        map[string]error
	}{
		{"正常异步写入", []string{"stats:abc", "stats:xyz"}, 0, nil, map[string]string{"stats:abc": "3", "stats:xyz": "5"}, map[string]error{}, map[string]error{}, map[string]error{}},
		{"Scan 失败", []string{}, 0, errors.New("redis error"), map[string]string{}, map[string]error{}, map[string]error{}, map[string]error{}},
		{"Get 失败(redis.Nil)", []string{"stats:abc", "stats:xyz"}, 0, nil, map[string]string{"stats:abc": "", "stats:xyz": "5"}, map[string]error{"stats:abc": redis.Nil, "stats:xyz": nil}, map[string]error{}, map[string]error{}},
		{"Get 失败(非redis.Nil)", []string{"stats:abc", "stats:xyz"}, 0, nil, map[string]string{"stats:abc": "", "stats:xyz": "5"}, map[string]error{"stats:abc": errors.New("redis error"), "stats:xyz": nil}, map[string]error{}, map[string]error{}},
		{"Increment 失败", []string{"stats:abc", "stats:xyz"}, 0, nil, map[string]string{"stats:abc": "3", "stats:xyz": "5"}, map[string]error{}, map[string]error{"abc": errors.New("mysql error"), "xyz": nil}, map[string]error{}},
		{"Del 失败", []string{"stats:abc", "stats:xyz"}, 0, nil, map[string]string{"stats:abc": "3", "stats:xyz": "5"}, map[string]error{}, map[string]error{}, map[string]error{"stats:abc": errors.New("redis error"), "stats:xyz": nil}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mp := &mockRepo{
				IncrementErrMap: test.mockIncrementErrMap,
			}
			mc := &mockCache{
				ScanResultKeys:   test.mockScanResultKeys,
				ScanResultCursor: test.mockScanResultCursor,
				ScanErr:          test.mockScanErr,
				GetResultMap:     test.mockGetResultMap,
				GetErrMap:        test.mockGetErrMap,
				DelErrMap:        test.mockDelErrMap,
			}

			s := NewService(mp, mc, "", 100)

			s.FlushStats()
		})
	}
}
