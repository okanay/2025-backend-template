package UserRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/okanay/backend-template/utils"
)

func (r *Repository) UpdateLastLogin(ctx context.Context, email string, updateAt time.Time) error {
	defer utils.TimeTrack(time.Now(), "User -> Update Last Login User")

	query := `UPDATE users SET last_login=$1, updated_at=$2 WHERE email=$3`

	// Context kontrolü
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context iptal edildi: %w", err)
	}

	_, err := r.db.ExecContext(ctx, query, updateAt, updateAt, email)
	if err != nil {
		return fmt.Errorf("son giriş güncelleme hatası: %w", err)
	}

	return nil
}
