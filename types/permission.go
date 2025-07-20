package types

import (
	"time"

	"github.com/google/uuid"
)

// Permission, string'den türetilmiş özel bir tiptir ve type-safety sağlar.
type Permission string

const (
	// Test
	CanGetIP Permission = "test:ip"

	// File Permissions
	CanGetPresignedURL Permission = "file:presigned-url"
	CanConfirmUpload   Permission = "file:confirm-upload"
	CanListFiles       Permission = "file:list"
	CanDeleteFile      Permission = "file:delete"

	// Content (Github) Permissions
	CanViewGithubCategories  Permission = "github:view-categories"
	CanGetGithubContent      Permission = "github:get"
	CanSaveGithubContent     Permission = "github:save"
	CanViewGithubDraftStatus Permission = "github:view-draft-status"
	CanPublishGithubContent  Permission = "github:publish"
	CanRestartGithubCategory Permission = "github:restart-category"
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
