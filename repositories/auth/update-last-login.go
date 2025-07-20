package AuthRepository

import (
	"context"

	"github.com/google/uuid"
)

func (r *Repository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := "UPDATE users SET last_login = NOW() WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
