package types

import (
	"time"

	"github.com/google/uuid"
)

// Permission, string'den türetilmiş özel bir tiptir ve type-safety sağlar.
type Permission string

const (
	// Post Permissions
	CanCreatePost    Permission = "post:create"
	CanEditOwnPost   Permission = "post:edit:own"
	CanEditAnyPost   Permission = "post:edit:any"
	CanDeleteOwnPost Permission = "post:delete:own"
	CanDeleteAnyPost Permission = "post:delete:any"

	// Auth Permissions
	CanRegister            Permission = "auth:register"
	CanLogin               Permission = "auth:login"
	CanUseProvider         Permission = "auth:provider:use"
	CanUseProviderCallback Permission = "auth:provider:callback"
	CanViewProfile         Permission = "auth:view-profile"
	CanLogout              Permission = "auth:logout"

	// File Permissions
	CanGetPresignedURL Permission = "file:presigned-url"
	CanConfirmUpload   Permission = "file:confirm-upload"
	CanListFiles       Permission = "file:list"
	CanDeleteFile      Permission = "file:delete"

	// Content (Github) Permissions
	CanViewCategories  Permission = "content:view-categories"
	CanGetContent      Permission = "content:get"
	CanSaveContent     Permission = "content:save"
	CanViewDraftStatus Permission = "content:view-draft-status"
	CanPublishContent  Permission = "content:publish"
	CanRestartCategory Permission = "content:restart-category"
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
