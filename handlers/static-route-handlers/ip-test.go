package StaticRoutesHandler

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) IPTestEndpoint(c *gin.Context) {
	// Mevcut durumda ne geldiğini göster
	clientIP := c.ClientIP()

	// Tüm header bilgilerini topla
	headers := map[string]string{
		"CF-Connecting-IP": c.Request.Header.Get("CF-Connecting-IP"),
		"CF-IPCountry":     c.Request.Header.Get("CF-IPCountry"),
		"CF-Ray":           c.Request.Header.Get("CF-Ray"),
		"X-Forwarded-For":  c.Request.Header.Get("X-Forwarded-For"),
		"X-Real-IP":        c.Request.Header.Get("X-Real-IP"),
		"True-Client-IP":   c.Request.Header.Get("True-Client-IP"),
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

	// Log da at ki console'da görebilelim
	log.Printf("[IP-TEST] Client: %s | CF: %s | Country: %s",
		clientIP,
		headers["CF-Connecting-IP"],
		headers["CF-IPCountry"])

	c.JSON(200, response)
}
