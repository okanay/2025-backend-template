package TokenRepository

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

// SelectRefreshTokenByToken, bir token dizesine göre geçerli (iptal edilmemiş ve süresi dolmamış) token'ı bulur.
func (r *Repository) SelectRefreshTokenByToken(ctx context.Context, tokenStr string) (*types.RefreshToken, error) {
	defer utils.TimeTrack(time.Now(), "Token -> SelectRefreshTokenByToken")

	var refreshToken types.RefreshToken
	// GÜNCELLEME: Sorgu, son kullanma ve iptal durumunu da kontrol ederek daha güvenli hale getirildi.
	query := `
        SELECT id, user_id, user_email, token, ip_address, user_agent, expires_at, created_at, last_used_at, is_revoked, revoked_reason
        FROM refresh_tokens
        WHERE token = $1 AND is_revoked = FALSE AND expires_at > NOW()
        LIMIT 1`

	err := r.db.QueryRowContext(ctx, query, tokenStr).Scan(
		&refreshToken.ID, &refreshToken.UserID, &refreshToken.UserEmail, &refreshToken.Token, &refreshToken.IPAddress, &refreshToken.UserAgent,
		&refreshToken.ExpiresAt, &refreshToken.CreatedAt, &refreshToken.LastUsedAt, &refreshToken.IsRevoked, &refreshToken.RevokedReason,
	)

	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

// SelectActiveTokensByUserID, bir kullanıcıya ait tüm aktif oturumları listeler.
func (r *Repository) SelectActiveTokensByUserID(ctx context.Context, userID uuid.UUID) ([]types.RefreshToken, error) {
	defer utils.TimeTrack(time.Now(), "Token -> SelectActiveTokensByUserID")

	query := `
        SELECT id, user_id, user_email, token, ip_address, user_agent, expires_at, created_at, last_used_at
        FROM refresh_tokens
        WHERE user_id = $1 AND is_revoked = FALSE AND expires_at > NOW()
        ORDER BY last_used_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []types.RefreshToken
	for rows.Next() {
		var token types.RefreshToken
		if err := rows.Scan(&token.ID, &token.UserID, &token.UserEmail, &token.Token, &token.IPAddress, &token.UserAgent, &token.ExpiresAt, &token.CreatedAt, &token.LastUsedAt); err != nil {
			log.Printf("Token tarama hatası (SelectActiveTokensByUserID): %v", err)
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}
