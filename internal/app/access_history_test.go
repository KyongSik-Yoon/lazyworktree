package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chmouel/lazyworktree/internal/config"
	"github.com/chmouel/lazyworktree/internal/models"
)

const testAccessWorktreePath = "/home/user/worktrees/feature-1"

func TestRecordAccess(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	// Record access to a worktree path
	m.recordAccess(testAccessWorktreePath)

	// Verify access was recorded
	ts, ok := m.state.data.accessHistory[testAccessWorktreePath]
	if !ok {
		t.Fatalf("expected access to be recorded for path %s", testAccessWorktreePath)
	}

	// Verify timestamp is recent (within last second)
	now := time.Now().Unix()
	if ts > now || ts < now-1 {
		t.Fatalf("expected timestamp to be recent, got %d (now: %d)", ts, now)
	}
}

func TestRecordAccessEmptyPath(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	// Record access with empty path - should be ignored
	m.recordAccess("")
	if len(m.state.data.accessHistory) != 0 {
		t.Fatalf("expected empty path to be ignored, got %d entries", len(m.state.data.accessHistory))
	}
}

func TestLoadAccessHistoryEmptyFile(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	// Load when no history file exists
	m.loadAccessHistory()
	if len(m.state.data.accessHistory) != 0 {
		t.Fatalf("expected empty history, got %d entries", len(m.state.data.accessHistory))
	}
}

func TestSaveAndLoadAccessHistory(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	// Record some accesses
	m.recordAccess(testAccessWorktreePath)
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	m.recordAccess("/home/user/worktrees/feature-2")

	// Create a new model and load history
	m2 := NewModel(cfg, "")
	m2.repoKey = testRepoKey
	m2.loadAccessHistory()

	if len(m2.state.data.accessHistory) != 2 {
		t.Fatalf("expected 2 entries after load, got %d", len(m2.state.data.accessHistory))
	}

	// Verify both paths are present
	if _, ok := m2.state.data.accessHistory[testAccessWorktreePath]; !ok {
		t.Fatal("expected feature-1 to be in loaded history")
	}
	if _, ok := m2.state.data.accessHistory["/home/user/worktrees/feature-2"]; !ok {
		t.Fatal("expected feature-2 to be in loaded history")
	}
}

func TestAccessHistoryFilePath(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	m.recordAccess("/home/user/worktrees/test-path")

	// Verify file was created at correct path
	expectedPath := filepath.Join(cfg.WorktreeDir, testRepoKey, models.AccessHistoryFilename)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("expected access history file to exist at %s", expectedPath)
	}

	// Verify file contains correct JSON
	// #nosec G304 -- expectedPath is constructed from vetted test directory
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatal(err)
	}

	var history map[string]int64
	if err := json.Unmarshal(data, &history); err != nil {
		t.Fatalf("failed to parse history file: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("expected 1 entry in file, got %d", len(history))
	}
	if _, ok := history["/home/user/worktrees/test-path"]; !ok {
		t.Fatal("expected test-path in history file")
	}
}

func TestLoadAccessHistoryInvalidJSON(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	// Create invalid JSON file
	historyPath := filepath.Join(cfg.WorktreeDir, testRepoKey, models.AccessHistoryFilename)
	if err := os.MkdirAll(filepath.Dir(historyPath), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(historyPath, []byte("invalid json"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Should not crash and should return empty history
	m.loadAccessHistory()
	if len(m.state.data.accessHistory) != 0 {
		t.Fatalf("expected empty history on invalid JSON, got %d entries", len(m.state.data.accessHistory))
	}
}

func TestRecordAccessUpdatesExisting(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	// Record first access
	m.recordAccess(testAccessWorktreePath)
	firstTS := m.state.data.accessHistory[testAccessWorktreePath]

	// Wait a bit and record again
	time.Sleep(50 * time.Millisecond)
	m.recordAccess(testAccessWorktreePath)
	secondTS := m.state.data.accessHistory[testAccessWorktreePath]

	// Second timestamp should be greater or equal
	if secondTS < firstTS {
		t.Fatalf("expected second timestamp (%d) to be >= first (%d)", secondTS, firstTS)
	}
}

func TestSortByLastSwitched(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
		SortMode:    "switched",
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	// Create worktrees with different LastSwitchedTS
	now := time.Now().Unix()
	m.state.data.worktrees = []*models.WorktreeInfo{
		{Path: "/worktrees/a", Branch: "a", LastSwitchedTS: now - 100},
		{Path: "/worktrees/b", Branch: "b", LastSwitchedTS: now - 50},
		{Path: "/worktrees/c", Branch: "c", LastSwitchedTS: now},
	}

	// Verify sortMode is set correctly
	if m.sortMode != sortModeLastSwitched {
		t.Fatalf("expected sortMode to be %d, got %d", sortModeLastSwitched, m.sortMode)
	}

	// Update table (which triggers sorting)
	m.updateTable()

	// After sorting by LastSwitchedTS (descending), order should be: c, b, a
	if len(m.state.data.filteredWts) != 3 {
		t.Fatalf("expected 3 filtered worktrees, got %d", len(m.state.data.filteredWts))
	}
	if m.state.data.filteredWts[0].Branch != "c" {
		t.Fatalf("expected first worktree to be 'c', got %q", m.state.data.filteredWts[0].Branch)
	}
	if m.state.data.filteredWts[1].Branch != "b" {
		t.Fatalf("expected second worktree to be 'b', got %q", m.state.data.filteredWts[1].Branch)
	}
	if m.state.data.filteredWts[2].Branch != "a" {
		t.Fatalf("expected third worktree to be 'a', got %q", m.state.data.filteredWts[2].Branch)
	}
}

func TestSortModeCycling(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
		SortMode:    "path",
	}
	m := NewModel(cfg, "")

	// Starting mode should be path (0)
	if m.sortMode != sortModePath {
		t.Fatalf("expected sortMode to start at %d, got %d", sortModePath, m.sortMode)
	}

	// Cycle to next mode (active)
	m.sortMode = (m.sortMode + 1) % 3
	if m.sortMode != sortModeLastActive {
		t.Fatalf("expected sortMode to be %d after first cycle, got %d", sortModeLastActive, m.sortMode)
	}

	// Cycle to next mode (switched)
	m.sortMode = (m.sortMode + 1) % 3
	if m.sortMode != sortModeLastSwitched {
		t.Fatalf("expected sortMode to be %d after second cycle, got %d", sortModeLastSwitched, m.sortMode)
	}

	// Cycle back to path
	m.sortMode = (m.sortMode + 1) % 3
	if m.sortMode != sortModePath {
		t.Fatalf("expected sortMode to be %d after third cycle, got %d", sortModePath, m.sortMode)
	}
}

func TestPersistLastSelectedRecordsAccess(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	m.persistLastSelected(testAccessWorktreePath)

	// Verify access was recorded
	ts, ok := m.state.data.accessHistory[testAccessWorktreePath]
	if !ok {
		t.Fatalf("expected persistLastSelected to record access for path %s", testAccessWorktreePath)
	}

	// Verify timestamp is recent
	now := time.Now().Unix()
	if ts > now || ts < now-1 {
		t.Fatalf("expected timestamp to be recent, got %d (now: %d)", ts, now)
	}
}
