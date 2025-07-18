package FileRepository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// DeleteImage bir resmi siler (veya durumunu 'deleted' olarak günceller)
func (r *Repository) DeleteFile(ctx context.Context, fileID uuid.UUID) error {
	// Soft delete - durumu 'deleted' olarak güncelle
	query := `
		UPDATE files
		SET status = 'deleted', updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, fileID)
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
