package middlewares

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-template/configs"
)

func TimeoutMiddleware() gin.HandlerFunc {
	duration := configs.REQUEST_MAX_DURATION

	return func(c *gin.Context) {
		// Context oluştur ve isteğe bağla
		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()

		// Yeni context'i isteğe ata
		c.Request = c.Request.WithContext(ctx)

		// Cevap kanalı oluştur
		done := make(chan struct{}, 1)
		panicChan := make(chan any, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			c.Next()
			done <- struct{}{}
		}()

		select {
		case p := <-panicChan:
			panic(p) // Panic'i yeniden fırlat
		case <-done:
			// İşlem normal şekilde tamamlandı
			return
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				log.Printf("[TIMEOUT] Request timed out: %s %s", c.Request.Method, c.Request.URL.Path)
				c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
					"success": false,
					"error":   "request_timeout",
					"message": "İstek zaman aşımına uğradı.",
				})
			}
		}
	}
}
