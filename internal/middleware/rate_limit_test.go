package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockLimiter struct {
	AllowResult bool
}

func (ml *mockLimiter) Allow(ip string) bool {
	return ml.AllowResult
}

func TestHandleRateLimit(t *testing.T) {
	// 设置 Gin 为测试模式，避免日志干扰
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		mockAllowResult bool
		expectedStatus  int
		expectedBody    string
	}{
		{"正常不限流", false, http.StatusOK, "ok"},
		{"限流", true, http.StatusTooManyRequests, `{"error":"too many requests"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ml := &mockLimiter{
				AllowResult: test.mockAllowResult,
			}

			r := gin.New()

			r.POST("/test", HandleRateLimit(ml), func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			})

			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != test.expectedStatus {
				t.Errorf("状态码错: got %d, want %d", rec.Code, test.expectedStatus)
			}

			body := rec.Body.String()

			if body != test.expectedBody {
				t.Errorf("响应体错误: got %s, want %s", body, test.expectedBody)
			}
		})
	}
}
