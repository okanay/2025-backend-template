package AuthRepository

import (
	"context"

	"github.com/google/uuid"
	"github.com/okanay/backend-template/types"
)

// SelectPermissionsByUserID, bir kullanıcının sahip olduğu tüm izinlerin listesini döndürür.
func (r *Repository) SelectPermissionsByUserID(ctx context.Context, userID uuid.UUID) ([]types.Permission, error) {
	query := `
        SELECT p.name
        FROM permissions p
        JOIN user_permissions up ON p.id = up.permission_id
        WHERE up.user_id = $1
    `
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []types.Permission
	for rows.Next() {
		var p types.Permission
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Eğer hiç izni yoksa boş bir slice döner, bu bir hata değildir.
	return permissions, nil
}
