package GithubHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) PublishCategory(c *gin.Context) {
	categoryParam := c.Param("category")

	var req struct {
		Message string `json:"message"`
	}
	c.ShouldBindJSON(&req) // Optional message

	contentType := ContentType(categoryParam)
	_, exists := h.categories[contentType]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content category"})
		return
	}

	result := h.publishCategory(contentType, req.Message)
	c.JSON(http.StatusOK, result)
}
