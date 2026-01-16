package utils

import (
	"fmt"
	"strings"

	"github.com/chmouel/lazyworktree/internal/models"
)

// GeneratePRWorktreeName generates a worktree name from a PR using a template.
// Supports placeholders: {number}, {title}, {generated}, {pr_author}
func GeneratePRWorktreeName(pr *models.PRInfo, template, generatedTitle string) string {
	// Sanitize all components (no length limit here, will truncate final result)
	title := SanitizeBranchName(pr.Title, 0)
	generated := title
	if generatedTitle != "" {
		generated = SanitizeBranchName(generatedTitle, 0)
	}
	author := SanitizeBranchName(pr.Author, 0)

	// Replace placeholders in template
	name := template
	name = strings.ReplaceAll(name, "{number}", fmt.Sprintf("%d", pr.Number))
	name = strings.ReplaceAll(name, "{title}", title)
	name = strings.ReplaceAll(name, "{generated}", generated)
	name = strings.ReplaceAll(name, "{pr_author}", author)

	// Remove trailing hyphens that might result from empty title
	name = strings.TrimRight(name, "-")

	// Truncate to 100 characters
	if len(name) > 100 {
		name = name[:100]
		// Make sure we don't end with a hyphen after truncation
		name = strings.TrimRight(name, "-")
	}

	return name
}

// GenerateIssueWorktreeName generates a worktree name from an issue using a template.
// Supports placeholders: {number}, {title}, {generated}
func GenerateIssueWorktreeName(issue *models.IssueInfo, template, generatedTitle string) string {
	// Sanitize all components (no length limit here, will truncate final result)
	title := SanitizeBranchName(issue.Title, 0)
	generated := title
	if generatedTitle != "" {
		generated = SanitizeBranchName(generatedTitle, 0)
	}

	// Replace placeholders in template
	name := template
	name = strings.ReplaceAll(name, "{number}", fmt.Sprintf("%d", issue.Number))
	name = strings.ReplaceAll(name, "{title}", title)
	name = strings.ReplaceAll(name, "{generated}", generated)

	// Remove trailing hyphens that might result from empty title
	name = strings.TrimRight(name, "-")

	// Truncate to 100 characters
	if len(name) > 100 {
		name = name[:100]
		// Make sure we don't end with a hyphen after truncation
		name = strings.TrimRight(name, "-")
	}

	return name
}
