package app

import (
	"os"
	"strings"

	"github.com/chmouel/lazyworktree/internal/models"
)

// determineCurrentWorktree finds the worktree that matches the current working directory.
func (m *Model) determineCurrentWorktree() *models.WorktreeInfo {
	if wt := m.selectedWorktree(); wt != nil {
		return wt
	}

	if cwd, err := os.Getwd(); err == nil {
		for _, wt := range m.worktrees {
			if strings.HasPrefix(cwd, wt.Path) {
				return wt
			}
		}
	}

	for _, wt := range m.worktrees {
		if wt.IsMain {
			return wt
		}
	}

	return nil
}

// selectedWorktree returns the currently selected worktree from the filtered list.
func (m *Model) selectedWorktree() *models.WorktreeInfo {
	indices := []int{m.worktreeTable.Cursor(), m.selectedIndex}
	for _, idx := range indices {
		if wt := m.worktreeAtIndex(idx); wt != nil {
			return wt
		}
	}
	return nil
}

// worktreeAtIndex returns the worktree at the given index in the filtered list.
func (m *Model) worktreeAtIndex(idx int) *models.WorktreeInfo {
	if idx < 0 || idx >= len(m.filteredWts) {
		return nil
	}
	return m.filteredWts[idx]
}
