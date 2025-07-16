package UserRepository

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/google/uuid"
	types "github.com/okanay/backend-template/types"
)

func (r *Repository) FindOrCreateFromProvider(ctx context.Context, data *types.ProviderUserData) (*types.User, error) {
	// Önce ProviderID ile tam eşleşme arayalım. Bu en hızlı ve en kesin yoldur.
	user, err := r.selectByProviderID(ctx, data.Provider, data.ProviderID)
	if err == nil && user != nil {
		// Kullanıcı doğrudan bulundu, son giriş zamanını güncelle ve döndür.
		return r.updateLastLogin(ctx, user.ID)
	}
	// Eğer `sql.ErrNoRows` dışında bir hata varsa, işlemi durdur.
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// ProviderID ile bulunamadı. Şimdi e-posta ile arayarak mevcut bir hesabı bağlamaya çalışalım.
	// E-posta ile arama ve oluşturma işlemlerini tek bir transaction içinde yapalım.
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // Hata durumunda tüm işlemleri geri al.

	// E-posta ile kullanıcıyı bulmaya çalış.
	existingUser, err := r.selectByEmailTx(ctx, tx, data.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err // Sorgu hatası.
	}

	// --- SENARYO 1: E-posta ile mevcut bir kullanıcı bulundu. ---
	if existingUser != nil {
		// Mevcut kullanıcının detaylarına yeni provider bilgisini ekle (hesap birleştirme).
		if err := r.linkProviderToUserTx(ctx, tx, existingUser.ID, data); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		// Son giriş zamanını güncelle ve güncel kullanıcıyı döndür.
		return r.updateLastLogin(ctx, existingUser.ID)
	}

	// --- SENARYO 2: E-posta ile de kullanıcı bulunamadı. Yeni bir kullanıcı oluştur. ---
	newUser := &types.User{
		Email:         data.Email,
		AuthProvider:  data.Provider,
		Role:          types.RoleUser,
		EmailVerified: true, // Sağlayıcıdan gelen e-postaları genellikle doğrulanmış kabul ederiz.
		Status:        types.UserStatusActive,
	}

	// Yeni kullanıcıyı 'users' tablosuna ekle.
	userID, err := r.createUserTx(ctx, tx, newUser)
	if err != nil {
		return nil, err
	}
	newUser.ID = *userID

	// Yeni kullanıcının detaylarını 'user_details' tablosuna ekle.
	if err := r.createUserDetailsTx(ctx, tx, newUser.ID, data); err != nil {
		return nil, err
	}

	// Her şey başarılı, transaction'ı onayla.
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	log.Printf("Yeni kullanıcı oluşturuldu: %s, Sağlayıcı: %s", newUser.Email, newUser.AuthProvider)
	return newUser, nil
}

// --- Transaction İçinde Çalışan Yardımcı Fonksiyonlar ---

// selectByProviderID, belirli bir sağlayıcı ve ID'ye sahip kullanıcıyı bulur.
func (r *Repository) selectByProviderID(ctx context.Context, provider types.AuthProvider, providerID string) (*types.User, error) {
	// İki tabloyu birleştirerek (JOIN) arama yapıyoruz.
	query := `
        SELECT
            u.id, u.email, u.auth_provider, u.hashed_password, u.role,
            u.email_verified, u.status, u.deleted_at, u.created_at,
            u.last_login, u.updated_at
        FROM users u
        JOIN user_details ud ON u.id = ud.user_id
        WHERE ud.provider_id = $1 AND u.auth_provider = $2
        LIMIT 1
    `
	var user types.User
	// Gelen satırı doğrudan 'user' struct'ının alanlarına tarıyoruz.
	err := r.db.QueryRowContext(ctx, query, providerID, provider).Scan(
		&user.ID, &user.Email, &user.AuthProvider, &user.HashedPassword, &user.Role,
		&user.EmailVerified, &user.Status, &user.DeletedAt, &user.CreatedAt,
		&user.LastLogin, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err // Hata varsa (sql.ErrNoRows dahil) döndür.
	}

	return &user, nil
}

// selectByEmailTx, transaction içinde e-posta ile kullanıcı arar.
func (r *Repository) selectByEmailTx(ctx context.Context, tx *sql.Tx, email string) (*types.User, error) {
	var user types.User
	query := "SELECT id, email, role, status FROM users WHERE email = $1 LIMIT 1"
	err := tx.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.Role, &user.Status)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// linkProviderToUserTx, mevcut bir kullanıcının detaylarına yeni sağlayıcı bilgilerini ekler/günceller.
func (r *Repository) linkProviderToUserTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID, data *types.ProviderUserData) error {
	query := `
        UPDATE user_details SET
            provider_id = $1,
            display_name = COALESCE(display_name, $2), -- Sadece boşsa güncelle
            avatar_url = COALESCE(avatar_url, $3)
        WHERE user_id = $4
    `
	_, err := tx.ExecContext(ctx, query, data.ProviderID, data.DisplayName, data.AvatarURL, userID)
	return err
}

// createUserTx, transaction içinde 'users' tablosuna yeni bir kayıt ekler.
func (r *Repository) createUserTx(ctx context.Context, tx *sql.Tx, user *types.User) (*uuid.UUID, error) {
	var id uuid.UUID
	query := `
        INSERT INTO users (email, auth_provider, role, email_verified, status)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
	err := tx.QueryRowContext(ctx, query, user.Email, user.AuthProvider, user.Role, user.EmailVerified, user.Status).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// createUserDetailsTx, transaction içinde 'user_details' tablosuna yeni bir kayıt ekler.
func (r *Repository) createUserDetailsTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID, data *types.ProviderUserData) error {
	query := `
        INSERT INTO user_details (user_id, provider_id, display_name, first_name, last_name, avatar_url)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := tx.ExecContext(ctx, query, userID, data.ProviderID, data.DisplayName, data.FirstName, data.LastName, data.AvatarURL)
	return err
}

// updateLastLogin, bir kullanıcının son giriş zamanını günceller ve güncel halini döndürür.
func (r *Repository) updateLastLogin(ctx context.Context, userID uuid.UUID) (*types.User, error) {
	var user types.User
	query := "UPDATE users SET last_login = NOW() WHERE id = $1 RETURNING id, email, role, status"
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&user.ID, &user.Email, &user.Role, &user.Status)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
