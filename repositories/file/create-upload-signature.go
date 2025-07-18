package FileRepository

import (
	"context"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
)

// CreateUploadSignature, bir dosya yükleme işlemi için veritabanında ön-imza kaydı oluşturur.
// ID'ler artık veritabanı tarafından değil, doğrudan Go backend'i içinde (UUIDv7 kullanarak) oluşturulur.
func (r *Repository) CreateUploadSignature(ctx context.Context, input types.UploadSignatureInput) (uuid.UUID, error) {
	// 1. Backend'de yeni bir UUID (versiyon 7) oluştur.
	// Bu ID, veritabanına eklenecek olan kaydın Primary Key'i olacak.
	newSignatureID, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err // UUID üretimi başarısız olursa devam edemeyiz.
	}

	// 2. SQL sorgusunu, backend'de oluşturulan ID'yi içerecek şekilde düzenle.
	query := `
		INSERT INTO files_signatures (
			id, presigned_url, upload_url, filename, file_type, file_category, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	` // RETURNING id kaldırıldı, çünkü ID'yi zaten biliyoruz ve fonksiyonda döndürüyoruz.

	// 3. Sorguyu çalıştır.
	// İlk parametre olarak backend'de oluşturduğumuz 'newSignatureID'yi geçiyoruz.
	_, err = r.db.ExecContext(
		ctx,
		query,
		newSignatureID, // Backend'de oluşturulan ID
		input.PresignedURL,
		input.UploadURL,
		input.Filename,
		input.FileType,
		input.FileCategory,
		input.ExpiresAt,
	)

	if err != nil {
		return uuid.Nil, err
	}

	// 4. Başarılıysa, oluşturulan yeni ID'yi döndür.
	return newSignatureID, nil
}
