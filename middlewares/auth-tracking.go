package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-template/configs"
	"github.com/okanay/backend-template/utils"
)

// AuthTrackingMiddleware iş başvuru takip oturumunu doğrular
func AuthTrackingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenCookie, err := c.Cookie(configs.JOBS_TRACKING_COOKIE)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
				"message": "İş başvurularını görüntülemek için takip kodunu girin",
			})
			c.Abort()
			return
		}

		token, err := utils.VerifyApplicationTrackingToken(tokenCookie)
		if err != nil {
			c.SetCookie(configs.JOBS_TRACKING_COOKIE, "", -1, "/", "", false, true)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
				"message": "Oturum süresi doldu, lütfen tekrar giriş yapın",
			})
			c.Abort()
			return
		}

		// Email'i context'e ekle
		c.Set("tracking_email", token.Email)
		c.Next()
	}
}
