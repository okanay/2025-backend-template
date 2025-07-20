package AuthHandler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-template/configs"
	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

// Register, şifre ile yeni kullanıcı kaydını yönetir.
func (h *Handler) Register(c *gin.Context) {
	var input types.UserCreateRequest
	if h.ValidationService.Validate(c, &input) != nil {
		return
	}

	// 1. Yeni kullanıcı oluştur
	user, err := h.UserRepository.CreateUser(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "user_creation_failed",
			"message": "Kullanıcı oluşturulamadı",
		})
		return
	}

	// 2. Token'ları oluştur
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

	// 3. Refresh token'ı veritabanına kaydet
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

	// 4. Cookie'leri ayarla
	utils.SetAuthCookies(c, accessToken, refreshToken)

	// 5. Cookie Ayarlandi Frontned - Get-Me Call Et.
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Kayıt başarılı",
	})
}
