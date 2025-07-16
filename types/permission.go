package types

import (
	"time"

	"github.com/google/uuid"
)

// Permission, string'den türetilmiş özel bir tiptir ve type-safety sağlar.
type Permission string

const (
	// Post Permissions
	CanReadLongPost  Permission = "post:read:long"
	CanCreatePost    Permission = "post:create"
	CanEditOwnPost   Permission = "post:edit:own"
	CanEditAnyPost   Permission = "post:edit:any"
	CanDeleteOwnPost Permission = "post:delete:own"
	CanDeleteAnyPost Permission = "post:delete:any"

	// Media Permissions
	CanUploadImage Permission = "media:upload:image"
	CanUploadVideo Permission = "media:upload:video"
	CanDeleteMedia Permission = "media:delete"

	// User Management Permissions
	CanListUsers   Permission = "user:list"
	CanEditUser    Permission = "user:edit"
	CanManagePerms Permission = "user:permissions:manage"
)

// --- Veritabanı Modelleri ---

// PermissionDB, 'permissions' tablosunu temsil eder.
type PermissionDB struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}

// UserPermission, 'user_permissions' bağlantı tablosunu temsil eder.
type UserPermission struct {
	UserID       uuid.UUID `db:"user_id"`
	PermissionID uuid.UUID `db:"permission_id"`
}
