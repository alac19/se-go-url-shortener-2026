package handler

// import "net/http"

// func InitHandler(operation string) http.HandlerFunc {
// 	switch {
// 	case operation == "GET":
// 		println("handler 层调用 service 层")

// 		return func(w http.ResponseWriter, r *http.Request) {}
// 	case operation == "POST":
// 		println("handler 层调用 service 层")

// 		return func(w http.ResponseWriter, r *http.Request) {}
// 	}

// 	return func(w http.ResponseWriter, r *http.Request) {}
// }

// import (
// 	"net/http"

// 	service "github.com/alac19/se-go-url-shortener-2026/internal/service"
// )

// func GetCreateHandler(s service.Service) http.HandlerFunc {
// 	println("handler 层调用 service 层")

// 	return func(w http.ResponseWriter, r *http.Request) {
// 		s.DoGet()
// 	}
// }

// func PostCreateHandler(s service.Service) http.HandlerFunc {
// 	println("handler 层调用 service 层")

// 	return func(w http.ResponseWriter, r *http.Request) {
// 		s.DoPost()
// 	}
// }

import (
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
			// 处理错误...
		}

		code := s.CreatShortLink(req.URL)
		c.JSON(200, gin.H{"shortCode": code})
	}
}

func HandleRedirect(s service.Service) gin.HandlerFunc {
	println("handler 层调用 service 层")

	return func(c *gin.Context) {
		code := c.Param("code")
		longURL := s.Redirect(code)
		c.Redirect(302, longURL)
	}
}
