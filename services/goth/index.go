package GothService

import (
	"os"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"github.com/okanay/backend-template/types"
)

func SetupGothProviders() {
	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			os.Getenv("GOOGLE_REDIRECT_URL"),
			"profile", "email",
		),
	)
}

// Service, auth iş mantığını yönetir.
type Service struct {
}

// NewService, yeni bir auth servisi oluşturur.
func NewService() *Service {
	return &Service{}
}

func (s *Service) HandleProviderCallback(gothUser goth.User) *types.ProviderUserData {

	return &types.ProviderUserData{
		RawData:     gothUser.RawData,
		Provider:    types.AuthProvider(gothUser.Provider),
		ProviderID:  gothUser.UserID,
		Email:       gothUser.Email,
		DisplayName: gothUser.Name,
		FirstName:   gothUser.FirstName,
		LastName:    gothUser.LastName,
		AvatarURL:   gothUser.AvatarURL,
	}
}
