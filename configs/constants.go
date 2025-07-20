package configs

import (
	"time"
)

const (
	// Project Rules
	PROJECT_NAME = "Backend Template"

	// TIMEOUT RULES
	REQUEST_MAX_DURATION = 120 * time.Second

	// RATE LIMIT RULES
	RATE_LIMIT_CLEANUP_DURATION = 1 * time.Hour

	// Session Rules
	REFRESH_TOKEN_LENGTH   = 32
	REFRESH_TOKEN_DURATION = 30 * 24 * time.Hour
	REFRESH_TOKEN_NAME     = "refresh_token"
	ACCESS_TOKEN_NAME      = "access_token"
	ACCESS_TOKEN_DURATION  = 5 * time.Minute
	JWT_ISSUER             = "backend-template"
)
