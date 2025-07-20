package middlewares

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-template/configs"
	userRepo "github.com/okanay/backend-template/repositories/auth"
	tokenRepo "github.com/okanay/backend-template/repositories/token"
	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

// OptionalAuthMiddleware, bir isteğin kimliğini doğrulamaya çalışır ancak başarısız olursa
// isteği sonlandırmaz. Bunun yerine, kimlik doğrulama durumunu context'e yazar ve devam eder.
func OptionalAuthMiddleware(uRepo *userRepo.Repository, tRepo *tokenRepo.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Access Token'ı doğrulamayı dene.
		accessToken, err := c.Cookie(configs.ACCESS_TOKEN_NAME)
		if err == nil {
			claims, err := utils.ValidateAccessToken(accessToken)
			// Token geçerliyse, context'i ayarla ve devam et.
			if err == nil {
				setAuthenticatedContext(c, claims.ID, claims.Role)
				c.Next()
				return
			}
		}

		// 2. Access Token yoksa veya geçersizse, Refresh Token ile yenilemeyi dene.
		refreshToken, err := c.Cookie(configs.REFRESH_TOKEN_NAME)
		if err != nil {
			// 1.2.0: Refresh token da yoksa, misafir olarak işaretle ve devam et.
			setAnonymousContext(c)
			c.Next()
			return
		}

		// 3. Refresh token'ı ve kullanıcıyı doğrula.
		user, err := validateRefreshTokenAndGetUser(c.Request.Context(), refreshToken, uRepo, tRepo)
		if err != nil {
			// 1.3.2: Refresh token geçersizse veya kullanıcı bulunamazsa, misafir olarak işaretle ve devam et.
			// Burada, süresi dolmuş veya geçersiz token'lar için cookie'leri temizlemek iyi bir pratiktir.
			utils.ClearAuthCookies(c)
			setAnonymousContext(c)
			c.Next()
			return
		}

		// 4. BAŞARILI: Refresh token geçerli. Yeni bir access token oluştur.
		newClaims := types.TokenClaims{ID: user.ID, Role: user.Role}
		newAccessToken, err := utils.GenerateAccessToken(newClaims)
		if err != nil {
			// Token üretiminde sunucu taraflı bir hata olursa bunu logla ama isteği durdurma.
			fmt.Println("OptionalAuthMiddleware: Error generating access token:", err)
			setAnonymousContext(c)
			c.Next()
			return
		}

		// 1.3.1: Yeni token'ları cookie'ye yaz.
		utils.SetAuthCookies(c, newAccessToken, refreshToken)

		// 1.3.3: Context'i kimliği doğrulanmış kullanıcı olarak ayarla.
		setAuthenticatedContext(c, user.ID, user.Role)

		// Refresh token'ın son kullanım zamanını arka planda güncelle.
		go tRepo.UpdateRefreshTokenLastUsed(c.Request.Context(), refreshToken)

		c.Next()
	}
}

// validateRefreshTokenAndGetUser, refresh token'ın veritabanında geçerli olup olmadığını
// kontrol eder ve ilişkili kullanıcıyı döndürür.
func validateRefreshTokenAndGetUser(ctx context.Context, token string, uRepo *userRepo.Repository, tRepo *tokenRepo.Repository) (*types.User, error) {
	// Veritabanından token'ı bul. (Süresi dolmuş veya iptal edilmiş olanlar zaten gelmeyecek)
	dbToken, err := tRepo.SelectRefreshTokenByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Token'a bağlı kullanıcıyı bul.
	user, err := uRepo.SelectByID(ctx, dbToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Kullanıcının aktif olup olmadığını kontrol et.
	if user.Status != types.UserStatusActive {
		return nil, fmt.Errorf("user account is not active")
	}

	return user, nil
}

// setAuthenticatedContext, kimliği doğrulanmış kullanıcı için context değerlerini ayarlar.
func setAuthenticatedContext(c *gin.Context, userID uuid.UUID, role types.Role) {
	c.Set("is_authenticated", true)
	c.Set("user_id", userID)
	c.Set("user_role", role)
}

// setAnonymousContext, kimliği doğrulanmamış (misafir) kullanıcı için context değerlerini ayarlar.
func setAnonymousContext(c *gin.Context) {
	c.Set("is_authenticated", false)
	c.Set("user_id", uuid.Nil)              // uuid.Nil, boş bir UUID'yi temsil eder.
	c.Set("user_role", types.Role("Guest")) // veya boş bir string: types.Role("")
}
