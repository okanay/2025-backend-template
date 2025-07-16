package middlewares

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/okanay/backend-template/configs"
	TokenRepository "github.com/okanay/backend-template/repositories/token"
	UserRepository "github.com/okanay/backend-template/repositories/user"
	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

func AuthMiddleware(ur *UserRepository.Repository, tr *TokenRepository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie(configs.ACCESS_TOKEN_NAME)

		if err != nil {
			handleTokenRenewal(c, ur, tr)
			return
		}

		claims, err := utils.ValidateAccessToken(accessToken)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				handleTokenRenewal(c, ur, tr)
				return
			}

			handleUnauthorized(c, "Invalid session token.")
			return
		}

		setContextValues(c, claims.ID, claims.Username, claims.Email, claims.Role, claims.EmailVerified, claims.Status, claims.CreatedAt, claims.LastLogin)

		c.Next()
	}
}

func handleTokenRenewal(c *gin.Context, ur *UserRepository.Repository, tr *TokenRepository.Repository) {
	defer utils.TimeTrack(time.Now(), "Token -> Renewal User Token")

	refreshToken, err := c.Cookie(configs.REFRESH_TOKEN_NAME)
	if err != nil {
		handleUnauthorized(c, "Session not found.")
		return
	}

	dbToken, err := tr.SelectRefreshTokenByToken(c, refreshToken)
	if err != nil {
		handleUnauthorized(c, "Invalid session.")
		return
	}

	if dbToken.IsRevoked {
		handleUnauthorized(c, "Session has been revoked.")
		return
	}

	if dbToken.ExpiresAt.Before(time.Now()) {
		handleUnauthorized(c, "Session has expired.")
		return
	}

	user, err := ur.SelectByUsername(c, dbToken.UserUsername)
	if err != nil {
		handleUnauthorized(c, "User not found.")
		return
	}

	if user.Status != types.UserStatusActive {
		handleUnauthorized(c, "Your account is not active.")
		return
	}

	tokenClaims := types.TokenClaims{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	newAccessToken, err := utils.GenerateAccessToken(tokenClaims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "token_generation_failed",
			"message": "An error occurred while renewing the session.",
		})
		c.Abort()
		return
	}

	err = tr.UpdateRefreshTokenLastUsed(c, refreshToken)
	if err != nil {
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		configs.ACCESS_TOKEN_NAME,
		newAccessToken,
		int(configs.ACCESS_TOKEN_DURATION.Seconds()),
		"/",
		"",
		false,
		true,
	)

	setContextValues(c, user.ID, user.Username, user.Email, user.Role, user.EmailVerified, user.Status, user.CreatedAt, user.LastLogin)
	c.Next()
}

func handleUnauthorized(c *gin.Context, message string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(configs.ACCESS_TOKEN_NAME, "", -1, "/", "", false, true)
	c.SetCookie(configs.REFRESH_TOKEN_NAME, "", -1, "/", "", false, true)

	c.JSON(http.StatusUnauthorized, gin.H{
		"success": false,
		"error":   "unauthorized",
		"message": message,
	})
	c.Abort()
}

func setContextValues(c *gin.Context, userID uuid.UUID, username string, email string, role types.Role, emailVerified bool, status types.UserStatus, createdAt time.Time, lastLogin time.Time) {
	c.Set("user_id", userID)
	c.Set("username", username)
	c.Set("email", email)
	c.Set("role", role)
	c.Set("email_verified", emailVerified)
	c.Set("status", status)
	c.Set("created_at", createdAt)
	c.Set("last_login", lastLogin)
}
