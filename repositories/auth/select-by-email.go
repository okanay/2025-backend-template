package AuthRepository

import (
	"context"

	"github.com/okanay/backend-template/types"
)

func (r *Repository) SelectByEmail(ctx context.Context, email string) (*types.User, error) {
	var user types.User
	query := `SELECT * FROM users WHERE email = $1 LIMIT 1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.AuthProvider,
		&user.HashedPassword,
		&user.Role,
		&user.EmailVerified,
		&user.Status,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.LastLogin,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
