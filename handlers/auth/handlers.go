package AuthHandler

import (
	UserRepository "github.com/okanay/backend-template/repositories/auth"
	TokenRepository "github.com/okanay/backend-template/repositories/token"
	GothService "github.com/okanay/backend-template/services/goth"
	ValidationService "github.com/okanay/backend-template/services/validation"
)

type Handler struct {
	AuthService       *GothService.Service
	UserRepository    *UserRepository.Repository
	TokenRepository   *TokenRepository.Repository
	ValidationService *ValidationService.Service
}

func NewHandler(authService *GothService.Service, userRepository *UserRepository.Repository, tokenRepository *TokenRepository.Repository, validationService *ValidationService.Service) *Handler {
	return &Handler{
		AuthService:       authService,
		UserRepository:    userRepository,
		TokenRepository:   tokenRepository,
		ValidationService: validationService,
	}
}
