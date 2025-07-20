package configs

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsConfig() gin.HandlerFunc {
	var origins = []string{
		"https://www.hoi.com.tr",
		"https://tanstack.hoi.com.tr",
		"https://hoi.com.tr",
	}

	if gin.Mode() == gin.DebugMode {
		origins = append(origins, "http://localhost:3000", "http://localhost:8788")
	}

	return cors.New(cors.Config{
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "HEAD", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Accept", "Origin", "X-Requested-With", "cf-turnstile-token", "CF-Turnstile-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowOrigins:     origins,
		AllowCredentials: true,
		MaxAge:           60 * 24 * 30,
	})
}
