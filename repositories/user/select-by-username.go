package UserRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/okanay/backend-template/types"
	"github.com/okanay/backend-template/utils"
)

func (r *Repository) SelectByUsername(ctx context.Context, username string) (types.User, error) {
	defer utils.TimeTrack(time.Now(), "User -> Select User By Username")

	var user types.User

	query := `SELECT * FROM users WHERE username = $1 LIMIT 1`

	// Context kontrolü
	if err := ctx.Err(); err != nil {
		return user, fmt.Errorf("context iptal edildi: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, username)
	if err != nil {
		return user, fmt.Errorf("kullanıcı sorgu hatası: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return user, fmt.Errorf("kullanıcı bulunamadı")
	}

	err = utils.ScanStructByDBTags(rows, &user)
	if err != nil {
		return user, fmt.Errorf("kullanıcı verileri okunamadı: %w", err)
	}

	return user, nil
}
