package middlewares

import (
	"log"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	AuthRepository "github.com/okanay/backend-template/repositories/auth"
	"github.com/okanay/backend-template/services/cache"
	"github.com/okanay/backend-template/types"
)

var PermissionMap = map[string]types.Permission{
	// "GET:/v1/test/ip-address": types.CanGetIP,

	// File Routes
	"GET:/v1/files":                 types.CanListFiles,
	"DELETE:/v1/files/:id":          types.CanDeleteFile,
	"POST:/v1/files/presigned-url":  types.CanGetPresignedURL,
	"POST:/v1/files/confirm-upload": types.CanConfirmUpload,

	// Github Content Routes
	"GET:/v1/github/categories":             types.CanViewGithubCategories,
	"GET:/v1/github/:category":              types.CanGetGithubContent,
	"POST:/v1/github/:category/save":        types.CanSaveGithubContent,
	"GET:/v1/github/:category/draft-status": types.CanViewGithubDraftStatus,
	"POST:/v1/github/:category/publish":     types.CanPublishGithubContent,
	"DELETE:/v1/github/:category/restart":   types.CanRestartGithubCategory,
}

func PermissionMiddleware(cs cache.CacheService, ar *AuthRepository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Admin rolündeki kullanıcılar her zaman tam yetkilidir.
		roleVal, _ := c.Get("user_role")
		if role, ok := roleVal.(types.Role); ok && role == types.RoleAdmin {
			c.Next()
			return
		}

		// 2. Mevcut isteğin yoluna göre bir izin gerekip gerekmediğini kontrol et.
		routeKey := c.Request.Method + ":" + c.FullPath()
		requiredPermission, exists := PermissionMap[routeKey]
		if !exists {
			c.Next() // Bu rota için bir izin tanımlanmamış, devam et.
			return
		}

		// 3. Kullanıcı kimliğini al.
		userIDVal, _ := c.Get("user_id")
		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Geçersiz kullanıcı kimliği."})
			return
		}

		// 4. Kullanıcının izinlerini, yeni cache yeteneğimizi kullanarak getir.
		var userPermissions []types.Permission

		// Cache'de veri bulunamazsa çalıştırılacak olan veritabanı sorgusunu tanımla.
		dbFallback := func() (any, error) {
			return ar.SelectPermissionsByUserID(c.Request.Context(), userID)
		}

		// "permissions" grubunda, bu kullanıcıya ait izinleri cache'den getirmeyi dene.
		// Bulamazsan, dbFallback fonksiyonunu çalıştırıp sonucu cache'e yaz.
		err := cs.GetOrSet(
			cache.PermissionCacheGroup, // Merkezi bir yerden cache grup adı
			userID.String(),            // Cache için benzersiz anahtar
			&userPermissions,           // Sonucun yazılacağı hedef
			dbFallback,                 // Cache'de yoksa çalışacak fonksiyon
		)

		if err != nil {
			log.Printf("PERMISSION_CHECK_ERROR: UserID %s için izinler alınamadı: %v", userID, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "permission_check_failed"})
			return
		}

		// 5. Kullanıcının sahip olduğu izinler arasında gerekli olan var mı diye kontrol et.
		if slices.Contains(userPermissions, requiredPermission) {
			c.Next() // İzin var, devam et.
		} else {
			// İzin yok, engelle.
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":               "insufficient_permissions",
				"message":             "Bu işlemi yapmak için yetkiniz bulunmamaktadır.",
				"required_permission": requiredPermission,
			})
		}
	}
}
