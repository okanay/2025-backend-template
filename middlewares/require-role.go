// middlewares/require_role.go
package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-template/types"
)

// RequireRole belirli bir role sahip olmayı gerektiren middleware
func RequireRole(requiredRole types.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, _ := c.Get("user_role")
		userRole, _ := roleVal.(types.Role)

		// Admin her zaman yetkilidir.
		if userRole == types.RoleAdmin {
			c.Next()
			return
		}

		// Gerekli role sahip değilse engelle.
		if userRole != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}
