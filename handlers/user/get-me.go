package UserHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

// GetMe, mevcut kullanıcının bilgilerini döndürür
func (h *Handler) GetMe(c *gin.Context) {
	// Auth middleware'den user ID'yi al
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "unauthorized",
			"message": "Kullanıcı kimliği bulunamadı",
		})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "invalid_user_id",
			"message": "Geçersiz kullanıcı kimliği",
		})
		return
	}

	// Kullanıcı bilgilerini veritabanından al
	user, err := h.UserRepository.SelectByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "user_fetch_failed",
			"message": "Kullanıcı bilgileri alınamadı",
		})
		return
	}

	// Kullanıcı durumunu kontrol et
	if user.Status != types.UserStatusActive {
		utils.ClearAuthCookies(c)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "account_inactive",
			"message": "Hesabınız aktif değil",
		})
		return
	}

	// Kullanıcı view'ını oluştur
	userView := types.UserView{
		ID:            user.ID,
		Role:          user.Role,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
	}

	// TODO: Kullanıcı detaylarını da ekleyebilirsiniz
	// userDetails, _ := h.UserRepository.GetUserDetailsByID(c.Request.Context(), userID)
	// if userDetails != nil {
	//     userView.DisplayName = userDetails.DisplayName
	//     userView.AvatarURL = userDetails.AvatarURL
	// }

	// TODO: Kullanıcı izinlerini de ekleyebilirsiniz
	permissions := []types.Permission{} // Boş izin listesi

	response := types.LoginResponse{
		User:        userView,
		Permissions: permissions,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}
