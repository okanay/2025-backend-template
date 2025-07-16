package TokenRepository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/utils"
)

func (r *Repository) RevokeRefreshToken(ctx context.Context, token string, reason string) error {
	defer utils.TimeTrack(time.Now(), "Token -> Revoke Refresh Token")

	// Transaction başlat
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("transaction başlatılamadı: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Context kontrolü
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context iptal edildi: %w", err)
	}

	query := `UPDATE refresh_tokens SET is_revoked = TRUE, revoked_reason = $1 WHERE token = $2`

	_, err = tx.ExecContext(ctx, query, reason, token)
	if err != nil {
		return fmt.Errorf("token iptal hatası: %w", err)
	}

	// Transaction'ı commit et
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit hatası: %w", err)
	}

	return nil
}

func (r *Repository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, reason string) error {
	defer utils.TimeTrack(time.Now(), "Token -> Revoke All User Tokens")

	// Transaction başlat
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("transaction başlatılamadı: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Context kontrolü
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context iptal edildi: %w", err)
	}

	query := `UPDATE refresh_tokens
              SET is_revoked = TRUE, revoked_reason = $1
              WHERE user_id = $2 AND is_revoked = FALSE`

	result, err := tx.ExecContext(ctx, query, reason, userID)
	if err != nil {
		return fmt.Errorf("token iptal hatası: %w", err)
	}

	// Etkilenen satır sayısını kontrol et
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("etkilenen satır sayısı alınamadı: %w", err)
	}

	// İşlem yapılmadıysa log yaz
	if rowsAffected == 0 {
		log.Printf("İptal edilecek aktif token bulunamadı (userID: %s)", userID)
	}

	// Transaction'ı commit et
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit hatası: %w", err)
	}

	return nil
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
