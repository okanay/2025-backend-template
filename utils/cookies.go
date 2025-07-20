package utils

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-template/configs"
)

func SetAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	var secure bool
	var domain string
	domain = os.Getenv("COOKIE_DOMAIN")

	// Access Token Cookie'sini Ayarla
	c.SetCookie(
		configs.ACCESS_TOKEN_NAME,
		accessToken,
		int(configs.ACCESS_TOKEN_DURATION.Seconds()),
		"/",    // Path
		domain, // Domain
		secure, // Secure
		true,   // HttpOnly
	)

	// Refresh Token Cookie'sini Ayarla
	c.SetCookie(
		configs.REFRESH_TOKEN_NAME,
		refreshToken,
		int(configs.REFRESH_TOKEN_DURATION.Seconds()),
		"/",    // Path
		domain, // Domain
		secure, // Secure
		true,   // HttpOnly
	)
}

func ClearAuthCookies(c *gin.Context) {
	var domain string
	if gin.Mode() == gin.ReleaseMode {
		domain = os.Getenv("COOKIE_DOMAIN")
	} else {
		domain = "localhost"
	}

	c.SetCookie(configs.ACCESS_TOKEN_NAME, "", -1, "/", domain, false, true)
	c.SetCookie(configs.REFRESH_TOKEN_NAME, "", -1, "/", domain, false, true)
}
