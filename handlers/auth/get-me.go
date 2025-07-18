package AuthHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-template/utils"
)

func (h *Handler) GetMe(c *gin.Context) {
	userID, ok := c.MustGet("user_id").(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid_user_id"})
		return
	}

	// Merkezi fonksiyonu çağır.
	response, err := h.assembleLoginResponse(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "kullanıcı hesabı aktif değil" {
			utils.ClearAuthCookies(c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "account_inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "data_assembly_failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}
