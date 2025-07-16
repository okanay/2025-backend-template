package UserHandler

import (
	"github.com/gin-gonic/gin"
)

// LoginHandler, kullanıcıyı doğru sağlayıcının izin ekranına yönlendirir.
// Rota: /auth/:provider (örn: /auth/google)
func (h *Handler) LoginHandler(c *gin.Context) {

}

// CallbackHandler, tüm sağlayıcılardan gelen callback'leri işler.
// Rota: /auth/:provider/callback (örn: /auth/google/callback)
func (h *Handler) CallbackHandler(c *gin.Context) {

}
