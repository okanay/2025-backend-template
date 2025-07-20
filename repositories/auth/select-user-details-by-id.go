package AuthRepository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
)

// SelectUserDetailsByID, bir kullanıcıya ait profil detaylarını getirir.
func (r *Repository) SelectUserDetailsByID(ctx context.Context, userID uuid.UUID) (*types.UserDetails, error) {
	var details types.UserDetails
	query := `
        SELECT id, user_id, provider_id, display_name, first_name, last_name, avatar_url
        FROM user_details
        WHERE user_id = $1
        LIMIT 1
    `
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&details.ID,
		&details.UserID,
		&details.ProviderID,
		&details.DisplayName,
		&details.FirstName,
		&details.LastName,
		&details.AvatarURL,
	)

	if err != nil {
		// Eğer hiç detay kaydı yoksa bu bir hata değildir, sadece nil döneriz.
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &details, nil
}
