package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/chmouel/lazyworktree/internal/config"
	"github.com/chmouel/lazyworktree/internal/models"
)

const (
	testFeat = "feat"
	testWt1  = "wt1"
	testWt2  = "wt2"
)

func TestHandlePageDownUpOnStatusPane(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.focusedPane = 1
	m.statusViewport = viewport.New(10, 2)
	m.statusViewport.SetContent(strings.Repeat("line\n", 10))

	start := m.statusViewport.YOffset
	_, _ = m.handlePageDown(tea.KeyMsg{Type: tea.KeyPgDown})
	if m.statusViewport.YOffset <= start {
		t.Fatalf("expected YOffset to increase, got %d", m.statusViewport.YOffset)
	}

	m.statusViewport.YOffset = 2
	_, _ = m.handlePageUp(tea.KeyMsg{Type: tea.KeyPgUp})
	if m.statusViewport.YOffset >= 2 {
		t.Fatalf("expected YOffset to decrease, got %d", m.statusViewport.YOffset)
	}
}

func TestHandleEnterKeySelectsWorktree(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.focusedPane = 0
	m.filteredWts = []*models.WorktreeInfo{
		{Path: filepath.Join(cfg.WorktreeDir, "wt"), Branch: testFeat},
	}
	m.selectedIndex = 0

	_, cmd := m.handleEnterKey()
	if m.selectedPath == "" {
		t.Fatal("expected selected path to be set")
	}
	if cmd == nil {
		t.Fatal("expected command to be returned")
	}
}

func TestFilterEnterSelectsHighlightedMatch(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
		SortMode:    "path",
	}
	m := NewModel(cfg, "")
	m.focusedPane = 0

	m.worktrees = []*models.WorktreeInfo{
		{Path: filepath.Join(cfg.WorktreeDir, "b-worktree"), Branch: testFeat},
		{Path: filepath.Join(cfg.WorktreeDir, "a-worktree"), Branch: testFeat},
	}
	m.filterQuery = testFeat
	m.filterInput.SetValue(testFeat)
	m.updateTable()
	m.showingFilter = true
	m.filterInput.Focus()
	m.worktreeTable.SetCursor(1)
	m.selectedIndex = 1

	updated, cmd := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel, ok := updated.(*Model)
	if !ok {
		t.Fatalf("expected updated model, got %T", updated)
	}
	m = updatedModel

	if cmd == nil {
		t.Fatal("expected quit command to be returned")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected quit message, got %T", msg)
	}
	// Should select the item at cursor position 1 (b-worktree)
	expected := filepath.Join(cfg.WorktreeDir, "b-worktree")
	if m.selectedPath != expected {
		t.Fatalf("expected selected path %q, got %q", expected, m.selectedPath)
	}
}

func TestFilterAltNPMovesSelectionAndFills(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
		SortMode:    "path",
	}
	m := NewModel(cfg, "")

	wt1Path := filepath.Join(cfg.WorktreeDir, testWt1)
	wt2Path := filepath.Join(cfg.WorktreeDir, testWt2)
	m.worktrees = []*models.WorktreeInfo{
		{Path: wt1Path, Branch: "feat-one"},
		{Path: wt2Path, Branch: "feat-two"},
	}
	m.filterQuery = testFeat
	m.filterInput.SetValue(testFeat)
	m.updateTable()
	m.showingFilter = true
	m.filterInput.Focus()
	m.worktreeTable.SetCursor(0)
	m.selectedIndex = 0

	updated, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}, Alt: true})
	updatedModel, ok := updated.(*Model)
	if !ok {
		t.Fatalf("expected updated model, got %T", updated)
	}
	m = updatedModel

	if m.filterInput.Value() != testWt2 || m.filterQuery != testWt2 {
		t.Fatalf("expected filter query to match selected worktree, got %q", m.filterQuery)
	}
	if len(m.filteredWts) != 1 || m.filteredWts[0].Path != wt2Path {
		t.Fatalf("expected filtered worktree %q, got %v", wt2Path, m.filteredWts)
	}

	updated, _ = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}, Alt: true})
	updatedModel, ok = updated.(*Model)
	if !ok {
		t.Fatalf("expected updated model, got %T", updated)
	}
	m = updatedModel

	if m.filterInput.Value() != testWt1 || m.filterQuery != testWt1 {
		t.Fatalf("expected filter query to match selected worktree, got %q", m.filterQuery)
	}
	if len(m.filteredWts) != 1 || m.filteredWts[0].Path != wt1Path {
		t.Fatalf("expected filtered worktree %q, got %v", wt1Path, m.filteredWts)
	}
}

func TestFilterArrowKeysNavigateWithoutFilling(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
		SortMode:    "path",
	}
	m := NewModel(cfg, "")

	wt1Path := filepath.Join(cfg.WorktreeDir, testWt1)
	wt2Path := filepath.Join(cfg.WorktreeDir, testWt2)
	m.worktrees = []*models.WorktreeInfo{
		{Path: wt1Path, Branch: "feat-one"},
		{Path: wt2Path, Branch: "feat-two"},
	}
	m.filterQuery = testFeat
	m.filterInput.SetValue(testFeat)
	m.updateTable()
	m.showingFilter = true
	m.filterInput.Focus()
	m.worktreeTable.SetCursor(0)
	m.selectedIndex = 0

	updated, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyDown})
	updatedModel, ok := updated.(*Model)
	if !ok {
		t.Fatalf("expected updated model, got %T", updated)
	}
	m = updatedModel

	if m.filterInput.Value() != testFeat || m.filterQuery != testFeat {
		t.Fatalf("expected filter query unchanged, got %q", m.filterQuery)
	}

	updated, _ = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyUp})
	updatedModel, ok = updated.(*Model)
	if !ok {
		t.Fatalf("expected updated model, got %T", updated)
	}
	m = updatedModel

	if m.filterInput.Value() != testFeat || m.filterQuery != testFeat {
		t.Fatalf("expected filter query unchanged, got %q", m.filterQuery)
	}
}

func TestFilterEmptyEnterSelectsCurrent(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
		SortMode:    "path",
	}
	m := NewModel(cfg, "")

	wt1Path := filepath.Join(cfg.WorktreeDir, testWt1)
	wt2Path := filepath.Join(cfg.WorktreeDir, testWt2)
	m.worktrees = []*models.WorktreeInfo{
		{Path: wt1Path, Branch: "feat-one"},
		{Path: wt2Path, Branch: "feat-two"},
	}
	m.filterQuery = ""
	m.filterInput.SetValue("")
	m.updateTable()
	m.showingFilter = true
	m.filterInput.Focus()
	m.worktreeTable.SetCursor(1)
	m.selectedIndex = 1

	updated, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel, ok := updated.(*Model)
	if !ok {
		t.Fatalf("expected updated model, got %T", updated)
	}
	m = updatedModel

	if m.showingFilter {
		t.Fatal("expected filter to be closed")
	}
	if m.selectedIndex != 1 {
		t.Fatalf("expected selectedIndex to remain 1, got %d", m.selectedIndex)
	}
}

func TestFilterCtrlCExitsFilter(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
		SortMode:    "path",
	}
	m := NewModel(cfg, "")

	wt1Path := filepath.Join(cfg.WorktreeDir, testWt1)
	m.worktrees = []*models.WorktreeInfo{
		{Path: wt1Path, Branch: "feat-one"},
	}
	m.filterQuery = "something"
	m.filterInput.SetValue("something")
	m.updateTable()
	m.showingFilter = true
	m.filterInput.Focus()

	updated, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyCtrlC})
	updatedModel, ok := updated.(*Model)
	if !ok {
		t.Fatalf("expected updated model, got %T", updated)
	}
	m = updatedModel

	if m.showingFilter {
		t.Fatal("expected filter to be closed after Ctrl+C")
	}
	if m.filterInput.Focused() {
		t.Fatal("expected filter input to be blurred")
	}
}

func TestHandleCachedWorktreesUpdatesState(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.selectedIndex = 0
	m.worktreeTable.SetWidth(80)

	msg := cachedWorktreesMsg{
		worktrees: []*models.WorktreeInfo{
			{Path: filepath.Join(cfg.WorktreeDir, "wt1"), Branch: "main"},
		},
	}

	_, cmd := m.handleCachedWorktrees(msg)
	if cmd != nil {
		t.Fatal("expected no command")
	}
	if len(m.worktrees) != 1 {
		t.Fatalf("expected worktrees to be set, got %d", len(m.worktrees))
	}
	if m.statusContent != loadingRefreshWorktrees {
		t.Fatalf("unexpected status content: %q", m.statusContent)
	}
	if !strings.Contains(m.infoContent, "wt1") {
		t.Fatalf("expected info content to include worktree path, got %q", m.infoContent)
	}
}

func TestHandlePRDataLoadedUpdatesTable(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.worktreeTable.SetWidth(100)
	m.worktreesLoaded = true
	m.worktrees = []*models.WorktreeInfo{
		{Path: filepath.Join(cfg.WorktreeDir, "wt1"), Branch: "feature"},
	}
	m.filteredWts = m.worktrees
	m.worktreeTable.SetCursor(0)

	msg := prDataLoadedMsg{
		prMap: map[string]*models.PRInfo{
			"feature": {Number: 12, State: "OPEN", Title: "Test PR", URL: "https://example.com"},
		},
	}

	_, cmd := m.handlePRDataLoaded(msg)
	if cmd == nil {
		t.Fatal("expected command to be returned")
	}
	if !m.prDataLoaded {
		t.Fatal("expected prDataLoaded to be true")
	}
	if m.worktrees[0].PR == nil {
		t.Fatal("expected PR info to be applied to worktree")
	}
	if len(m.worktreeTable.Columns()) != 5 {
		t.Fatalf("expected 5 columns after PR data, got %d", len(m.worktreeTable.Columns()))
	}
}

func TestHandlePRDataLoadedWithWorktreePRs(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.worktreeTable.SetWidth(100)
	m.worktreesLoaded = true
	wtPath := filepath.Join(cfg.WorktreeDir, "wt1")
	m.worktrees = []*models.WorktreeInfo{
		{Path: wtPath, Branch: "local-branch-name"},
	}
	m.filteredWts = m.worktrees
	m.worktreeTable.SetCursor(0)

	// Simulate a case where the local branch name differs from the PR's headRefName
	// So prMap won't match, but worktreePRs (from gh pr view) will
	msg := prDataLoadedMsg{
		prMap: map[string]*models.PRInfo{
			"remote-branch-name": {Number: 99, State: "OPEN", Title: "Fork PR", URL: "https://example.com"},
		},
		worktreePRs: map[string]*models.PRInfo{
			wtPath: {Number: 99, State: "OPEN", Title: "Fork PR", URL: "https://example.com"},
		},
	}

	_, cmd := m.handlePRDataLoaded(msg)
	if cmd == nil {
		t.Fatal("expected command to be returned")
	}
	if !m.prDataLoaded {
		t.Fatal("expected prDataLoaded to be true")
	}
	if m.worktrees[0].PR == nil {
		t.Fatal("expected PR info to be applied to worktree via worktreePRs")
	}
	if m.worktrees[0].PR.Number != 99 {
		t.Fatalf("expected PR number 99, got %d", m.worktrees[0].PR.Number)
	}
}

func TestHandleCIStatusLoadedUpdatesCache(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.filteredWts = []*models.WorktreeInfo{
		{
			Path:   filepath.Join(cfg.WorktreeDir, "wt1"),
			Branch: "feature",
			PR: &models.PRInfo{
				Number: 1,
				State:  "OPEN",
				Title:  "Test",
				URL:    testPRURL,
			},
		},
	}
	m.selectedIndex = 0

	msg := ciStatusLoadedMsg{
		branch: "feature",
		checks: []*models.CICheck{
			{Name: "build", Status: "completed", Conclusion: "success"},
		},
	}

	_, cmd := m.handleCIStatusLoaded(msg)
	if cmd != nil {
		t.Fatal("expected no command")
	}
	if entry, ok := m.ciCache["feature"]; !ok || len(entry.checks) != 1 {
		t.Fatalf("expected CI cache to be updated, got %v", entry)
	}
	if !strings.Contains(m.infoContent, "CI Checks:") {
		t.Fatalf("expected info content to include CI checks, got %q", m.infoContent)
	}
}

func TestFilterEnterSelectsHighlightedItem(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir:      t.TempDir(),
		SortMode:         "path",
		SearchAutoSelect: false,
	}
	m := NewModel(cfg, "")
	m.focusedPane = 0

	wt1Path := filepath.Join(cfg.WorktreeDir, "srv-api")
	wt2Path := filepath.Join(cfg.WorktreeDir, "srv-auth")
	wt3Path := filepath.Join(cfg.WorktreeDir, "srv-worker")
	m.worktrees = []*models.WorktreeInfo{
		{Path: wt1Path, Branch: "feature/srv-api"},
		{Path: wt2Path, Branch: "feature/srv-auth"},
		{Path: wt3Path, Branch: "feature/srv-worker"},
	}

	// Apply filter for "srv"
	m.filterQuery = "srv"
	m.filterInput.SetValue("srv")
	m.updateTable()
	m.showingFilter = true
	m.filterInput.Focus()

	// Navigate to the second item (srv-auth)
	m.worktreeTable.SetCursor(1)
	m.selectedIndex = 1

	// Press Enter - should select the highlighted item (srv-auth)
	updated, cmd := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel, ok := updated.(*Model)
	if !ok {
		t.Fatalf("expected updated model, got %T", updated)
	}
	m = updatedModel

	if cmd == nil {
		t.Fatal("expected quit command to be returned")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected quit message, got %T", msg)
	}
	if m.selectedPath != wt2Path {
		t.Fatalf("expected selected path %q, got %q", wt2Path, m.selectedPath)
	}
}

func TestFilterNavigationThroughMultipleFilteredItems(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
		SortMode:    "path",
	}
	m := NewModel(cfg, "")

	// Create 5 worktrees, 3 of which match "srv" filter
	wt1Path := filepath.Join(cfg.WorktreeDir, "main")
	wt2Path := filepath.Join(cfg.WorktreeDir, "srv-api")
	wt3Path := filepath.Join(cfg.WorktreeDir, "frontend")
	wt4Path := filepath.Join(cfg.WorktreeDir, "srv-auth")
	wt5Path := filepath.Join(cfg.WorktreeDir, "srv-worker")

	m.worktrees = []*models.WorktreeInfo{
		{Path: wt1Path, Branch: "main", IsMain: true},
		{Path: wt2Path, Branch: "feature/srv-api"},
		{Path: wt3Path, Branch: "feature/frontend"},
		{Path: wt4Path, Branch: "feature/srv-auth"},
		{Path: wt5Path, Branch: "feature/srv-worker"},
	}

	// Apply filter for "srv"
	m.filterQuery = "srv"
	m.filterInput.SetValue("srv")
	m.updateTable()
	m.showingFilter = true
	m.filterInput.Focus()
	m.worktreeTable.SetCursor(0)
	m.selectedIndex = 0

	// Verify we have exactly 3 filtered items
	if len(m.filteredWts) != 3 {
		t.Fatalf("expected 3 filtered items, got %d", len(m.filteredWts))
	}

	// Navigate down through all filtered items
	for i := 0; i < 2; i++ {
		updated, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyDown})
		updatedModel, ok := updated.(*Model)
		if !ok {
			t.Fatalf("expected updated model, got %T", updated)
		}
		m = updatedModel
	}

	// Should be at the last filtered item (index 2)
	cursor := m.worktreeTable.Cursor()
	if cursor != 2 {
		t.Fatalf("expected cursor at index 2, got %d", cursor)
	}

	// Try to navigate down again - should stay at last item
	updated, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyDown})
	updatedModel, ok := updated.(*Model)
	if !ok {
		t.Fatalf("expected updated model, got %T", updated)
	}
	m = updatedModel

	cursor = m.worktreeTable.Cursor()
	if cursor != 2 {
		t.Fatalf("expected cursor to stay at index 2, got %d", cursor)
	}

	// Navigate back up
	for i := 0; i < 2; i++ {
		updated, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyUp})
		updatedModel, ok := updated.(*Model)
		if !ok {
			t.Fatalf("expected updated model, got %T", updated)
		}
		m = updatedModel
	}

	// Should be at the first filtered item (index 0)
	cursor = m.worktreeTable.Cursor()
	if cursor != 0 {
		t.Fatalf("expected cursor at index 0, got %d", cursor)
	}

	// Try to navigate up again - should stay at first item
	updated, _ = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyUp})
	updatedModel, ok = updated.(*Model)
	if !ok {
		t.Fatalf("expected updated model, got %T", updated)
	}
	m = updatedModel

	cursor = m.worktreeTable.Cursor()
	if cursor != 0 {
		t.Fatalf("expected cursor to stay at index 0, got %d", cursor)
	}
}

// TestStatusFileNavigation tests j/k navigation through status files in pane 1.
func TestStatusFileNavigation(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.focusedPane = 1
	m.statusViewport = viewport.New(40, 10)

	// Set up status files
	m.statusFiles = []StatusFile{
		{Filename: "file1.go", Status: ".M", IsUntracked: false},
		{Filename: "file2.go", Status: "M.", IsUntracked: false},
		{Filename: "file3.go", Status: " ?", IsUntracked: true},
	}
	m.statusFileIndex = 0

	// Test navigation down with j
	_, _ = m.handleNavigationDown(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.statusFileIndex != 1 {
		t.Fatalf("expected statusFileIndex 1 after j, got %d", m.statusFileIndex)
	}

	// Test navigation down again
	_, _ = m.handleNavigationDown(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.statusFileIndex != 2 {
		t.Fatalf("expected statusFileIndex 2 after second j, got %d", m.statusFileIndex)
	}

	// Test boundary - should not go past last item
	_, _ = m.handleNavigationDown(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.statusFileIndex != 2 {
		t.Fatalf("expected statusFileIndex to stay at 2, got %d", m.statusFileIndex)
	}

	// Test navigation up with k
	_, _ = m.handleNavigationUp(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.statusFileIndex != 1 {
		t.Fatalf("expected statusFileIndex 1 after k, got %d", m.statusFileIndex)
	}

	// Navigate to first item
	_, _ = m.handleNavigationUp(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.statusFileIndex != 0 {
		t.Fatalf("expected statusFileIndex 0 after second k, got %d", m.statusFileIndex)
	}

	// Test boundary - should not go below 0
	_, _ = m.handleNavigationUp(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.statusFileIndex != 0 {
		t.Fatalf("expected statusFileIndex to stay at 0, got %d", m.statusFileIndex)
	}
}

func TestLogPaneCtrlJMovesNextCommit(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.focusedPane = 2
	m.logTable.Focus()
	m.filteredWts = []*models.WorktreeInfo{
		{Path: t.TempDir(), Branch: testFeat},
	}
	m.selectedIndex = 0
	m.logEntries = []commitLogEntry{
		{sha: "abc123", message: "first"},
		{sha: "def456", message: "second"},
	}
	m.logTable.SetRows([]table.Row{
		{"abc123", "first"},
		{"def456", "second"},
	})
	m.logTable.SetCursor(0)

	execCalled := false
	m.execProcess = func(_ *exec.Cmd, cb tea.ExecCallback) tea.Cmd {
		return func() tea.Msg {
			execCalled = true
			return cb(nil)
		}
	}

	updated, cmd := m.handleBuiltInKey(tea.KeyMsg{Type: tea.KeyCtrlJ})
	updatedModel, ok := updated.(*Model)
	if !ok {
		t.Fatalf("expected updated model, got %T", updated)
	}
	m = updatedModel

	if m.logTable.Cursor() != 1 {
		t.Fatalf("expected log cursor at 1, got %d", m.logTable.Cursor())
	}
	if cmd == nil {
		t.Fatal("expected command to be returned")
	}
	_ = cmd()
	if !execCalled {
		t.Fatal("expected commit view to be opened")
	}
}

// TestStatusFileNavigationEmptyList tests navigation with no status files.
func TestStatusFileNavigationEmptyList(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.focusedPane = 1
	m.statusViewport = viewport.New(40, 10)
	m.statusFiles = nil
	m.statusFileIndex = 0

	// Should not panic with empty list
	_, _ = m.handleNavigationDown(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.statusFileIndex != 0 {
		t.Fatalf("expected statusFileIndex to stay at 0, got %d", m.statusFileIndex)
	}

	_, _ = m.handleNavigationUp(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.statusFileIndex != 0 {
		t.Fatalf("expected statusFileIndex to stay at 0, got %d", m.statusFileIndex)
	}
}

// TestStatusFileEnterShowsDiff tests that Enter on pane 1 triggers showFileDiff.
func TestStatusFileEnterShowsDiff(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.focusedPane = 1
	m.statusViewport = viewport.New(40, 10)

	// Set up worktree and status files
	m.filteredWts = []*models.WorktreeInfo{
		{Path: filepath.Join(cfg.WorktreeDir, "wt1"), Branch: "feature"},
	}
	m.selectedIndex = 0
	m.statusFiles = []StatusFile{
		{Filename: "file1.go", Status: ".M", IsUntracked: false},
		{Filename: "file2.go", Status: "M.", IsUntracked: false},
	}
	m.statusFileIndex = 1

	// Mock execProcess to capture the command
	var capturedCmd bool
	m.execProcess = func(_ *exec.Cmd, cb tea.ExecCallback) tea.Cmd {
		capturedCmd = true
		return func() tea.Msg { return cb(nil) }
	}

	_, cmd := m.handleEnterKey()
	if cmd == nil {
		t.Fatal("expected command to be returned")
	}

	// Execute the command
	_ = cmd()

	if !capturedCmd {
		t.Fatal("expected execProcess to be called")
	}
}

func TestStatusFileEditOpensEditor(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
		Editor:      "nvim",
	}
	m := NewModel(cfg, "")
	m.focusedPane = 1
	m.statusViewport = viewport.New(40, 10)

	wtPath := filepath.Join(cfg.WorktreeDir, "wt1")
	if err := os.MkdirAll(wtPath, 0o700); err != nil {
		t.Fatalf("failed to create worktree dir: %v", err)
	}
	filename := "file1.go"
	if err := os.WriteFile(filepath.Join(wtPath, filename), []byte("package main\n"), 0o600); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	m.filteredWts = []*models.WorktreeInfo{
		{Path: wtPath, Branch: "feature"},
	}
	m.selectedIndex = 0
	m.statusFiles = []StatusFile{
		{Filename: filename, Status: ".M", IsUntracked: false},
	}
	m.statusFileIndex = 0

	var gotCmd *exec.Cmd
	m.execProcess = func(cmd *exec.Cmd, cb tea.ExecCallback) tea.Cmd {
		gotCmd = cmd
		return func() tea.Msg { return cb(nil) }
	}

	_, cmd := m.handleBuiltInKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if cmd == nil {
		t.Fatal("expected command to be returned")
	}
	_ = cmd()

	if gotCmd == nil {
		t.Fatal("expected execProcess to be called")
	}
	if gotCmd.Dir != wtPath {
		t.Fatalf("expected worktree dir %q, got %q", wtPath, gotCmd.Dir)
	}
	if len(gotCmd.Args) < 3 || gotCmd.Args[0] != "bash" || gotCmd.Args[1] != "-c" {
		t.Fatalf("expected bash -c command, got %v", gotCmd.Args)
	}
	if !strings.Contains(gotCmd.Args[2], "nvim") || !strings.Contains(gotCmd.Args[2], filename) {
		t.Fatalf("expected editor command to include nvim and file, got %q", gotCmd.Args[2])
	}
}

// TestStatusFileEnterNoFilesDoesNothing tests Enter with no status files.
func TestStatusFileEnterNoFilesDoesNothing(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.focusedPane = 1
	m.statusFiles = nil

	_, cmd := m.handleEnterKey()
	if cmd != nil {
		t.Fatal("expected no command when no status files")
	}
}

// TestBuildStatusContentParsesFiles tests that buildStatusContent parses git status correctly.
func TestBuildStatusContentParsesFiles(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.focusedPane = 1
	m.statusViewport = viewport.New(40, 10)

	// Simulated git status --porcelain=v2 output
	statusRaw := `1 .M N... 100644 100644 100644 abc123 abc123 modified.go
1 M. N... 100644 100644 100644 def456 def456 staged.go
? untracked.txt
1 A. N... 100644 100644 100644 ghi789 ghi789 added.go
1 .D N... 100644 100644 100644 jkl012 jkl012 deleted.go`

	_ = m.buildStatusContent(statusRaw)

	if len(m.statusFiles) != 5 {
		t.Fatalf("expected 5 status files, got %d", len(m.statusFiles))
	}

	// Check first file (modified)
	if m.statusFiles[0].Filename != "modified.go" {
		t.Fatalf("expected filename 'modified.go', got %q", m.statusFiles[0].Filename)
	}
	if m.statusFiles[0].Status != ".M" {
		t.Fatalf("expected status '.M', got %q", m.statusFiles[0].Status)
	}
	if m.statusFiles[0].IsUntracked {
		t.Fatal("expected IsUntracked to be false for modified file")
	}

	// Check untracked file
	if m.statusFiles[2].Filename != "untracked.txt" {
		t.Fatalf("expected filename 'untracked.txt', got %q", m.statusFiles[2].Filename)
	}
	if !m.statusFiles[2].IsUntracked {
		t.Fatal("expected IsUntracked to be true for untracked file")
	}
}

// TestBuildStatusContentCleanTree tests that clean working tree is handled.
func TestBuildStatusContentCleanTree(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.focusedPane = 1
	m.statusViewport = viewport.New(40, 10)
	m.statusFiles = []StatusFile{{Filename: "old.go", Status: ".M"}}
	m.statusFileIndex = 5

	result := m.buildStatusContent("")

	if len(m.statusFiles) != 0 {
		t.Fatalf("expected 0 status files for clean tree, got %d", len(m.statusFiles))
	}
	if m.statusFileIndex != 0 {
		t.Fatalf("expected statusFileIndex reset to 0, got %d", m.statusFileIndex)
	}
	if !strings.Contains(result, "Clean working tree") {
		t.Fatalf("expected 'Clean working tree' in result, got %q", result)
	}
}

// TestRenderStatusFilesHighlighting tests that selected file is highlighted.
func TestRenderStatusFilesHighlighting(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.focusedPane = 1
	m.statusViewport = viewport.New(40, 10)
	m.statusFiles = []StatusFile{
		{Filename: "file1.go", Status: ".M", IsUntracked: false},
		{Filename: "file2.go", Status: ".M", IsUntracked: false},
	}
	m.statusFileIndex = 1

	result := m.renderStatusFiles()

	// The result should contain both filenames
	if !strings.Contains(result, "file1.go") {
		t.Fatalf("expected result to contain 'file1.go', got %q", result)
	}
	if !strings.Contains(result, "file2.go") {
		t.Fatalf("expected result to contain 'file2.go', got %q", result)
	}

	// Result should have multiple lines
	lines := strings.Split(result, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}

// TestStatusFileIndexClamping tests that statusFileIndex is clamped to valid range.
func TestStatusFileIndexClamping(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.focusedPane = 1
	m.statusViewport = viewport.New(40, 10)

	// Set index out of range before parsing
	m.statusFileIndex = 100

	statusRaw := `1 .M N... 100644 100644 100644 abc123 abc123 file1.go
1 .M N... 100644 100644 100644 abc123 abc123 file2.go`

	_ = m.buildStatusContent(statusRaw)

	// Index should be clamped to last valid index
	if m.statusFileIndex != 1 {
		t.Fatalf("expected statusFileIndex clamped to 1, got %d", m.statusFileIndex)
	}

	// Test negative index
	m.statusFileIndex = -5
	_ = m.buildStatusContent(statusRaw)

	if m.statusFileIndex != 0 {
		t.Fatalf("expected statusFileIndex clamped to 0, got %d", m.statusFileIndex)
	}
}

// TestMouseScrollNavigatesFiles tests that mouse scroll navigates files in pane 1.
func TestMouseScrollNavigatesFiles(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.focusedPane = 1
	m.statusViewport = viewport.New(40, 10)
	m.windowWidth = 100
	m.windowHeight = 30

	m.statusFiles = []StatusFile{
		{Filename: "file1.go", Status: ".M", IsUntracked: false},
		{Filename: "file2.go", Status: ".M", IsUntracked: false},
		{Filename: "file3.go", Status: ".M", IsUntracked: false},
	}
	m.statusFileIndex = 0

	// Scroll down should increment index
	msg := tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelDown,
		X:      60, // Right side of screen (pane 1)
		Y:      5,
	}

	_, _ = m.handleMouse(msg)
	if m.statusFileIndex != 1 {
		t.Fatalf("expected statusFileIndex 1 after scroll down, got %d", m.statusFileIndex)
	}

	// Scroll up should decrement index
	msg.Button = tea.MouseButtonWheelUp
	_, _ = m.handleMouse(msg)
	if m.statusFileIndex != 0 {
		t.Fatalf("expected statusFileIndex 0 after scroll up, got %d", m.statusFileIndex)
	}
}
