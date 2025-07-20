package AuthRepository

import (
	"context"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/utils"
)

// UpdatePassword, bir kullanıcının şifresini günceller.
func (r *Repository) UpdatePassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	// Yeni şifreyi güvenli bir şekilde hash'le.
	hashedPassword, err := utils.EncryptPassword(newPassword)
	if err != nil {
		return err
	}

	// Veritabanında güncelleme yap.
	query := "UPDATE users SET hashed_password = $1 WHERE id = $2"
	result, err := r.db.ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		return err
	}

	// Güncellemenin gerçekten bir satırı etkilediğinden emin ol.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return nil
	}

	return nil
}
