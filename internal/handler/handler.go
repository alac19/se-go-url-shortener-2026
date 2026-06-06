package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	service "github.com/alac19/se-go-url-shortener-2026/internal/service"
)

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

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": "生成短链失败"})
			return
		}

		c.JSON(200, gin.H{"short_url": "http://localhost:8080/" + code})
	}
}

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
