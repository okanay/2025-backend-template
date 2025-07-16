package UserHandler

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
	if err := utils.ValidateRequest(c, &input); err != nil {
		return // ValidateRequest zaten yanıt gönderdi
	}

	// 1. Yeni kullanıcı oluştur
	user, err := h.UserRepository.CreateUser(c.Request.Context(), input)
	if err != nil {
		// Database error handler'ı kullan
		if utils.HandleDatabaseError(c, err, "Kullanıcı kaydı") {
			return
		}
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

	// 5. Kullanıcı bilgilerini yanıt olarak döndür
	userView := types.UserView{
		ID:            user.ID,
		Role:          user.Role,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
	}

	response := types.LoginResponse{
		User:        userView,
		Permissions: []types.Permission{}, // Boş izin listesi
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Kayıt başarılı",
		"data":    response,
	})
}
