package FileRepository

import (
	"context"

	"github.com/okanay/backend-template/types"
)

// GetImagesByUserID kullanıcıya ait resimleri getirir
func (r *Repository) GetFilesByCategory(ctx context.Context, category string) ([]types.File, error) {
	query := `
		SELECT id, url, filename, file_type, file_category, size_in_bytes, status, created_at, updated_at
		FROM files
		WHERE file_category = $1 AND status = 'active'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []types.File
	for rows.Next() {
		var file types.File
		err := rows.Scan(
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
			return nil, err
		}
		files = append(files, file)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}
