package GothService

import (
	"github.com/markbates/goth"
	"github.com/okanay/backend-template/types"
)

// Service, auth iş mantığını yönetir.
type Service struct {
}

// NewService, yeni bir auth servisi oluşturur.
func NewService() *Service {
	return &Service{}
}

func (s *Service) HandleProviderCallback(gothUser goth.User) *types.ProviderUserData {
	// Goth'un standart 'goth.User' objesini bizim kendi standart 'ProviderUserData' objemize çevir.
	return &types.ProviderUserData{
		Provider:    types.AuthProvider(gothUser.Provider),
		ProviderID:  gothUser.UserID,
		Email:       gothUser.Email,
		DisplayName: gothUser.Name,
		FirstName:   gothUser.FirstName,
		LastName:    gothUser.LastName,
		AvatarURL:   gothUser.AvatarURL,
	}
}
