package FileRepository

import (
	"context"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
)

func (r *Repository) CreateUploadSignature(ctx context.Context, input types.UploadSignatureInput) (uuid.UUID, error) {
	var id uuid.UUID

	query := `
		INSERT INTO files_signatures (
			presigned_url, upload_url, filename, file_type, file_category, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		input.PresignedURL,
		input.UploadURL,
		input.Filename,
		input.FileType,
		input.FileCategory,
		input.ExpiresAt,
	).Scan(&id)

	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
