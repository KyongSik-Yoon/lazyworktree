package utils

import (
	"regexp"
	"strings"
)

// SanitizeBranchName sanitizes a branch/title for use as a worktree directory name.
// - Converts to lowercase
// - Keeps only alphanumeric characters, replaces everything else with hyphens
// - Collapses consecutive hyphens
// - Trims leading/trailing hyphens
// - Optionally limits length (0 = no limit)
func SanitizeBranchName(name string, maxLength int) string {
	sanitized := strings.ToLower(strings.TrimSpace(name))

	// Replace all non-alphanumeric characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	sanitized = re.ReplaceAllString(sanitized, "-")

	// Collapse consecutive hyphens
	sanitized = regexp.MustCompile(`-+`).ReplaceAllString(sanitized, "-")

	// Trim leading/trailing hyphens
	sanitized = strings.Trim(sanitized, "-")

	// Apply length limit if specified
	if maxLength > 0 && len(sanitized) > maxLength {
		sanitized = strings.TrimRight(sanitized[:maxLength], "-")
	}

	return sanitized
}
