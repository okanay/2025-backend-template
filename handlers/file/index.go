// handlers/file/index.go
package FileHandler

import (
	FileRepository "github.com/okanay/backend-template/repositories/file"
	R2Repository "github.com/okanay/backend-template/services/r2"
	ValidationService "github.com/okanay/backend-template/services/validation"
)

type Handler struct {
	FileRepository    *FileRepository.Repository
	R2Repository      *R2Repository.Service
	ValidationService *ValidationService.Service
}

func NewHandler(f *FileRepository.Repository, r2 *R2Repository.Service, validationService *ValidationService.Service) *Handler {
	return &Handler{
		FileRepository:    f,
		R2Repository:      r2,
		ValidationService: validationService,
	}
}
