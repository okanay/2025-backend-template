package AuthRepository

import (
	"context"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
)

// SelectByID, bir kullanıcıyı ID'sine göre bulur.
func (r *Repository) SelectByID(ctx context.Context, id uuid.UUID) (*types.User, error) {
	var user types.User
	query := `SELECT * FROM users WHERE id = $1 LIMIT 1`

	// Veritabanı sorgusunu çalıştır ve sonucu 'user' struct'ına tara.
	err := r.db.QueryRowContext(ctx, query, id).Scan(
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
		return nil, err // Kullanıcı bulunamazsa veya başka bir hata olursa.
	}

	return &user, nil
}
