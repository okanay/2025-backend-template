package types

import (
	"time"

	"github.com/google/uuid"
)

// File dosya tablosundaki kayıtlar için
type File struct {
	ID           uuid.UUID `json:"id"`
	URL          string    `json:"url"`
	FileType     string    `json:"fileType"`
	Filename     string    `json:"filename"`
	FileCategory string    `json:"fileCategory"`
	SizeInBytes  int64     `json:"sizeInBytes"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// FileCreateInput bir dosya oluşturmak için girdi (artık kullanılmıyor olabilir, ama güncelleyelim)
type FileCreateInput struct {
	URL          string `json:"url" validate:"required"`
	Filename     string `json:"filename" validate:"required"`
	FileType     string `json:"fileType" validate:"required"`
	FileCategory string `json:"fileCategory"`
	SizeInBytes  int64  `json:"sizeInBytes" validate:"required,max=10485760"`
}

type UploadSignature struct {
	ID           uuid.UUID `json:"id"`
	PresignedURL string    `json:"presignedUrl"`
	UploadURL    string    `json:"uploadUrl"`
	Filename     string    `json:"filename"`
	FileType     string    `json:"fileType"`
	FileCategory string    `json:"fileCategory"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Completed    bool      `json:"completed"`
	CreatedAt    time.Time `json:"createdAt"`
}

// UploadSignature imza tablosundaki kayıtlar için
type UploadSignatureInput struct {
	PresignedURL string
	UploadURL    string
	Filename     string
	FileType     string
	FileCategory string
	ExpiresAt    time.Time
}

type SaveFileInput struct {
	URL          string
	Filename     string
	FileType     string
	FileCategory string
	SizeInBytes  int64
}

// PresignURLInput Presigned URL oluşturmak için girdi (artık kullanılmıyor olabilir, birleştirildi)
type PresignURLInput struct {
	Filename     string `json:"filename" validate:"required"`
	ContentType  string `json:"contentType" validate:"required"`
	FileCategory string `json:"fileCategory"`
	SizeInBytes  int64  `json:"sizeInBytes" validate:"required,max=10485760"`
}

type PresignedURLOutput struct {
	PresignedURL string    `json:"presignedUrl"`
	UploadURL    string    `json:"uploadUrl"`
	ObjectKey    string    `json:"objectKey"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type CreatePresignedURLInput struct {
	Filename     string `json:"filename" validate:"required"`
	ContentType  string `json:"contentType" validate:"required"`
	FileCategory string `json:"fileCategory" validate:"omitempty"`                 // Opsiyonel olabilir
	SizeInBytes  int64  `json:"sizeInBytes" validate:"required,gt=0,max=10485760"` // 0'dan büyük olmalı
}

type CreatePresignedURLResponse struct {
	ID           string    `json:"id"`
	PresignedURL string    `json:"presignedUrl"`
	UploadURL    string    `json:"uploadUrl"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Filename     string    `json:"filename"`
}

type ConfirmUploadInput struct {
	SignatureID  string `json:"signatureId" validate:"required,uuid"`
	URL          string `json:"url" validate:"required,url"`
	FileCategory string `json:"fileCategory" validate:"omitempty"`
	SizeInBytes  int64  `json:"sizeInBytes" validate:"required,gt=0,max=10485760"`
	Width        int    `json:"width,omitempty" validate:"omitempty,gt=0"`  // Varsa 0'dan büyük olmalı
	Height       int    `json:"height,omitempty" validate:"omitempty,gt=0"` // Varsa 0'dan büyük olmalı
	AltText      string `json:"altText,omitempty"`
}
