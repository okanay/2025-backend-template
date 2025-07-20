// repositories/r2/presigned.go
package R2Service

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

// GeneratePresignedURL dosya yüklemek için presigned URL oluşturur
func (r *Service) GeneratePresignedURL(ctx context.Context, input types.PresignURLInput) (*types.PresignedURLOutput, error) {
	filename := input.Filename
	fileExt := ""
	dotIndex := strings.LastIndex(filename, ".")

	if dotIndex != -1 {
		fileExt = filename[dotIndex:]  // .docx
		filename = filename[:dotIndex] // Okan-Ay---Vize
	}

	// Sadece dosya adını sanitize et
	safeFilename := sanitizeFilename(filename)

	// Rastgele hash oluştur (8 karakter)
	hashSuffix := utils.GenerateRandomString(8)

	// Final dosya adını oluştur: orijinal-dosya-adi-ABCDEFGH.docx
	finalFilename := fmt.Sprintf("%s-%s%s", safeFilename, hashSuffix, fileExt)

	// File category'ye göre klasör yolu oluştur
	var objectPath string
	fileCategory := "general" // Varsayılan kategori

	if input.FileCategory != "" && len(strings.TrimSpace(input.FileCategory)) > 0 {
		fileCategory = strings.TrimSpace(input.FileCategory)
	}

	objectPath = path.Join(r.folderName, fileCategory, finalFilename)

	// Presigned URL için client oluştur
	presignClient := s3.NewPresignClient(r.client)

	// Presigned URL oluştur
	putObjectRequest, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.bucketName),
		Key:           aws.String(objectPath),
		ContentType:   aws.String(input.ContentType),
		ContentLength: &input.SizeInBytes,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Minute * 5
	})

	if err != nil {
		return nil, fmt.Errorf("presigned URL oluşturulamadı: %w", err)
	}

	// Public erişim URL'sini oluştur
	publicURL := fmt.Sprintf("%s/%s", r.publicURLBase, objectPath)

	return &types.PresignedURLOutput{
		PresignedURL: putObjectRequest.URL,
		UploadURL:    publicURL,
		ObjectKey:    objectPath,
		ExpiresAt:    time.Now().Add(time.Minute * 5),
	}, nil
}

// Dosya adını güvenli hale getiren yardımcı fonksiyon
func sanitizeFilename(filename string) string {
	// Boşlukları tire ile değiştir
	sanitized := strings.ReplaceAll(filename, " ", "-")

	// Sadece alfanumerik, nokta, tire ve alt çizgi karakterlerine izin ver
	sanitized = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, sanitized)

	return sanitized
}
