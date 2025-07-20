package middlewares

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-template/configs"
	"golang.org/x/time/rate"
)

type rateLimiterInfo struct {
	limiter  *rate.Limiter
	lastUsed time.Time
}

func RateLimiterMiddleware(limit int, period time.Duration) gin.HandlerFunc {
	limiters := make(map[string]*rateLimiterInfo)
	cleanupInterval := configs.RATE_LIMIT_CLEANUP_DURATION
	var mu sync.RWMutex

	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for ip, info := range limiters {
				if now.Sub(info.lastUsed) > 30*time.Minute {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.RLock()
		info, exists := limiters[ip]
		mu.RUnlock()

		if !exists {
			mu.Lock()
			if info, exists = limiters[ip]; !exists {
				limiter := rate.NewLimiter(rate.Every(period/time.Duration(limit)), limit)
				info = &rateLimiterInfo{
					limiter:  limiter,
					lastUsed: time.Now(),
				}
				limiters[ip] = info
			}
			mu.Unlock()
		} else {
			mu.Lock()
			info.lastUsed = time.Now()
			mu.Unlock()
		}

		if !info.limiter.Allow() {
			c.JSON(429, gin.H{
				"success": false,
				"error":   "rate_limit_exceeded",
				"message": "Çok fazla istek gönderdiniz. Lütfen daha sonra tekrar deneyin.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
