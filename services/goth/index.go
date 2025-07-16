package GothService

import (
	"context"

	"github.com/markbates/goth"
	userRepo "github.com/okanay/backend-template/repositories/user"
	types "github.com/okanay/backend-template/types"
)

// Service, auth iş mantığını yönetir.
type Service struct {
	userRepo *userRepo.Repository
}

// NewService, yeni bir auth servisi oluşturur.
func NewService(userRepo *userRepo.Repository) *Service {
	return &Service{userRepo: userRepo}
}

// HandleProviderCallback, goth'tan gelen kullanıcıyı işler ve veritabanına kaydeder/günceller.
func (s *Service) HandleProviderCallback(gothUser goth.User) (*types.User, error) {
	// Goth'un standart 'goth.User' objesini bizim kendi standart 'ProviderUserData' objemize çevir.
	providerData := &types.ProviderUserData{
		Provider:    types.AuthProvider(gothUser.Provider),
		ProviderID:  gothUser.UserID,
		Email:       gothUser.Email,
		DisplayName: gothUser.Name,
		FirstName:   gothUser.FirstName,
		LastName:    gothUser.LastName,
		AvatarURL:   gothUser.AvatarURL,
	}

	// Repository katmanını çağırarak kullanıcıyı bul veya oluştur.
	// Servis katmanı, veritabanı sorgusunun nasıl yazıldığını bilmez.
	user, err := s.userRepo.FindOrCreateFromProvider(context.Background(), providerData)
	if err != nil {
		return nil, err
	}

	return user, nil
}
