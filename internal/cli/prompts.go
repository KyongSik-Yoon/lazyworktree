package cli

import (
	"fmt"

	"github.com/chmouel/lazyworktree/internal/models"
)

// formatWorktreeForList formats worktree info for selection display.
func formatWorktreeForList(wt *models.WorktreeInfo) string {
	status := "clean"
	if wt.Dirty {
		status = "dirty"
	}

	extraInfo := ""
	if wt.Ahead > 0 {
		extraInfo = fmt.Sprintf(", %d commits ahead", wt.Ahead)
	}
	if wt.Behind > 0 {
		if extraInfo != "" {
			extraInfo += fmt.Sprintf(", %d behind", wt.Behind)
		} else {
			extraInfo = fmt.Sprintf(", %d commits behind", wt.Behind)
		}
	}

	return fmt.Sprintf("%s (%s%s)", wt.Branch, status, extraInfo)
}
