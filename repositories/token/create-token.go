package TokenRepository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

// CreateRefreshToken, yeni bir refresh token'ı veritabanına kaydeder.
// ID'ler artık veritabanı tarafından değil, doğrudan Go backend'i içinde (UUIDv7 kullanarak) oluşturulur.
// Bu yaklaşım, veritabanı bağımsızlığını artırır ve sıralı UUID'lerin performans avantajlarından yararlanır.
func (r *Repository) CreateRefreshToken(ctx context.Context, request types.TokenCreateRequest) (*types.RefreshToken, error) {
	defer utils.TimeTrack(time.Now(), "Token -> CreateRefreshToken")

	// 1. Backend'de yeni bir UUID (versiyon 7) oluştur.
	// Bu ID, veritabanına eklenecek olan kaydın Primary Key'i olacak.
	newRefreshTokenID, err := uuid.NewV7()
	if err != nil {
		// UUID üretimi sırasında bir hata olması kritik bir durumdur.
		return nil, err
	}

	// 2. SQL sorgusunu, backend'de oluşturulan ID'yi içerecek şekilde düzenle.
	// 'id' kolonu VALUES kısmına eklendi ve RETURNING'den kaldırıldı.
	// Diğer alanları (created_at vb.) yine de veritabanından almak için RETURNING kullanıyoruz.
	query := `
        INSERT INTO refresh_tokens (id, user_id, user_email, token, ip_address, user_agent, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, user_id, user_email, token, ip_address, user_agent, expires_at, created_at, last_used_at, is_revoked, revoked_reason`

	var token types.RefreshToken
	// 3. Sorguyu çalıştır ve sonucu tara.
	// İlk parametre olarak backend'de oluşturduğumuz 'newRefreshTokenID'yi geçiyoruz.
	err = r.db.QueryRowContext(ctx, query,
		newRefreshTokenID, // Backend'de oluşturulan ID
		request.UserID,
		request.UserEmail,
		request.Token,
		request.IPAddress,
		request.UserAgent,
		request.ExpiresAt,
	).Scan(
		&token.ID, &token.UserID, &token.UserEmail, &token.Token, &token.IPAddress, &token.UserAgent,
		&token.ExpiresAt, &token.CreatedAt, &token.LastUsedAt, &token.IsRevoked, &token.RevokedReason,
	)

	if err != nil {
		// Veritabanı işlemi sırasında bir hata oluşursa, boş bir token ve hatayı döndür.
		return nil, err
	}

	// Her şey başarılıysa, veritabanına kaydedilmiş olan token'ın tam halini döndür.
	return &token, nil
}
