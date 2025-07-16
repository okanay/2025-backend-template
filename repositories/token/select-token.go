package TokenRepository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

func (r *Repository) SelectRefreshTokenByToken(ctx context.Context, token string) (types.RefreshToken, error) {
	defer utils.TimeTrack(time.Now(), "Token -> Select Refresh Token By Token")

	var refreshToken types.RefreshToken

	// Context kontrolü
	if err := ctx.Err(); err != nil {
		return refreshToken, fmt.Errorf("context iptal edildi: %w", err)
	}

	query := `SELECT * FROM refresh_tokens WHERE token = $1 AND is_revoked = FALSE`

	rows, err := r.db.QueryContext(ctx, query, token)
	if err != nil {
		return refreshToken, fmt.Errorf("token sorgulanırken hata: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return refreshToken, fmt.Errorf("token bulunamadı")
	}

	err = utils.ScanStructByDBTags(rows, &refreshToken)
	if err != nil {
		return refreshToken, fmt.Errorf("token verileri okunurken hata: %w", err)
	}

	return refreshToken, nil
}

func (r *Repository) SelectActiveTokensByUserID(ctx context.Context, userID uuid.UUID) ([]types.RefreshToken, error) {
	defer utils.TimeTrack(time.Now(), "Token -> Select Active Tokens By User ID")

	var tokens []types.RefreshToken

	// Context kontrolü
	if err := ctx.Err(); err != nil {
		return tokens, fmt.Errorf("context iptal edildi: %w", err)
	}

	query := `SELECT * FROM refresh_tokens
              WHERE user_id = $1 AND is_revoked = FALSE AND expires_at > NOW()`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return tokens, fmt.Errorf("token sorgu hatası: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var token types.RefreshToken
		if err := utils.ScanStructByDBTags(rows, &token); err != nil {
			// Hata logla ve devam et
			log.Printf("Token tarama hatası: %v", err)
			continue
		}
		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return tokens, fmt.Errorf("token tarama hatası: %w", err)
	}

	return tokens, nil
}
