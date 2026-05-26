package auth

import (
	"os"
	"strings"

	"edu-schedule-system/internal/pkg/env"
)

func recoveryPreviewEnabled() bool {
	if !env.Active().IsDev() {
		return false
	}
	value, ok := os.LookupEnv("AUTH_PASSWORD_RECOVERY_PREVIEW_ENABLED")
	if !ok {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
