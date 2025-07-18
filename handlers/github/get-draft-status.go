package GithubHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetDraftStatus(c *gin.Context) {
	categoryParam := c.Param("category")
	contentType := ContentType(categoryParam)

	_, exists := h.categories[contentType]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content category"})
		return
	}

	status := h.getCategoryDraftStatus(contentType)
	c.JSON(http.StatusOK, status)
}
