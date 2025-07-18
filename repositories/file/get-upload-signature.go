package FileRepository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
)

func (r *Repository) GetUploadSignatureByID(ctx context.Context, signatureID uuid.UUID) (*types.UploadSignature, error) {
	query := `
		SELECT id, presigned_url, upload_url, filename, file_type, file_category, expires_at, completed, created_at
		FROM files_signatures
		WHERE id = $1
	`

	var signature types.UploadSignature
	err := r.db.QueryRowContext(ctx, query, signatureID).Scan(
		&signature.ID,
		&signature.PresignedURL,
		&signature.UploadURL,
		&signature.Filename,
		&signature.FileType,
		&signature.FileCategory,
		&signature.ExpiresAt,
		&signature.Completed,
		&signature.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Signature bulunamadı
		}
		return nil, err
	}

	return &signature, nil
}
