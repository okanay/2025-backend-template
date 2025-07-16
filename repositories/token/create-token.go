package TokenRepository

import (
	"context"
	"time"

	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

// CreateRefreshToken, yeni bir refresh token'ı veritabanına kaydeder.
func (r *Repository) CreateRefreshToken(ctx context.Context, request types.TokenCreateRequest) (*types.RefreshToken, error) {
	defer utils.TimeTrack(time.Now(), "Token -> CreateRefreshToken")

	var token types.RefreshToken
	// GÜNCELLEME: SQL sorgusundan `user_username` kaldırıldı.
	query := `
        INSERT INTO refresh_tokens (user_id, user_email, token, ip_address, user_agent, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, user_id, user_email, token, ip_address, user_agent, expires_at, created_at, last_used_at, is_revoked, revoked_reason`

	err := r.db.QueryRowContext(ctx, query,
		request.UserID,
		request.UserEmail,
		request.Token,
		request.IPAddress,
		request.UserAgent,
		request.ExpiresAt,
	).Scan(
		&token.ID, &token.UserID, &token.UserEmail, &token.Token, &token.IPAddress, &token.UserAgent,
		&token.ExpiresAt, &token.CreatedAt, &token.LastUsedAt, &token.IsRevoked, &token.RevokedReason,
	)

	if err != nil {
		return nil, err
	}
	return &token, nil
}
