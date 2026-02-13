package app

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/chmouel/lazyworktree/internal/app/services"
	"github.com/chmouel/lazyworktree/internal/models"
)

func worktreeNoteKey(path string) string {
	return filepath.Clean(path)
}

func (m *Model) loadWorktreeNotes() {
	notes, err := services.LoadWorktreeNotes(m.getRepoKey(), m.getWorktreeDir())
	if err != nil {
		m.debugf("failed to parse worktree notes: %v", err)
		return
	}
	if notes == nil {
		notes = map[string]models.WorktreeNote{}
	}
	m.worktreeNotes = notes
}

func (m *Model) saveWorktreeNotes() {
	if err := services.SaveWorktreeNotes(m.getRepoKey(), m.getWorktreeDir(), m.worktreeNotes); err != nil {
		m.debugf("failed to write worktree notes: %v", err)
	}
}

func (m *Model) getWorktreeNote(path string) (models.WorktreeNote, bool) {
	if strings.TrimSpace(path) == "" {
		return models.WorktreeNote{}, false
	}
	key := worktreeNoteKey(path)
	note, ok := m.worktreeNotes[key]
	return note, ok
}

func (m *Model) setWorktreeNote(path, noteText string, pinned bool) {
	if strings.TrimSpace(path) == "" {
		return
	}
	if m.worktreeNotes == nil {
		m.worktreeNotes = make(map[string]models.WorktreeNote)
	}

	trimmed := strings.TrimSpace(noteText)
	key := worktreeNoteKey(path)
	if trimmed == "" && !pinned {
		delete(m.worktreeNotes, key)
		m.saveWorktreeNotes()
		return
	}

	m.worktreeNotes[key] = models.WorktreeNote{
		Note:      trimmed,
		Pinned:    pinned,
		UpdatedAt: time.Now().Unix(),
	}
	m.saveWorktreeNotes()
}

func (m *Model) deleteWorktreeNote(path string) {
	if strings.TrimSpace(path) == "" || len(m.worktreeNotes) == 0 {
		return
	}
	key := worktreeNoteKey(path)
	if _, ok := m.worktreeNotes[key]; !ok {
		return
	}
	delete(m.worktreeNotes, key)
	m.saveWorktreeNotes()
}

func (m *Model) migrateWorktreeNote(oldPath, newPath string) {
	if strings.TrimSpace(oldPath) == "" || strings.TrimSpace(newPath) == "" || len(m.worktreeNotes) == 0 {
		return
	}
	oldKey := worktreeNoteKey(oldPath)
	note, ok := m.worktreeNotes[oldKey]
	if !ok {
		return
	}

	delete(m.worktreeNotes, oldKey)
	note.UpdatedAt = time.Now().Unix()
	m.worktreeNotes[worktreeNoteKey(newPath)] = note
	m.saveWorktreeNotes()
}

func (m *Model) pruneStaleWorktreeNotes(worktrees []*models.WorktreeInfo) {
	if len(m.worktreeNotes) == 0 {
		return
	}

	validPaths := make(map[string]bool, len(worktrees))
	for _, wt := range worktrees {
		if wt == nil || strings.TrimSpace(wt.Path) == "" {
			continue
		}
		validPaths[worktreeNoteKey(wt.Path)] = true
	}

	changed := false
	for key := range m.worktreeNotes {
		if !validPaths[key] {
			delete(m.worktreeNotes, key)
			changed = true
		}
	}
	if changed {
		m.saveWorktreeNotes()
	}
}

func (m *Model) worktreeNoteBadge(note models.WorktreeNote) string {
	iconSet := strings.ToLower(strings.TrimSpace(m.config.IconSet))
	if iconSet == "nerd-font-v3" {
		if note.Pinned {
			return "★"
		}
		return "✎"
	}
	if note.Pinned {
		return "[P]"
	}
	return "[N]"
}
