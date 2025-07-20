package FileRepository

import (
	"context"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
)

// CreateFileRecord, yüklemesi tamamlanan bir dosyanın bilgilerini 'files' tablosuna kaydeder.
// ID'ler artık veritabanı yerine Go backend'inde (UUIDv7 kullanarak) oluşturulur.
func (r *Repository) CreateFileRecord(ctx context.Context, input types.SaveFileInput) (uuid.UUID, error) {
	// 1. Backend'de yeni bir UUID (versiyon 7) oluştur.
	newFileID, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}

	// 2. SQL sorgusunu, backend'de oluşturulan ID'yi içerecek şekilde düzenle.
	query := `
		INSERT INTO files (
			id, url, filename, file_type, file_category, size_in_bytes, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, 'active'
		)
	` // RETURNING id kaldırıldı.

	// 3. Sorguyu çalıştır.
	_, err = r.db.ExecContext(
		ctx,
		query,
		newFileID, // Backend'de oluşturulan ID
		input.URL,
		input.Filename,
		input.FileType,
		input.FileCategory,
		input.SizeInBytes,
	)

	if err != nil {
		return uuid.Nil, err
	}

	// 4. Başarılıysa, oluşturulan yeni dosya ID'sini döndür.
	return newFileID, nil
}
