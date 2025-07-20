package AuthHandler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-template/configs"
	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

// Login, şifre ile giriş işlemini yönetir.
func (h *Handler) Login(c *gin.Context) {
	var input types.UserLoginRequest
	if h.ValidationService.Validate(c, &input) != nil {
		return
	}

	// 1. Kullanıcıyı e-posta adresine göre veritabanından bul.
	user, err := h.UserRepository.SelectByEmail(c.Request.Context(), input.Email)
	if err != nil {
		// Eğer kullanıcı bulunamazsa, güvenlik nedeniyle "kullanıcı bulunamadı" demek yerine
		// genel bir "E-posta veya şifre hatalı" mesajı döndürülür.
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid_credentials",
				"message": "E-posta veya şifre hatalı",
			})
			return
		}
		// Diğer veritabanı hataları için sunucu hatası döndür.
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "database_error",
			"message": "Veritabanı hatası",
		})
		return
	}

	// 2. Kullanıcının hesap durumunu kontrol et (aktif mi, askıya alınmış mı vb.).
	if user.Status != types.UserStatusActive {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "account_inactive",
			"message": "Hesabınız aktif değil",
		})
		return
	}

	// --- GÜVENLİK KONTROLÜ: Sosyal Medya Hesabı mı? ---
	// 3. Kullanıcının bir şifresi olup olmadığını kontrol et.
	// Eğer `HashedPassword` alanı `nil` (boş) ise, bu kullanıcı sisteme
	// Google gibi bir sosyal medya sağlayıcısı ile kayıt olmuştur ve şifresi yoktur.
	if user.HashedPassword == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "provider_account", // Hata kodunu daha spesifik hale getirebiliriz.
			"message": "Bu hesap sosyal medya ile oluşturulmuştur. Lütfen o yöntemle giriş yapın.",
		})
		return
	}

	// 4. Eğer şifre varsa, gelen şifre ile veritabanındaki hash'i karşılaştır.
	if !utils.CheckPassword(input.Password, *user.HashedPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "invalid_credentials",
			"message": "E-posta veya şifre hatalı",
		})
		return
	}

	// 5. Başarılı kimlik doğrulama sonrası token'ları oluştur.
	tokenClaims := types.TokenClaims{
		ID:   user.ID,
		Role: user.Role,
	}

	accessToken, err := utils.GenerateAccessToken(tokenClaims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_generation_failed"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_generation_failed"})
		return
	}

	// 6. Yeni refresh token'ı oturum bilgileriyle veritabanına kaydet.
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_save_failed"})
		return
	}

	// 7. Kullanıcının son giriş zamanını güncelle.
	_ = h.UserRepository.UpdateLastLogin(c.Request.Context(), user.ID) // Hata olursa bile akışı kesme.

	// 8. Token'ları güvenli cookie'lere yaz.
	utils.SetAuthCookies(c, accessToken, refreshToken)

	// 9. Frontend'e başarılı yanıtı dön. Frontend bu yanıttan sonra /auth/me isteği yapabilir.
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Giriş başarılı",
	})
}
