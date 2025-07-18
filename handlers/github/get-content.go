package GithubHandler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetContent(c *gin.Context) {
	categoryParam := c.Param("category")
	filePath := c.Query("path")

	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path query parameter is required"})
		return
	}

	contentType := ContentType(categoryParam)
	category, exists := h.categories[contentType]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content category"})
		return
	}

	// Güvenlik kontrolleri
	if !strings.HasPrefix(filePath, category.Path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this path"})
		return
	}

	if !h.isAllowedExtension(filePath, category.Extensions) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File extension not allowed for this category"})
		return
	}

	// Hangi branch'dan okuyacağımızı belirle
	branchToRead := h.mainBranch
	draftExists, err := h.repository.BranchExists(category.DraftBranch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Could not check for draft branch",
			"details": err.Error(),
		})
		return
	}
	if draftExists {
		branchToRead = category.DraftBranch
	}

	content, sha, err := h.repository.GetFileContent(branchToRead, filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":  "Content not found",
			"branch": branchToRead,
			"path":   filePath,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content":  string(content),
		"sha":      sha,
		"branch":   branchToRead,
		"category": category,
		"path":     filePath,
		"success":  true,
	})
}
