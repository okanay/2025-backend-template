package UserRepository

import (
	"context"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

// CreateUser, şifre ile yeni bir kullanıcı oluşturur.
// Hem 'users' hem de 'user_details' tablolarına kayıt atar.
func (r *Repository) CreateUser(ctx context.Context, data types.UserCreateRequest) (*types.User, error) {
	// Şifreyi güvenli bir şekilde hash'le.
	hashedPassword, err := utils.EncryptPassword(data.Password)
	if err != nil {
		return nil, err
	}

	// Tüm işlemleri tek bir "atomik" paket olarak sarmak için transaction başlat.
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // Hata olursa her şeyi geri al.

	// 1. Yeni kullanıcıyı 'users' tablosuna ekle ve yeni ID'sini al.
	var userID uuid.UUID
	userQuery := `
        INSERT INTO users (email, auth_provider, hashed_password, role, status)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
	err = tx.QueryRowContext(ctx, userQuery,
		data.Email,
		types.ProviderCredentials,
		hashedPassword,
		types.RoleUser,
		types.UserStatusActive,
	).Scan(&userID)
	if err != nil {
		return nil, err // E-posta zaten varsa, veritabanı hatası burada yakalanır.
	}

	// 2. Yeni kullanıcıya ait boş bir 'user_details' kaydı oluştur.
	detailsQuery := "INSERT INTO user_details (user_id) VALUES ($1)"
	if _, err := tx.ExecContext(ctx, detailsQuery, userID); err != nil {
		return nil, err
	}

	// 3. Her şey başarılıysa, transaction'ı onayla.
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 4. Oluşturulan kullanıcının tam halini veritabanından çek ve döndür.
	return r.SelectByID(ctx, userID)
}
