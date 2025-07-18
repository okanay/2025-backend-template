package middlewares

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	AuthRepository "github.com/okanay/backend-template/repositories/auth"
	"github.com/okanay/backend-template/services/cache"
	"github.com/okanay/backend-template/types"
)

var PermissionMap = map[string]types.Permission{
	// File Routes
	"GET:/v1/files":                 types.CanListFiles,
	"DELETE:/v1/files/:id":          types.CanDeleteFile,
	"POST:/v1/files/presigned-url":  types.CanGetPresignedURL,
	"POST:/v1/files/confirm-upload": types.CanConfirmUpload,

	// Github Content Routes
	"GET:/v1/content/categories":             types.CanViewCategories,
	"GET:/v1/content/:category":              types.CanGetContent,
	"POST:/v1/content/:category/save":        types.CanSaveContent,
	"GET:/v1/content/:category/draft-status": types.CanViewDraftStatus,
	"POST:/v1/content/:category/publish":     types.CanPublishContent,
	"DELETE:/v1/content/:category/restart":   types.CanRestartCategory,
}

func PermissionMiddleware(cs cache.CacheService, ar *AuthRepository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Admin rolündeki kullanıcılar her zaman tam yetkilidir, kontrol etmeye gerek yok.
		roleVal, _ := c.Get("user_role")
		if role, ok := roleVal.(types.Role); ok && role == types.RoleAdmin {
			c.Next()
			return
		}

		// 2. Mevcut isteğin metodunu ve yolunu alıp bir anahtar oluştur.
		routeKey := c.Request.Method + ":" + c.FullPath()

		// 3. Merkezi haritadan bu yol için bir izin gerekip gerekmediğini kontrol et.
		requiredPermission, exists := PermissionMap[routeKey]
		if !exists {
			// Bu yol için özel bir izin tanımlanmamış, devam et.
			c.Next()
			return
		}

		// 4. Kullanıcı ID'sini AuthMiddleware'den al.
		userIDVal, _ := c.Get("user_id")
		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Geçersiz kullanıcı kimliği."})
			return
		}

		// 5. Kullanıcının sahip olduğu tüm izinleri getir.
		// TODO: Bu bölüm, performans için cache'den okunacak şekilde geliştirilmelidir.
		userPermissions, _ := ar.SelectPermissionsByUserID(c.Request.Context(), userID)

		// 6. Kullanıcının izinleri arasında gerekli izin var mı diye kontrol et.
		hasPermission := slices.Contains(userPermissions, requiredPermission)

		// 7. Sonuç: İzin varsa devam et, yoksa engelle.
		if hasPermission {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":               "insufficient_permissions",
				"message":             "Bu işlemi yapmak için yetkiniz bulunmamaktadır.",
				"required_permission": requiredPermission,
			})
		}
	}
}
