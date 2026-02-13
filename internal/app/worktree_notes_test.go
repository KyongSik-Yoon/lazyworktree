package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	appscreen "github.com/chmouel/lazyworktree/internal/app/screen"
	"github.com/chmouel/lazyworktree/internal/config"
	"github.com/chmouel/lazyworktree/internal/models"
)

func TestSetAndLoadWorktreeNotes(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	path := "/tmp/worktrees/feature-a"
	m.setWorktreeNote(path, "line one\nline two", true)

	m2 := NewModel(cfg, "")
	m2.repoKey = testRepoKey
	m2.loadWorktreeNotes()

	note, ok := m2.getWorktreeNote(path)
	if !ok {
		t.Fatal("expected note to be loaded")
	}
	if note.Note != "line one\nline two" {
		t.Fatalf("unexpected note text: %q", note.Note)
	}
	if !note.Pinned {
		t.Fatal("expected note to be pinned")
	}
}

func TestSetWorktreeNoteClearsEntryWhenEmptyAndUnpinned(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	path := "/tmp/worktrees/feature-a"
	m.setWorktreeNote(path, "keep me", true)
	m.setWorktreeNote(path, "   ", false)

	if _, ok := m.getWorktreeNote(path); ok {
		t.Fatal("expected note to be deleted")
	}
}

func TestMigrateWorktreeNote(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	oldPath := "/tmp/worktrees/old-branch"
	newPath := "/tmp/worktrees/new-branch"
	m.setWorktreeNote(oldPath, "old note", true)
	m.migrateWorktreeNote(oldPath, newPath)

	if _, ok := m.getWorktreeNote(oldPath); ok {
		t.Fatal("expected old note to be removed")
	}
	note, ok := m.getWorktreeNote(newPath)
	if !ok {
		t.Fatal("expected note to be migrated")
	}
	if note.Note != "old note" || !note.Pinned {
		t.Fatalf("unexpected migrated note: %#v", note)
	}
}

func TestPruneStaleWorktreeNotes(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	keepPath := "/tmp/worktrees/keep"
	dropPath := "/tmp/worktrees/drop"
	m.setWorktreeNote(keepPath, "keep", false)
	m.setWorktreeNote(dropPath, "drop", false)

	m.pruneStaleWorktreeNotes([]*models.WorktreeInfo{{Path: keepPath, Branch: "keep"}})

	if _, ok := m.getWorktreeNote(dropPath); ok {
		t.Fatal("expected stale note to be pruned")
	}
	if _, ok := m.getWorktreeNote(keepPath); !ok {
		t.Fatal("expected valid note to remain")
	}
}

func TestShowAnnotateWorktreeOpensTextarea(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	m.state.view.FocusedPane = 0
	m.state.data.filteredWts = []*models.WorktreeInfo{{Path: "/tmp/wt", Branch: "feat"}}
	m.state.data.selectedIndex = 0

	cmd := m.showAnnotateWorktree()
	if cmd == nil {
		t.Fatal("expected blink command")
	}
	if !m.state.ui.screenManager.IsActive() || m.state.ui.screenManager.Type() != appscreen.TypeTextarea {
		t.Fatalf("expected textarea screen, got active=%v type=%v", m.state.ui.screenManager.IsActive(), m.state.ui.screenManager.Type())
	}
}

func TestHandleBuiltInKeyAnnotate(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	m.state.data.filteredWts = []*models.WorktreeInfo{{Path: "/tmp/wt", Branch: "feat"}}
	m.state.data.selectedIndex = 0

	m.state.view.FocusedPane = 1
	_, _ = m.handleBuiltInKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if m.state.ui.screenManager.IsActive() {
		t.Fatal("did not expect screen when pane is not worktree pane")
	}

	m.state.view.FocusedPane = 0
	_, _ = m.handleBuiltInKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if !m.state.ui.screenManager.IsActive() || m.state.ui.screenManager.Type() != appscreen.TypeTextarea {
		t.Fatalf("expected textarea screen, got active=%v type=%v", m.state.ui.screenManager.IsActive(), m.state.ui.screenManager.Type())
	}
}

func TestAnnotateWorktreeCtrlSSaves(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	path := "/tmp/wt"

	m.state.view.FocusedPane = 0
	m.state.data.filteredWts = []*models.WorktreeInfo{{Path: path, Branch: "feat"}}
	m.state.data.selectedIndex = 0

	_ = m.showAnnotateWorktree()
	scr := m.state.ui.screenManager.Current().(*appscreen.TextareaScreen)
	scr.Input.SetValue("one line\ntwo line")

	// Toggle pinned via checkbox.
	_, _ = m.handleScreenKey(tea.KeyMsg{Type: tea.KeyTab})
	_, _ = m.handleScreenKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	_, _ = m.handleScreenKey(tea.KeyMsg{Type: tea.KeyTab})

	updated, _ := m.handleScreenKey(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(*Model)

	if m.state.ui.screenManager.IsActive() {
		t.Fatal("expected annotate modal to close on save")
	}
	note, ok := m.getWorktreeNote(path)
	if !ok {
		t.Fatal("expected note to be saved")
	}
	if note.Note != "one line\ntwo line" {
		t.Fatalf("unexpected saved note: %q", note.Note)
	}
	if !note.Pinned {
		t.Fatal("expected note to be pinned")
	}
}

func TestUpdateRenameWorktreeResultMigratesNote(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	oldPath := "/tmp/wt-old"
	newPath := "/tmp/wt-new"

	m.state.data.worktrees = []*models.WorktreeInfo{{Path: oldPath, Branch: "old"}}
	m.setWorktreeNote(oldPath, "rename me", true)

	updated, _ := m.Update(renameWorktreeResultMsg{
		oldPath: oldPath,
		newPath: newPath,
		worktrees: []*models.WorktreeInfo{
			{Path: newPath, Branch: "new"},
		},
	})
	m = updated.(*Model)

	if _, ok := m.getWorktreeNote(oldPath); ok {
		t.Fatal("expected old path note to be removed")
	}
	if _, ok := m.getWorktreeNote(newPath); !ok {
		t.Fatal("expected note to move to new path")
	}
}
