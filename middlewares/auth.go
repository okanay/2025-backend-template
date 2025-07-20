// middlewares/auth_middleware.go

package middlewares

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/okanay/backend-template/configs"
	userRepo "github.com/okanay/backend-template/repositories/auth"
	tokenRepo "github.com/okanay/backend-template/repositories/token"
	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

// AuthMiddleware, gelen isteklerde kimlik doğrulama yapar.
// Önce Access Token'ı kontrol eder, geçersizse Refresh Token ile yenilemeye çalışır.
func AuthMiddleware(uRepo *userRepo.Repository, tRepo *tokenRepo.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Access token'ı cookie'den oku.
		accessToken, err := c.Cookie(configs.ACCESS_TOKEN_NAME)
		if err != nil {
			// Access token yoksa, doğrudan yenileme sürecine geç.
			handleTokenRenewal(c, uRepo, tRepo)
			return
		}

		// 2. Access token'ı doğrula.
		claims, err := utils.ValidateAccessToken(accessToken)
		if err != nil {
			// Token süresi dolmuşsa, yenileme sürecine geç.
			if errors.Is(err, jwt.ErrTokenExpired) {
				handleTokenRenewal(c, uRepo, tRepo)
				return
			}
			// Diğer geçersiz token hataları için yetkisiz yanıtı dön.
			handleUnauthorized(c, "Invalid or malformed access token.")
			return
		}

		// 3. Token geçerliyse, kullanıcı kimliğini context'e ekle ve devam et.
		setContextValues(c, claims.ID, claims.Role)
		c.Next()
	}
}

// handleTokenRenewal, refresh token kullanarak yeni bir access token üretir.
func handleTokenRenewal(c *gin.Context, uRepo *userRepo.Repository, tRepo *tokenRepo.Repository) {
	defer utils.TimeTrack(time.Now(), "Auth -> handleTokenRenewal")

	// 1. Refresh token'ı cookie'den oku.
	refreshToken, err := c.Cookie(configs.REFRESH_TOKEN_NAME)
	if err != nil {
		handleUnauthorized(c, "Session not found. Please log in.")
		return
	}

	// 2. Refresh token'ın veritabanında geçerli olup olmadığını kontrol et.
	// SelectRefreshTokenByToken fonksiyonu süresi dolmuş ve iptal edilmişleri zaten eler.
	dbToken, err := tRepo.SelectRefreshTokenByToken(c.Request.Context(), refreshToken)
	if err != nil {
		handleUnauthorized(c, "Invalid session. Please log in again.")
		return
	}

	// 3. Token'a bağlı kullanıcıyı bul ve durumunu kontrol et (hesap askıya alınmış mı vb.).
	user, err := uRepo.SelectByID(c.Request.Context(), dbToken.UserID)
	if err != nil {
		handleUnauthorized(c, "User associated with the session not found.")
		return
	}
	if user.Status != types.UserStatusActive {
		handleUnauthorized(c, "Your account is not active.")
		return
	}

	// 4. Yeni bir Access Token için minimal 'claims' oluştur.
	// En güncel rol bilgisini doğrudan veritabanından gelen 'user' objesinden alıyoruz.
	newClaims := types.TokenClaims{
		ID:   user.ID,
		Role: user.Role,
	}

	newAccessToken, err := utils.GenerateAccessToken(newClaims)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "token_generation_failed"})
		return
	}

	// 5. (Opsiyonel ama önerilen) Refresh token'ın son kullanım zamanını güncelle.
	go tRepo.UpdateRefreshTokenLastUsed(c.Request.Context(), refreshToken)

	// 6. Yeni access token'ı ve mevcut refresh token'ı merkezi bir fonksiyonla cookie'ye yaz.
	// Refresh token'ı tekrar yazmak, cookie'nin ömrünü tarayıcıda da uzatır.
	utils.SetAuthCookies(c, newAccessToken, refreshToken)

	// 7. Kullanıcı kimliğini context'e ekle ve isteğin devam etmesini sağla.
	setContextValues(c, user.ID, user.Role)
	c.Next()
}

// handleUnauthorized, yetkisiz durumlarda cookie'leri temizler ve 401 yanıtı döner.
func handleUnauthorized(c *gin.Context, message string) {
	// Merkezi cookie temizleme fonksiyonunu kullan.
	utils.ClearAuthCookies(c)

	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"success": false,
		"error":   "unauthorized",
		"message": message,
	})
}

// setContextValues, doğrulanmış kullanıcı kimliğini sonraki handler'ların
// kullanabilmesi için Gin context'ine ekler.
func setContextValues(c *gin.Context, userID uuid.UUID, role types.Role) {
	c.Set("user_id", userID)
	c.Set("user_role", role)
}
