// handlers/file/delete-file.go
package FileHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DeleteFile bir dosyayı siler
func (h *Handler) DeleteFile(c *gin.Context) {
	// Dosya ID'sini al
	fileIDStr := c.Param("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_file_id",
			"message": "Geçersiz dosya ID'si",
		})
		return
	}

	// Dosya bilgilerini getir
	file, err := h.FileRepository.GetFileByID(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "file_fetch_failed",
			"message": "Dosya bilgileri getirilemedi: " + err.Error(),
		})
		return
	}

	if file == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "file_not_found",
			"message": "Dosya bulunamadı",
		})
		return
	}

	// R2'den dosyayı sil
	// URL'den object key'i çıkarmak gerekiyor, eğer URL: https://base/uploads/image.jpg ise
	// objectKey: uploads/image.jpg olmalı
	objectKey := extractObjectKeyFromURL(file.URL)

	err = h.R2Repository.DeleteObject(c.Request.Context(), objectKey)
	if err != nil {
		// Sadece loglama yap, silme işlemine devam et
		// log.Printf("R2 object deletion failed: %v", err)
	}

	// Veritabanından dosyayı sil
	err = h.FileRepository.DeleteFile(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "file_delete_failed",
			"message": "Dosya silinemedi: " + err.Error(),
		})
		return
	}

	// Başarılı yanıt döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Dosya başarıyla silindi",
	})
}

// URL'den object key'i çıkaran yardımcı fonksiyon
func extractObjectKeyFromURL(url string) string {
	// Base URL'yi çıkar, örn: https://files.project-test.info/uploads/image.jpg -> uploads/image.jpg
	// Bu örnek için basit bir implementasyon, gerçek uygulamada daha sağlam bir çözüm gerekebilir
	baseURL := "https://files.project-test.info/"
	if len(url) > len(baseURL) && url[:len(baseURL)] == baseURL {
		return url[len(baseURL):]
	}
	return url
}
