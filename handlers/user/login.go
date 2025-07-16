package UserHandler

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
	if err := utils.ValidateRequest(c, &input); err != nil {
		return // ValidateRequest zaten yanıt gönderdi
	}

	// 1. Kullanıcıyı e-posta ile bul
	user, err := h.UserRepository.SelectByEmail(c.Request.Context(), input.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid_credentials",
				"message": "E-posta veya şifre hatalı",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "database_error",
			"message": "Veritabanı hatası",
		})
		return
	}

	// 2. Hesap durumunu kontrol et
	if user.Status != types.UserStatusActive {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "account_inactive",
			"message": "Hesabınız aktif değil",
		})
		return
	}

	// 3. Şifre kontrolü
	if user.HashedPassword == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "invalid_credentials",
			"message": "Bu hesap sosyal medya ile oluşturulmuş",
		})
		return
	}

	if !utils.CheckPassword(input.Password, *user.HashedPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "invalid_credentials",
			"message": "E-posta veya şifre hatalı",
		})
		return
	}

	// 4. Token'ları oluştur
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

	// 5. Refresh token'ı veritabanına kaydet
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

	// 6. Son giriş zamanını güncelle
	err = h.UserRepository.UpdateLastLogin(c.Request.Context(), user.ID)
	if err != nil {
		// Log hatasını ama işleme devam et
	}

	// 7. Cookie'leri ayarla
	utils.SetAuthCookies(c, accessToken, refreshToken)

	// 8. Kullanıcı bilgilerini yanıt olarak döndür
	userView := types.UserView{
		ID:            user.ID,
		Role:          user.Role,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
	}

	// TODO: Kullanıcı izinlerini de ekleyebilirsiniz
	response := types.LoginResponse{
		User:        userView,
		Permissions: []types.Permission{}, // Boş izin listesi
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Giriş başarılı",
		"data":    response,
	})
}
