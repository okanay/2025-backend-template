package AuthRepository

import (
	"context"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

// CreateUser, şifre ile yeni bir kullanıcı oluşturur.
// Hem 'users' hem de 'user_details' tablolarına kayıt atar.
func (r *Repository) CreateUser(ctx context.Context, data types.UserCreateRequest) (*types.User, error) {
	// Kullanıcı tarafından sağlanan şifreyi güvenli bir şekilde hash'le.
	hashedPassword, err := utils.EncryptPassword(data.Password)
	if err != nil {
		return nil, err
	}

	// Tüm işlemleri bir bütün olarak ele almak için bir transaction başlat.
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // Transaction sırasında bir hata oluşursa, yapılan tüm değişiklikleri geri al.

	// Yeni kullanıcı için benzersiz bir UUID oluştur.
	userID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	// 1. Yeni kullanıcıyı 'users' tablosuna ekle.
	userQuery := `
				INSERT INTO users (id, email, auth_provider, hashed_password, role, status)
				VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, userQuery,
		userID,
		data.Email,
		types.ProviderCredentials,
		hashedPassword,
		types.RoleUser,
		types.UserStatusActive,
	)

	// Eğer e-posta zaten mevcutsa, bu noktada bir veritabanı hatası döner.
	if err != nil {
		return nil, err
	}

	// 2. Yeni kullanıcıya ait bir 'user_details' kaydı oluştur.
	detailsID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	detailsQuery := "INSERT INTO user_details (id, user_id) VALUES ($1, $2)"
	if _, err := tx.ExecContext(ctx, detailsQuery, detailsID, userID); err != nil {
		return nil, err
	}

	// 3. Tüm işlemler başarılı bir şekilde tamamlandıysa, transaction'ı onayla.
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 4. Oluşturulan kullanıcıyı veritabanından çek ve döndür.
	return r.SelectByID(ctx, userID)
}
