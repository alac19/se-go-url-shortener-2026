// Package urlcheck provides utilities for validating URL format and checking URL reachability with retry logic.
package urlcheck

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrFormat = errors.New("invalid url format")
var ErrNetWork = errors.New("url not reachable")

var httpClient = &http.Client{}

var TimeoutSeconds time.Duration
var MaxRetries int
var RetryDelaySeconds time.Duration

// IsValidURL 检查 URL 格式是否合法。
// 要求协议为 http 或 https，且主机名不为空。返回 nil 表示合法，否则返回 ErrFormat。
func IsValidURL(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)

	gotURL, err := url.Parse(rawURL)

	if err != nil {
		return ErrFormat

	}

	if scheme := strings.ToLower(gotURL.Scheme); scheme != "http" && scheme != "https" {
		return ErrFormat
	}

	if gotURL.Host == "" {
		return ErrFormat
	}

	return nil
}

// IsURLReachableWithRetry 检查 URL 是否可达，支持自动重试（最多 3 次）和 HEAD 降级到 GET。
// 成功返回 nil，失败返回 ErrNetWork。
func IsURLReachableWithRetry(rawURL string) error {
	maxRetries := MaxRetries

loop:
	for i := 0; i < maxRetries; i++ {
		resp, err := httpClient.Head(rawURL)

		if err == nil {
			// 关闭 Body
			resp.Body.Close()

			// 某些服务器不支持 HEAD 但支持 GET，会返回 405 状态码，可以降级使用 GET 并立即关闭响应体
			if resp.StatusCode == 405 {
				for j := 0; j < maxRetries; j++ {
					resp, err := httpClient.Get(rawURL)

					if err == nil {
						// 关闭 Body
						resp.Body.Close()

						if resp.StatusCode >= 200 && resp.StatusCode < 400 {
							return nil
						}
					}

					if j < maxRetries-1 {
						time.Sleep(RetryDelaySeconds)
					}
				}

				break loop
			}
			if resp.StatusCode >= 200 && resp.StatusCode < 400 {
				return nil
			}
		}

		// 最后一次重试失败后不再等待
		if i < maxRetries-1 {
			time.Sleep(1 * time.Second)
		}
	}

	return ErrNetWork
}

func Configure(timeoutSec int, maxRetries int, retryDelaySec int) {
	TimeoutSeconds = time.Duration(timeoutSec) * time.Second
	httpClient = &http.Client{Timeout: TimeoutSeconds}
	MaxRetries = maxRetries
	RetryDelaySeconds = time.Duration(retryDelaySec) * time.Second
}
