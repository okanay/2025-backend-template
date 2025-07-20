package GithubHandler

import (
	GithubRepository "github.com/okanay/backend-template/services/github"
	ValidationService "github.com/okanay/backend-template/services/validation"
)

type ContentType string

const (
	ContentTypeI18n   ContentType = "i18n"
	ContentTypeConfig ContentType = "config"
	ContentTypeTheme  ContentType = "theme"
)

type ContentCategory struct {
	Type        ContentType `json:"type"`
	Name        string      `json:"name"`
	Path        string      `json:"path"`
	Description string      `json:"description"`
	Extensions  []string    `json:"extensions"`
	DraftBranch string      `json:"draftBranch"`
	MaxSize     int64       `json:"maxSize"` // bytes
}

type ContentChange struct {
	Path   string `json:"path"`
	Status string `json:"status"` // "added", "modified", "deleted"
}

type DraftStatusResponse struct {
	Category     string          `json:"category"`
	HasChanges   bool            `json:"hasChanges"`
	Branch       string          `json:"branch,omitempty"`
	ChangedFiles []ContentChange `json:"changedFiles"`
	TotalFiles   int             `json:"totalFiles"`
	CanPublish   bool            `json:"canPublish"`
	Message      string          `json:"message"`
}

type Handler struct {
	repository        *GithubRepository.Service
	mainBranch        string
	categories        map[ContentType]ContentCategory
	validationService *ValidationService.Service
}

func NewHandler(r *GithubRepository.Service, validationService *ValidationService.Service) *Handler {
	return &Handler{
		validationService: validationService,
		repository:        r,
		mainBranch:        "main",
		categories: map[ContentType]ContentCategory{
			ContentTypeI18n: {
				Type:        ContentTypeI18n,
				Name:        "Internationalization",
				Path:        "src/messages",
				Description: "Translation files for different languages",
				Extensions:  []string{".json"},
				DraftBranch: "i18n-draft",
				MaxSize:     1024 * 1024, // 1MB
			},
			ContentTypeConfig: {
				Type:        ContentTypeConfig,
				Name:        "Configuration",
				Path:        "src/config",
				Description: "Application configuration files",
				Extensions:  []string{".json", ".js", ".ts"},
				DraftBranch: "config-draft",
				MaxSize:     512 * 1024, // 512KB
			},
			ContentTypeTheme: {
				Type:        ContentTypeTheme,
				Name:        "Theme & Styling",
				Path:        "src/styles",
				Description: "CSS and SCSS styling files",
				Extensions:  []string{".css", ".scss"},
				DraftBranch: "theme-draft",
				MaxSize:     2 * 1024 * 1024, // 2MB
			},
		},
	}
}
