package StaticRoutesHandler

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) Index(c *gin.Context) {

	c.JSON(200, gin.H{
		"Project":   "HOI Holding Backend",
		"Language":  "Golang",
		"Framework": "Gin",
		"Database":  "PostgreSQL",
		"Status":    "System is running successfully.",
		"Version":   "1.0.0",
	})
}
