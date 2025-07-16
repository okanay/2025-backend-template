package types

import (
	"time"

	"github.com/google/uuid"
)

// --- Veritabanı Modeli ---

type RefreshToken struct {
	ID            uuid.UUID `db:"id"`
	UserID        uuid.UUID `db:"user_id"`
	UserEmail     string    `db:"user_email"`
	Token         string    `db:"token"`
	IPAddress     string    `db:"ip_address"`
	UserAgent     string    `db:"user_agent"`
	ExpiresAt     time.Time `db:"expires_at"`
	CreatedAt     time.Time `db:"created_at"`
	LastUsedAt    time.Time `db:"last_used_at"`
	IsRevoked     bool      `db:"is_revoked"`
	RevokedReason *string   `db:"revoked_reason"`
}

// TokenCreateRequest, veritabanına yeni bir refresh token eklemek için kullanılır.
type TokenCreateRequest struct {
	UserID    uuid.UUID
	UserEmail string
	Token     string
	IPAddress string
	UserAgent string
	ExpiresAt time.Time
}

// --- JWT Modeli ---

// TokenClaims, Access Token içinde taşınacak en temel ve değişmez kimlik bilgisidir.
type TokenClaims struct {
	ID   uuid.UUID `json:"id"`
	Role Role      `json:"role"`
}
