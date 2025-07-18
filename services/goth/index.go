package GothService

import (
	"github.com/davecgh/go-spew/spew"
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
	spew.Dump(gothUser)

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
