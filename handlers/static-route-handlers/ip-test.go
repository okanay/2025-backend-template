package StaticRoutesHandler

import (
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) IPTestEndpoint(c *gin.Context) {
	// Mevcut durumda ne geldiğini göster
	clientIP := c.ClientIP()

	// Tüm header bilgilerini topla
	headers := map[string]string{
		"CF-Connecting-IP": c.Request.Header.Get("CF-Connecting-IP"),
		"X-Forwarded-For":  c.Request.Header.Get("X-Forwarded-For"),
		"X-Real-IP":        c.Request.Header.Get("X-Real-IP"),
		"X-True-Client-IP": c.Request.Header.Get("X-True-Client-IP"),
		"Cookie":           c.Request.Header.Get("Cookie"),
	}

	// Rate limiter'da kullanılan IP'yi de göster
	rateLimitIP := c.ClientIP()

	response := gin.H{
		"success": true,
		"data": gin.H{
			"message":       "IP bilgileri başarıyla alındı",
			"client_ip":     clientIP,
			"rate_limit_ip": rateLimitIP,
			"is_same":       clientIP == rateLimitIP,
			"headers":       headers,
			"timestamp":     time.Now().Format(time.RFC3339),
		},
	}

	c.JSON(200, response)
}
