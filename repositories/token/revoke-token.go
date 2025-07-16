package TokenRepository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/utils"
)

// RevokeRefreshToken, belirli bir refresh token'ı iptal eder.
func (r *Repository) RevokeRefreshToken(ctx context.Context, tokenStr string, reason string) error {
	defer utils.TimeTrack(time.Now(), "Token -> RevokeRefreshToken")

	query := `UPDATE refresh_tokens SET is_revoked = TRUE, revoked_reason = $1 WHERE token = $2`
	result, err := r.db.ExecContext(ctx, query, reason, tokenStr)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// RevokeAllUserTokens, bir kullanıcıya ait tüm aktif token'ları iptal eder.
func (r *Repository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, reason string) (int64, error) {
	defer utils.TimeTrack(time.Now(), "Token -> RevokeAllUserTokens")

	query := `UPDATE refresh_tokens SET is_revoked = TRUE, revoked_reason = $1 WHERE user_id = $2 AND is_revoked = FALSE`
	result, err := r.db.ExecContext(ctx, query, reason, userID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *Repository) RevokeExpiredTokens() error {
	defer utils.TimeTrack(time.Now(), "Token -> Revoke Expired Tokens")

	query := `UPDATE refresh_tokens SET is_revoked = TRUE, revoked_reason = 'Token expired'
              WHERE expires_at < NOW() AND is_revoked = FALSE`

	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}
