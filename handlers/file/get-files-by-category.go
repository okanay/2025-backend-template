// handlers/file/get-files-by-category.go
package FileHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetFilesByCategory belirli kategorideki dosyaları getirir
func (h *Handler) GetFilesByCategory(c *gin.Context) {
	// Kategori parametresini al
	category := c.Query("c")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_parameter",
			"message": "Kategori parametresi gereklidir",
		})
		return
	}

	// Dosyaları getir
	files, err := h.FileRepository.GetFilesByCategory(c.Request.Context(), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "files_fetch_failed",
			"message": "Dosyalar getirilemedi: " + err.Error(),
		})
		return
	}

	// Yanıt için dönüştür
	var response []gin.H
	for _, file := range files {
		response = append(response, gin.H{
			"id":           file.ID.String(),
			"url":          file.URL,
			"filename":     file.Filename,
			"fileType":     file.FileType,
			"fileCategory": file.FileCategory,
			"sizeInBytes":  file.SizeInBytes,
			"createdAt":    file.CreatedAt,
		})
	}

	// Başarılı yanıt döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}
