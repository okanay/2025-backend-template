package GithubHandler

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

func (h *Handler) isAllowedExtension(filePath string, allowedExts []string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			return true
		}
	}
	return false
}

func (h *Handler) validateContent(content string, contentType ContentType, filePath string) error {
	switch contentType {
	case ContentTypeI18n, ContentTypeConfig:
		// JSON validation
		var jsonData interface{}
		if err := json.Unmarshal([]byte(content), &jsonData); err != nil {
			return fmt.Errorf("invalid JSON syntax: %s", err.Error())
		}

	case ContentTypeTheme:
		// CSS validation (basic)
		if strings.Contains(content, "javascript:") || strings.Contains(content, "@import") {
			return fmt.Errorf("potentially dangerous CSS content detected")
		}

		// Basic bracket matching
		openBraces := strings.Count(content, "{")
		closeBraces := strings.Count(content, "}")
		if openBraces != closeBraces {
			return fmt.Errorf("mismatched braces in CSS")
		}
	}

	return nil
}

func (h *Handler) getCategoryDraftStatus(contentType ContentType) DraftStatusResponse {
	category := h.categories[contentType]
	draftBranch := category.DraftBranch

	exists, err := h.repository.BranchExists(draftBranch)
	if err != nil {
		return DraftStatusResponse{
			Category:   string(contentType),
			HasChanges: false,
			Message:    "Error checking draft status: " + err.Error(),
			CanPublish: false,
		}
	}

	if !exists {
		return DraftStatusResponse{
			Category:     string(contentType),
			HasChanges:   false,
			ChangedFiles: []ContentChange{},
			TotalFiles:   0,
			CanPublish:   false,
			Message:      "No draft changes found",
		}
	}

	changes, err := h.repository.GetBranchChanges(h.mainBranch, draftBranch)
	if err != nil {
		return DraftStatusResponse{
			Category:   string(contentType),
			HasChanges: true,
			Branch:     draftBranch,
			Message:    "Error getting changes: " + err.Error(),
			CanPublish: false,
		}
	}

	// Sadece bu kategoriye ait dosyaları filtrele
	var categoryChanges []ContentChange
	for _, change := range changes {
		if strings.HasPrefix(change.Path, category.Path) {
			categoryChanges = append(categoryChanges, ContentChange{
				Path:   change.Path,
				Status: change.Status,
			})
		}
	}

	return DraftStatusResponse{
		Category:     string(contentType),
		HasChanges:   len(categoryChanges) > 0,
		Branch:       draftBranch,
		ChangedFiles: categoryChanges,
		TotalFiles:   len(categoryChanges),
		CanPublish:   len(categoryChanges) > 0,
		Message:      fmt.Sprintf("Found %d changed files in %s category", len(categoryChanges), category.Name),
	}
}

func (h *Handler) publishCategory(contentType ContentType, message string) map[string]interface{} {
	category := h.categories[contentType]
	draftBranch := category.DraftBranch

	exists, err := h.repository.BranchExists(draftBranch)
	if err != nil || !exists {
		return map[string]interface{}{
			"success": false,
			"error":   "No draft branch found to publish",
		}
	}

	// Custom commit message
	if message == "" {
		message = fmt.Sprintf("feat(%s): publish changes from %s", string(contentType), draftBranch)
	}

	report, err := h.repository.PublishBranchToMain(draftBranch)
	if err != nil {
		return map[string]interface{}{
			"success":  false,
			"error":    "Failed to publish",
			"details":  report,
			"category": string(contentType),
		}
	}

	// Draft branch'ı sil
	if err := h.repository.DeleteBranch(draftBranch); err != nil {
		// Log warning but don't fail
	}

	return map[string]interface{}{
		"success":  true,
		"message":  fmt.Sprintf("%s category published successfully", category.Name),
		"report":   report,
		"category": string(contentType),
	}
}
