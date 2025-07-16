package UserRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/okanay/backend-template/utils"
)

func (r *Repository) UpdatePassword(ctx context.Context, email string, password string) error {
	defer utils.TimeTrack(time.Now(), "User -> Update Password")

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

	hash, err := utils.EncryptPassword(password)
	if err != nil {
		return fmt.Errorf("şifre şifreleme hatası: %w", err)
	}

	// Context kontrolü
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context iptal edildi: %w", err)
	}

	query := `UPDATE users SET hashed_password=$1, updated_at=$2 WHERE email=$3`
	_, err = tx.ExecContext(ctx, query, hash, time.Now(), email)
	if err != nil {
		return fmt.Errorf("şifre güncelleme hatası: %w", err)
	}

	// Transaction'ı commit et
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit hatası: %w", err)
	}

	return nil
}
