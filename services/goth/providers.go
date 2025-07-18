package GothService

import (
	"os"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/apple"
	"github.com/markbates/goth/providers/google"
)

func SetupGothProviders() {
	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			os.Getenv("GOOGLE_REDIRECT_URL"),
			"profile", "email",
		),
		apple.New(
			os.Getenv("APPLE_CLIENT_ID"),
			os.Getenv("APPLE_CLIENT_SECRET"),
			os.Getenv("APPLE_REDIRECT_URL"),
			nil,
			apple.ScopeEmail,
			apple.ScopeName,
		),
	)
}
