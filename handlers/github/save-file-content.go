package GithubHandler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) SaveContent(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required"`
		Path     string `json:"path" binding:"required"`
		Content  string `json:"content" binding:"required"`
		SHA      string `json:"sha"`
		Message  string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	contentType := ContentType(req.Category)
	category, exists := h.categories[contentType]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content category"})
		return
	}

	// Güvenlik kontrolleri
	if !strings.HasPrefix(req.Path, category.Path) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this path"})
		return
	}

	if !h.isAllowedExtension(req.Path, category.Extensions) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File extension not allowed for this category"})
		return
	}

	// Content size kontrolü
	if category.MaxSize > 0 && int64(len(req.Content)) > category.MaxSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Content size exceeds maximum allowed size of %d bytes", category.MaxSize),
		})
		return
	}

	// Content validation (kategoriye göre)
	if err := h.validateContent(req.Content, contentType, req.Path); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Content validation failed",
			"details": err.Error(),
		})
		return
	}

	// Draft branch'ı oluştur (yoksa)
	draftBranch := category.DraftBranch
	exists, err := h.repository.BranchExists(draftBranch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Could not check for draft branch",
			"details": err.Error(),
		})
		return
	}
	if !exists {
		if err := h.repository.CreateBranch(h.mainBranch, draftBranch); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Could not create draft branch",
				"details": err.Error(),
			})
			return
		}
	}

	// Commit message'ı hazırla
	commitMessage := req.Message
	if commitMessage == "" {
		fileName := filepath.Base(req.Path)
		commitMessage = fmt.Sprintf("feat(%s): update %s", req.Category, fileName)
	}

	// Content'i commit et
	err = h.repository.CommitFile(draftBranch, req.Path, req.Content, req.SHA, commitMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Could not commit content",
			"details": err.Error(),
		})
		return
	}

	// Yeni SHA'yı al
	_, newSHA, err := h.repository.GetFileContent(draftBranch, req.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Could not get new SHA after commit",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   fmt.Sprintf("Content saved to %s successfully", draftBranch),
		"sha":      newSHA,
		"branch":   draftBranch,
		"category": req.Category,
		"path":     req.Path,
		"success":  true,
	})
}
