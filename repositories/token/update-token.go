package TokenRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/okanay/backend-template/utils"
)

// UpdateRefreshTokenLastUsed, bir token'ın son kullanım zamanını günceller.
func (r *Repository) UpdateRefreshTokenLastUsed(ctx context.Context, tokenStr string) error {
	defer utils.TimeTrack(time.Now(), "Token -> UpdateRefreshTokenLastUsed")

	query := `UPDATE refresh_tokens SET last_used_at = NOW() WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, tokenStr)
	return err
}

func (r *Repository) ExtendRefreshTokenExpiry(ctx context.Context, token string, duration time.Duration) error {
	defer utils.TimeTrack(time.Now(), "Token -> Extend Refresh Token Expiry")

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

	now := time.Now()
	newExpiryTime := now.Add(duration)

	query := `UPDATE refresh_tokens
              SET expires_at = $1, last_used_at = $2
              WHERE token = $3 AND is_revoked = FALSE`

	result, err := tx.ExecContext(ctx, query, newExpiryTime, now, token)
	if err != nil {
		return fmt.Errorf("token süre uzatma hatası: %w", err)
	}

	// Etkilenen satır sayısını kontrol et
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("etkilenen satır sayısı alınamadı: %w", err)
	}

	// Token bulunamadıysa hata döndür
	if rowsAffected == 0 {
		return fmt.Errorf("süre uzatılacak token bulunamadı veya iptal edilmiş")
	}

	// Transaction'ı commit et
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit hatası: %w", err)
	}

	return nil
}

func (r *Repository) UpdateRefreshToken(ctx context.Context, oldToken string, newToken string, expiresAt time.Time) error {
	defer utils.TimeTrack(time.Now(), "Token -> Update Refresh Token")

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

	now := time.Now()

	// Önce token'ın var olduğunu ve geçerli olduğunu kontrol et
	checkQuery := `SELECT COUNT(*) FROM refresh_tokens
                   WHERE token = $1 AND is_revoked = FALSE`

	var count int
	err = tx.QueryRowContext(ctx, checkQuery, oldToken).Scan(&count)
	if err != nil {
		return fmt.Errorf("token kontrol hatası: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("güncellenecek token bulunamadı veya iptal edilmiş")
	}

	// Token'ı güncelle
	updateQuery := `UPDATE refresh_tokens
                    SET token = $1, expires_at = $2, last_used_at = $3
                    WHERE token = $4 AND is_revoked = FALSE`

	result, err := tx.ExecContext(ctx, updateQuery, newToken, expiresAt, now, oldToken)
	if err != nil {
		return fmt.Errorf("token güncelleme hatası: %w", err)
	}

	// Etkilenen satır sayısını kontrol et
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("etkilenen satır sayısı alınamadı: %w", err)
	}

	// Token bulunamadıysa hata döndür (double-check)
	if rowsAffected == 0 {
		return fmt.Errorf("güncellenecek token bulunamadı veya iptal edilmiş")
	}

	// Transaction'ı commit et
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit hatası: %w", err)
	}

	return nil
}
