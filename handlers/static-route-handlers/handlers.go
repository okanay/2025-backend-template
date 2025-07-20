package StaticRoutesHandler

import ValidationService "github.com/okanay/backend-template/services/validation"

type Handler struct {
	validationService *ValidationService.Service
}

func NewHandler(validationService *ValidationService.Service) *Handler {
	return &Handler{
		validationService: validationService,
	}
}
