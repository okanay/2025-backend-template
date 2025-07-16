package UserHandler

import (
	TokenRepository "github.com/okanay/backend-template/repositories/token"
	UserRepository "github.com/okanay/backend-template/repositories/user"
	GothService "github.com/okanay/backend-template/services/goth"
)

type Handler struct {
	AuthService     *GothService.Service
	UserRepository  *UserRepository.Repository
	TokenRepository *TokenRepository.Repository
}

func NewHandler(authService *GothService.Service, userRepository *UserRepository.Repository, tokenRepository *TokenRepository.Repository) *Handler {
	return &Handler{
		AuthService:     authService,
		UserRepository:  userRepository,
		TokenRepository: tokenRepository,
	}
}
