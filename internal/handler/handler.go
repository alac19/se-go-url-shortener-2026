// Package handler provides HTTP handlers for shortlink service.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	service "github.com/alac19/se-go-url-shortener-2026/internal/service"
)

// HandleCreateShortLink 处理 POST /api/links 请求，解析 JSON 中的 URL，调用 service 生成短链并返回 JSON。
func HandleCreateShortLink(s service.Service) gin.HandlerFunc {
	println("handler 层调用 service 层")

	return func(c *gin.Context) {
		var req struct {
			URL string `json:"url"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
			return
		}

		code, err := s.CreateShortLink(req.URL)

		if errors.Is(err, service.ErrInValidURL) || errors.Is(err, service.ErrURLNotReachable) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": "生成短链失败"})
			return
		}

		c.JSON(200, gin.H{"short_url": "http://localhost:8080/" + code})
	}
}

// HandleRedirect 处理 GET /:code 请求，根据路径参数短码查询长链接，并返回 302 重定向。
func HandleRedirect(s service.Service) gin.HandlerFunc {
	println("handler 层调用 service 层")

	return func(c *gin.Context) {
		code := c.Param("code")

		longURL, err := s.Redirect(code)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "短链不存在"})
			return
		}

		c.Redirect(302, longURL)
	}
}
