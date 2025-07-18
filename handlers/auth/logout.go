package AuthHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-template/configs"
	"github.com/okanay/backend-template/utils"
)

// Logout, kullanıcının oturumunu sonlandırır
func (h *Handler) Logout(c *gin.Context) {
	// Refresh token'ı cookie'den al
	refreshToken, err := c.Cookie(configs.REFRESH_TOKEN_NAME)
	if err == nil && refreshToken != "" {
		// Token'ı veritabanından iptal et
		err = h.TokenRepository.RevokeRefreshToken(c.Request.Context(), refreshToken, "User logout")
		if err != nil {
			// Hata olsa bile devam et, cookie'leri temizle
		}
	}

	// Cookie'leri temizle
	utils.ClearAuthCookies(c)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Çıkış başarılı",
	})
}
