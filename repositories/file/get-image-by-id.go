package FileRepository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
)

// GetImageByID bir resmi ID'ye göre getirir
func (r *Repository) GetFileByID(ctx context.Context, fileID uuid.UUID) (*types.File, error) {
	query := `
		SELECT id, url, filename, file_type, file_category, size_in_bytes, status, created_at, updated_at
		FROM files
		WHERE id = $1 AND status = 'active'
	`

	var file types.File
	err := r.db.QueryRowContext(ctx, query, fileID).Scan(
		&file.ID,
		&file.URL,
		&file.Filename,
		&file.FileType,
		&file.FileCategory,
		&file.SizeInBytes,
		&file.Status,
		&file.CreatedAt,
		&file.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Dosya bulunamadı
		}
		return nil, err
	}

	return &file, nil
}
