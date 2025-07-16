package UserRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

func (r *Repository) CreateUser(ctx context.Context, request types.UserCreateRequest) (types.User, error) {
	defer utils.TimeTrack(time.Now(), "User -> Create User")

	// Transaction başlat
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return types.User{}, fmt.Errorf("transaction başlatılamadı: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var user types.User
	hashedPassword, err := utils.EncryptPassword(request.Password)
	if err != nil {
		return user, fmt.Errorf("şifre şifreleme hatası: %w", err)
	}

	query := `INSERT INTO users (email, username, hashed_password) VALUES ($1, $2, $3) RETURNING *`

	// Context kontrolü ekle
	if err := ctx.Err(); err != nil {
		return user, fmt.Errorf("context iptal edildi: %w", err)
	}

	// Transaction içinde sorgu çalıştır
	rows, err := tx.QueryContext(ctx, query, request.Email, request.Username, hashedPassword)
	if err != nil {
		return user, fmt.Errorf("kullanıcı oluşturma hatası: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return user, fmt.Errorf("kullanıcı oluşturuldu ancak veri döndürülemedi")
	}

	err = utils.ScanStructByDBTags(rows, &user)
	if err != nil {
		return user, fmt.Errorf("kullanıcı verileri okunamadı: %w", err)
	}

	// Transaction'ı commit et
	if err = tx.Commit(); err != nil {
		return user, fmt.Errorf("transaction commit hatası: %w", err)
	}

	return user, nil
}
