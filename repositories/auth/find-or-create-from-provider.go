package AuthRepository

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/google/uuid"
	types "github.com/okanay/backend-template/types"
)

// FindOrCreateFromProvider, bir sosyal medya sağlayıcısından (Google, Apple vb.) gelen kullanıcı verisiyle,
// mevcut kullanıcıyı bulur veya yeni bir kullanıcı oluşturur.
func (r *Repository) FindOrCreateFromProvider(ctx context.Context, data *types.ProviderUserData) (*types.User, error) {
	// 1. ADIM: Provider ID ile kullanıcıyı ara.
	// Bu en güvenli yöntemdir çünkü provider_id kullanıcıya özel ve tektir.
	user, err := r.selectByProviderID(ctx, data.Provider, data.ProviderID)
	if err == nil && user != nil {
		// Kullanıcı bulundu. Son giriş zamanını güncelleyip işlemi bitir.
		return r.updateLastLogin(ctx, user.ID)
	}
	// Eğer `sql.ErrNoRows` dışında bir hata varsa, bu beklenmedik bir durumdur, hatayı döndür.
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// 2. ADIM: E-posta ile kullanıcıyı ara ve işlemleri transaction içinde yap.
	// Provider ID ile bulunamadıysa, belki kullanıcı daha önce başka bir yöntemle (örn: şifreyle) kayıt olmuştur.
	// Bu yüzden e-posta ile arama yaparız. Veri tutarlılığı için tüm adımları bir transaction'da topluyoruz.
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err // Transaction başlatılamazsa devam edemeyiz.
	}
	// defer tx.Rollback() -> Hata durumunda veya fonksiyon sonunda commit edilmemişse tüm değişiklikleri geri alır.
	defer tx.Rollback()

	// E-posta ile kullanıcıyı bulmaya çalış.
	existingUser, err := r.selectByEmailTx(ctx, tx, data.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err // `sql.ErrNoRows` dışında bir veritabanı hatası varsa işlemi durdur.
	}

	// --- SENARYO 1: E-posta ile mevcut bir kullanıcı bulundu. ---
	if existingUser != nil {
		// Bu durumda, mevcut kullanıcının hesabını yeni sosyal medya sağlayıcısıyla bağlıyoruz.
		if err := r.linkProviderToUserTx(ctx, tx, existingUser.ID, data); err != nil {
			return nil, err
		}
		// Transaction'ı onayla.
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		// Son giriş zamanını güncelle ve güncel kullanıcı bilgilerini döndür.
		return r.updateLastLogin(ctx, existingUser.ID)
	}

	// --- SENARYO 2: Kullanıcı sistemde hiç yok. Yeni bir kullanıcı oluştur. ---
	// Backend'de yeni bir UUID v7 oluştur.
	// Bu sayede veritabanının UUID üretme fonksiyonuna bağımlı kalmıyoruz.
	newUserID, err := uuid.NewV7()
	if err != nil {
		return nil, err // UUID üretimi başarısız olursa devam edemeyiz.
	}

	// Oluşturulacak yeni kullanıcı için verileri hazırla.
	newUser := &types.User{
		ID:            newUserID, // Backend'de ürettiğimiz ID'yi ata.
		Email:         data.Email,
		AuthProvider:  data.Provider,
		Role:          types.RoleUser,
		EmailVerified: true, // Sağlayıcıdan gelen e-postaları genellikle doğrulanmış kabul ederiz.
		Status:        types.UserStatusActive,
	}

	// Yeni kullanıcıyı 'users' tablosuna ekle.
	if err := r.createUserTx(ctx, tx, newUser); err != nil {
		return nil, err
	}

	// Yeni kullanıcının detaylarını (isim, avatar vb.) 'user_details' tablosuna ekle.
	if err := r.createUserDetailsTx(ctx, tx, newUser.ID, data); err != nil {
		return nil, err
	}

	// Her şey başarılıysa, transaction'ı onayla ve değişiklikleri kalıcı hale getir.
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	log.Printf("Yeni kullanıcı oluşturuldu: %s, Sağlayıcı: %s", newUser.Email, newUser.AuthProvider)
	return newUser, nil
}

// --- Transaction İçinde Çalışan Yardımcı Fonksiyonlar ---

// selectByProviderID, belirli bir sağlayıcı ve provider ID'ye sahip kullanıcıyı bulur.
func (r *Repository) selectByProviderID(ctx context.Context, provider types.AuthProvider, providerID string) (*types.User, error) {
	// users ve user_details tablolarını birleştirerek arama yapıyoruz.
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
	err := r.db.QueryRowContext(ctx, query, providerID, provider).Scan(
		&user.ID, &user.Email, &user.AuthProvider, &user.HashedPassword, &user.Role,
		&user.EmailVerified, &user.Status, &user.DeletedAt, &user.CreatedAt,
		&user.LastLogin, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err // Hata varsa (kayıt bulunamadı dahil) döndür.
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
	// COALESCE fonksiyonu, mevcut değer NULL (boş) ise yeni değeri yazar, değilse eski değeri korur.
	// Bu sayede kullanıcının daha önce girdiği bilgileri ezmemiş oluruz.
	query := `
        UPDATE user_details SET
            provider_id = $1,
            display_name = COALESCE(display_name, $2),
            avatar_url = COALESCE(avatar_url, $3)
        WHERE user_id = $4
    `
	_, err := tx.ExecContext(ctx, query, data.ProviderID, data.DisplayName, data.AvatarURL, userID)
	return err
}

// createUserTx, transaction içinde 'users' tablosuna yeni bir kayıt ekler.
// ID backend'de oluşturulduğu için bu fonksiyon artık ID döndürmez.
func (r *Repository) createUserTx(ctx context.Context, tx *sql.Tx, user *types.User) error {
	// `id` kolonu da eklenecek değerler arasında yer alıyor.
	// RETURNING id kaldırıldı çünkü ID'yi zaten biliyoruz.
	query := `
        INSERT INTO users (id, email, auth_provider, role, email_verified, status)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := tx.ExecContext(ctx, query, user.ID, user.Email, user.AuthProvider, user.Role, user.EmailVerified, user.Status)
	return err
}

// createUserDetailsTx, transaction içinde 'user_details' tablosuna yeni bir kayıt ekler.
// Yeni ID'nin bu fonksiyona parametre olarak geçilmesi gerekir.
func (r *Repository) createUserDetailsTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID, data *types.ProviderUserData) error {
	// 1. user_details kaydı için YENİ ve AYRI bir UUID oluştur.
	newDetailsID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	// 2. Sorguya 'id' kolonunu ve oluşturulan yeni UUID'yi ekle.
	query := `
        INSERT INTO user_details (id, user_id, provider_id, display_name, first_name, last_name, avatar_url)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	// 3. Sorguyu yeni ID ile birlikte çalıştır.
	_, err = tx.ExecContext(ctx, query, newDetailsID, userID, data.ProviderID, data.DisplayName, data.FirstName, data.LastName, data.AvatarURL)
	return err
}

// updateLastLogin, bir kullanıcının son giriş zamanını günceller ve güncel halini döndürür.
func (r *Repository) updateLastLogin(ctx context.Context, userID uuid.UUID) (*types.User, error) {
	var user types.User
	// Sadece temel bilgileri döndürmek yeterlidir.
	query := "UPDATE users SET last_login = NOW() WHERE id = $1 RETURNING id, email, role, status"
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&user.ID, &user.Email, &user.Role, &user.Status)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
