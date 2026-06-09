package urlcheck

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

var counter int
var counterMu = &sync.Mutex{}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://example.com", false},
		{"http://localhost:8080", false},
		{"example.com", true},
		{"https://exa mple", true},
		{"ftp://example.com", true},
		{"https://", true},
		{" ", true},
		{"https://example.com/path?q=1#frag", false},
	}

	for _, test := range tests {
		err := IsValidURL(test.url)

		if (err != nil) != test.wantErr {
			t.Errorf("IsValidURL() error = %v, wantErr %v", err, test.wantErr)
		}
		if test.wantErr && err != ErrFormat {
			t.Errorf("IsValidURL() error type = %v, want %v", err, ErrFormat)
		}
	}
}

func TestIsURLReachableWithRetry(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc // 自定义服务器处理逻辑
		wantErr bool
	}{
		{
			name: "HEAD 成功返回 200",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodHead {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			},
			wantErr: false,
		},
		{
			name: "HEAD 返回 404（不存在）",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodHead {
					w.WriteHeader(http.StatusNotFound)
				} else {
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			},
			wantErr: true,
		},
		{
			name: "HEAD 返回 500（服务器错误）",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodHead {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			},
			wantErr: true,
		},
		{
			name: "HEAD 返回 500（服务器错误），重试后成功",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodHead {
					// 第一次请求返回 500，第二次及以后返回 200
					// 注意：因为重试会发送多次请求，需要记录调用次数
					// 使用闭包变量
					counterMu.Lock()
					defer counterMu.Unlock()
					counter++
					if counter >= 2 {
						w.WriteHeader(http.StatusOK)
					} else {
						w.WriteHeader(http.StatusInternalServerError)
					}
				} else {
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			},
			wantErr: false,
		},
		{
			name: "HEAD 返回 405，降级 GET 成功",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodHead:
					w.WriteHeader(http.StatusMethodNotAllowed)
				case http.MethodGet:
					w.WriteHeader(http.StatusOK)
				default:
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			},
			wantErr: false,
		},
		{
			name: "HEAD 405 且 GET 也失败",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodHead:
					w.WriteHeader(http.StatusMethodNotAllowed)
				case http.MethodGet:
					w.WriteHeader(http.StatusInternalServerError)
				default:
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			},
			wantErr: true,
		},
		{
			name: "连接超时（模拟延迟）",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(4 * time.Second) // 超过客户端 3 秒超时
				w.WriteHeader(http.StatusOK)
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// 重置计数器（如果使用）
			if test.name == "HEAD 返回 500（服务器错误），重试后成功" {
				counterMu.Lock()
				counter = 0
				counterMu.Unlock()
			}

			server := httptest.NewServer(test.handler)
			defer server.Close()

			// 注意：由于 IsURLReachableWithRetry 内部使用了全局 httpClient（超时 3 秒）
			// 为了精确控制超时，在测试中临时替换 httpClient 变量
			originalClient := httpClient
			if test.name == "连接超时（模拟延迟）" {
				// 保持原样，因为 3s 超时 < 4s 延迟
			}
			defer func() { httpClient = originalClient }() // 恢复

			err := IsURLReachableWithRetry(server.URL)

			if (err != nil) != test.wantErr {
				t.Errorf("IsURLReachableWithRetry() error = %v, wantErr %v", err, test.wantErr)
			}
			if test.wantErr && err != ErrNetWork {
				t.Errorf("IsValidURL() error type = %v, want %v", err, ErrNetWork)
			}
		})
	}
}
