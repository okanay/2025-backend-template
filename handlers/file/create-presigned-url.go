// handlers/file/create-presigned-url.go
package FileHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-template/types"
)

// CreatePresignedURL dosya yüklemek için presigned URL oluşturur
func (h *Handler) CreatePresignedURL(c *gin.Context) {
	var input types.CreatePresignedURLInput
	if h.ValidationService.Validate(c, &input) != nil {
		return
	}

	// Presigned URL oluştur
	presignedOutput, err := h.R2Repository.GeneratePresignedURL(c.Request.Context(), types.PresignURLInput{
		Filename:     input.Filename,
		ContentType:  input.ContentType,
		FileCategory: input.FileCategory,
		SizeInBytes:  input.SizeInBytes,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "presigned_url_failed",
			"message": "Yükleme URL'si oluşturulamadı: " + err.Error(),
		})
		return
	}

	// Veritabanında signature kaydı oluştur
	signatureInput := types.UploadSignatureInput{
		PresignedURL: presignedOutput.PresignedURL,
		UploadURL:    presignedOutput.UploadURL,
		Filename:     input.Filename,
		FileType:     input.ContentType,
		FileCategory: input.FileCategory,
		ExpiresAt:    presignedOutput.ExpiresAt,
	}

	signatureID, err := h.FileRepository.CreateUploadSignature(c.Request.Context(), signatureInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "signature_creation_failed",
			"message": "Yükleme kaydı oluşturulamadı: " + err.Error(),
		})
		return
	}

	// Başarılı yanıt döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": types.CreatePresignedURLResponse{
			ID:           signatureID.String(),
			PresignedURL: presignedOutput.PresignedURL,
			UploadURL:    presignedOutput.UploadURL,
			ExpiresAt:    presignedOutput.ExpiresAt,
			Filename:     input.Filename,
		},
	})
}
