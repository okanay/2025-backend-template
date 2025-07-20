package AuthHandler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/okanay/backend-template/configs"
	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

// ProviderHandler, kullanıcıyı doğru sağlayıcının izin ekranına yönlendirir.
// Rota: /auth/:provider (örn: /auth/google)
func (h *Handler) ProviderHandler(c *gin.Context) {
	provider := c.Param("provider")

	// Provider'ı validate et
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_provider",
			"message": "Sağlayıcı parametresi eksik",
		})
		return
	}

	providerName := c.Param("provider")
	_, err := goth.GetProvider(providerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_provider"})
		return
	}

	// Gothic session management için query parameter ekle
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	// Gothic ile OAuth flow'unu başlat
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// CallbackHandler, tüm sağlayıcılardan gelen callback'leri işler.
// Rota: /auth/:provider/callback (örn: /auth/google/callback)
func (h *Handler) CallbackHandler(c *gin.Context) {
	provider := c.Param("provider")

	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_provider",
			"message": "Sağlayıcı parametresi eksik",
		})
		return
	}

	// Gothic session management için query parameter ekle
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	// Gothic ile kullanıcı bilgilerini al
	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		// Eğer auth henüz tamamlanmamışsa, login sayfasına yönlendir
		if err.Error() == "user not authenticated" {
			c.Redirect(http.StatusTemporaryRedirect, "/auth/"+provider)
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "auth_completion_failed",
			"message": "Kimlik doğrulama tamamlanamadı: " + err.Error(),
		})
		return
	}

	// Goth User'ı kendi formatımıza çevir
	providerUserData := h.AuthService.HandleProviderCallback(gothUser)

	// Kullanıcıyı bul veya oluştur
	user, err := h.UserRepository.FindOrCreateFromProvider(c.Request.Context(), providerUserData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "user_processing_failed",
			"message": "Kullanıcı işlemi başarısız: " + err.Error(),
		})
		return
	}

	// Hesap durumunu kontrol et
	if user.Status != types.UserStatusActive {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "account_inactive",
			"message": "Hesabınız aktif değil",
		})
		return
	}

	// Token'ları oluştur
	tokenClaims := types.TokenClaims{
		ID:   user.ID,
		Role: user.Role,
	}

	accessToken, err := utils.GenerateAccessToken(tokenClaims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "token_generation_failed",
			"message": "Token oluşturulamadı",
		})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "token_generation_failed",
			"message": "Refresh token oluşturulamadı",
		})
		return
	}

	// Refresh token'ı veritabanına kaydet
	refreshTokenRequest := types.TokenCreateRequest{
		UserID:    user.ID,
		UserEmail: user.Email,
		Token:     refreshToken,
		IPAddress: utils.GetTrueClientIP(c),
		UserAgent: c.Request.UserAgent(),
		ExpiresAt: time.Now().Add(configs.REFRESH_TOKEN_DURATION),
	}

	_, err = h.TokenRepository.CreateRefreshToken(c.Request.Context(), refreshTokenRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "token_save_failed",
			"message": "Token kaydedilemedi",
		})
		return
	}

	// Son giriş zamanını güncelle
	err = h.UserRepository.UpdateLastLogin(c.Request.Context(), user.ID)
	if err != nil {
		// Log hatasını ama işleme devam et
	}

	// Cookie'leri ayarla
	utils.SetAuthCookies(c, accessToken, refreshToken)

	// Frontend'e redirect et (veya JSON response dön)
	frontendURL := "http://localhost:3000" // Development
	if gin.Mode() == gin.ReleaseMode {
		frontendURL = "https://yourdomain.com" // Production
	}

	// Başarılı giriş sonrası frontend'e yönlendir
	c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/dashboard")
}
