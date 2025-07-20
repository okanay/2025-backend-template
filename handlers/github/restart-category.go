package GithubHandler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RestartCategory(c *gin.Context) {
	categoryParam := c.Param("category")
	contentType := ContentType(categoryParam)

	category, exists := h.categories[contentType]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content category"})
		return
	}

	draftBranch := category.DraftBranch
	exists, err := h.repository.BranchExists(draftBranch)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "No draft branch found to delete"})
		return
	}

	if err := h.repository.DeleteBranch(draftBranch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Could not delete draft branch",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   fmt.Sprintf("Draft changes for %s category have been discarded", categoryParam),
		"category": categoryParam,
		"success":  true,
	})
}
