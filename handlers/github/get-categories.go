package GithubHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetCategories(c *gin.Context) {
	categories := make([]ContentCategory, 0, len(h.categories))
	for _, category := range h.categories {
		categories = append(categories, category)
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
		"success":    true,
	})
}
