package FileRepository

import (
	"context"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
)

func (r *Repository) SaveFile(ctx context.Context, input types.SaveFileInput) (uuid.UUID, error) {
	var id uuid.UUID

	query := `
		INSERT INTO files (
			url, filename, file_type, file_category, size_in_bytes, status
		) VALUES (
			$1, $2, $3, $4, $5, 'active'
		) RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		input.URL,
		input.Filename,
		input.FileType,
		input.FileCategory,
		input.SizeInBytes,
	).Scan(&id)

	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
