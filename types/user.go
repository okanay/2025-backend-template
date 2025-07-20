package types

import (
	"time"

	"github.com/google/uuid"
)

// Enum Tipleri
type UserStatus string

const (
	UserStatusActive    UserStatus = "Active"
	UserStatusSuspended UserStatus = "Suspended"
	UserStatusDeleted   UserStatus = "Deleted"
)

type Role string

const (
	RoleUser   Role = "User"
	RoleEditor Role = "Editor"
	RoleAdmin  Role = "Admin"
)

type AuthProvider string

const (
	ProviderCredentials AuthProvider = "credentials"
	ProviderGoogle      AuthProvider = "google"
)

// --- Veritabanı Modelleri ---
type User struct {
	ID             uuid.UUID    `db:"id"`
	Email          string       `db:"email"`
	AuthProvider   AuthProvider `db:"auth_provider"`
	HashedPassword *string      `db:"hashed_password"`
	Role           Role         `db:"role"`
	EmailVerified  bool         `db:"email_verified"`
	Status         UserStatus   `db:"status"`
	DeletedAt      *time.Time   `db:"deleted_at"`
	CreatedAt      time.Time    `db:"created_at"`
	LastLogin      time.Time    `db:"last_login"`
	UpdatedAt      time.Time    `db:"updated_at"`
}

type UserDetails struct {
	ID               uuid.UUID `db:"id"`
	UserID           uuid.UUID `db:"user_id"`
	ProviderID       *string   `db:"provider_id"`
	DisplayName      *string   `db:"display_name"`
	FirstName        *string   `db:"first_name"`
	LastName         *string   `db:"last_name"`
	AvatarURL        *string   `db:"avatar_url"`
	PhoneE164        *string   `db:"phone_e164"`
	PhoneCountryCode *string   `db:"phone_country_code"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

// --- API Modelleri ---
type UserView struct {
	ID            uuid.UUID `json:"id"`
	Role          Role      `json:"role"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	DisplayName   *string   `json:"displayName,omitempty"`
	AvatarURL     *string   `json:"avatarUrl,omitempty"`
}

type ProviderUserData struct {
	RawData     map[string]any
	Provider    AuthProvider
	ProviderID  string
	Email       string
	DisplayName string
	FirstName   string
	LastName    string
	AvatarURL   string
}

type UserCreateRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	User        UserView     `json:"user"`
	Permissions []Permission `json:"permissions"`
}

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
