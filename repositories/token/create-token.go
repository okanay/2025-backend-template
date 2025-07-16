package TokenRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

func (r *Repository) CreateRefreshToken(ctx context.Context, request types.TokenCreateRequest) (types.RefreshToken, error) {
	defer utils.TimeTrack(time.Now(), "Token -> Create Token")

	// Transaction başlat
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return types.RefreshToken{}, fmt.Errorf("transaction başlatılamadı: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var token types.RefreshToken

	// Context kontrolü
	if err := ctx.Err(); err != nil {
		return token, fmt.Errorf("context iptal edildi: %w", err)
	}

	query := `INSERT INTO refresh_tokens (user_id, user_email, user_username, token, ip_address, user_agent, expires_at, revoked_reason)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
              RETURNING *`

	rows, err := tx.QueryContext(ctx, query,
		request.UserID, request.UserEmail, request.UserUsername,
		request.Token, request.IPAddress, request.UserAgent,
		request.ExpiresAt, "")

	if err != nil {
		return token, fmt.Errorf("token oluşturma hatası: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return token, fmt.Errorf("token oluşturuldu ancak veri döndürülemedi")
	}

	err = utils.ScanStructByDBTags(rows, &token)
	if err != nil {
		return token, fmt.Errorf("token verileri okunamadı: %w", err)
	}

	// Transaction'ı commit et
	if err = tx.Commit(); err != nil {
		return token, fmt.Errorf("transaction commit hatası: %w", err)
	}

	return token, nil
}
