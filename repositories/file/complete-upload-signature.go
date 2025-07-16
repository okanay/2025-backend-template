package FileRepository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// CompleteUploadSignature bir yükleme imzasını tamamlandı olarak işaretler ve resim ile ilişkilendirir
func (r *Repository) CompleteUploadSignature(ctx context.Context, signatureID uuid.UUID) error {
	query := `
		UPDATE files_signatures
		SET completed = true
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, signatureID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
