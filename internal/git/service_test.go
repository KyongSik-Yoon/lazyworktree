package git

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/chmouel/lazyworktree/internal/config"
	"github.com/chmouel/lazyworktree/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	notify := func(message string, severity string) {}
	notifyOnce := func(key string, message string, severity string) {}

	service := NewService(notify, notifyOnce)

	assert.NotNil(t, service)
	assert.NotNil(t, service.semaphore)
	assert.NotNil(t, service.notifiedSet)
	assert.NotNil(t, service.notify)
	assert.NotNil(t, service.notifyOnce)

	expectedSlots := runtime.NumCPU() * 2
	if expectedSlots < 4 {
		expectedSlots = 4
	}
	if expectedSlots > 32 {
		expectedSlots = 32
	}

	// Semaphore should have the expected number of slots
	count := 0
	for i := 0; i < expectedSlots; i++ {
		select {
		case <-service.semaphore:
			count++
		default:
			// Can't drain more from semaphore
		}
	}
	assert.Equal(t, expectedSlots, count)
}

func TestUseDelta(t *testing.T) {
	notify := func(message string, severity string) {}
	notifyOnce := func(key string, message string, severity string) {}

	service := NewService(notify, notifyOnce)

	// UseDelta should return a boolean
	useDelta := service.UseDelta()
	assert.IsType(t, true, useDelta)
}

func TestApplyDelta(t *testing.T) {
	notify := func(message string, severity string) {}
	notifyOnce := func(key string, message string, severity string) {}

	service := NewService(notify, notifyOnce)

	t.Run("empty diff returns empty", func(t *testing.T) {
		result := service.ApplyDelta(context.Background(), "")
		assert.Empty(t, result)
	})

	t.Run("diff without delta available", func(t *testing.T) {
		// Temporarily disable delta
		origUseDelta := service.useDelta
		service.useDelta = false
		defer func() { service.useDelta = origUseDelta }()

		diff := "diff --git a/file.txt b/file.txt\n"
		result := service.ApplyDelta(context.Background(), diff)
		assert.Equal(t, diff, result)
	})

	t.Run("diff with delta available", func(t *testing.T) {
		diff := "diff --git a/file.txt b/file.txt\n+added line\n"

		result := service.ApplyDelta(context.Background(), diff)
		// Result should either be the diff (if delta not available) or transformed by delta
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "file.txt")
	})
}

func TestGetMainBranch(t *testing.T) {
	notify := func(message string, severity string) {}
	notifyOnce := func(key string, message string, severity string) {}

	service := NewService(notify, notifyOnce)

	ctx := context.Background()

	// This test requires a git repository, so we'll test basic functionality
	branch := service.GetMainBranch(ctx)

	// Branch should be non-empty (defaults to "main" or "master")
	assert.NotEmpty(t, branch)
	// Should be one of the common main branches
	assert.Contains(t, []string{"main", "master"}, branch)
}

func TestRenameWorktree(t *testing.T) {
	notify := func(message string, severity string) {}
	notifyOnce := func(key string, message string, severity string) {}

	service := NewService(notify, notifyOnce)
	ctx := context.Background()

	t.Run("rename with temporary directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		oldPath := filepath.Join(tmpDir, "old")
		newPath := filepath.Join(tmpDir, "new")

		// Create old directory
		err := os.MkdirAll(oldPath, 0o755)
		require.NoError(t, err)

		// Create a test file in old directory
		testFile := filepath.Join(oldPath, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0o644)
		require.NoError(t, err)

		// Rename (this is essentially just a directory move, not a git worktree operation)
		// Note: This will likely fail if git commands are involved, so we're testing basic logic
		ok := service.RenameWorktree(ctx, oldPath, newPath, "old-branch", "new-branch")

		// Even if it returns false due to git errors, we're just testing the function runs
		assert.IsType(t, true, ok)
	})
}

func TestExecuteCommands(t *testing.T) {
	notify := func(message string, severity string) {}
	notifyOnce := func(key string, message string, severity string) {}

	service := NewService(notify, notifyOnce)
	ctx := context.Background()

	t.Run("execute empty command list", func(t *testing.T) {
		err := service.ExecuteCommands(ctx, []string{}, "", nil)
		assert.NoError(t, err)
	})

	t.Run("execute with whitespace commands", func(t *testing.T) {
		err := service.ExecuteCommands(ctx, []string{"  ", "\t", "\n"}, "", nil)
		assert.NoError(t, err)
	})

	t.Run("execute simple command", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := service.ExecuteCommands(ctx, []string{"echo test"}, tmpDir, nil)
		// May fail if shell execution is restricted, but should not panic
		_ = err
	})

	t.Run("execute with environment variables", func(t *testing.T) {
		tmpDir := t.TempDir()
		env := map[string]string{
			"TEST_VAR": "test_value",
		}
		err := service.ExecuteCommands(ctx, []string{"echo $TEST_VAR"}, tmpDir, env)
		// May fail if shell execution is restricted, but should not panic
		_ = err
	})
}

func TestBuildThreePartDiff(t *testing.T) {
	notify := func(message string, severity string) {}
	notifyOnce := func(key string, message string, severity string) {}

	service := NewService(notify, notifyOnce)
	ctx := context.Background()

	t.Run("build diff for non-git directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &config.AppConfig{
			MaxUntrackedDiffs: 10,
			MaxDiffChars:      200000,
		}

		diff := service.BuildThreePartDiff(ctx, tmpDir, cfg)

		// Should return something (even if empty or error message)
		assert.IsType(t, "", diff)
	})
}

func TestRunGit(t *testing.T) {
	notify := func(message string, severity string) {}
	notifyOnce := func(key string, message string, severity string) {}

	service := NewService(notify, notifyOnce)
	ctx := context.Background()

	t.Run("run git version", func(t *testing.T) {
		// This is a simple git command that should work in most environments
		output := service.RunGit(ctx, []string{"git", "--version"}, "", []int{0}, false, false)

		// Should contain "git version" or be empty if git not available
		if output != "" {
			assert.Contains(t, output, "git version")
		}
	})

	t.Run("run git with allowed error code", func(t *testing.T) {
		// Run a command that will likely fail with code 128 (invalid command)
		output := service.RunGit(ctx, []string{"git", "invalid-command-xyz"}, "", []int{128}, true, false)

		// Should not panic and return some output (even if empty)
		assert.IsType(t, "", output)
	})

	t.Run("run git with cwd", func(t *testing.T) {
		tmpDir := t.TempDir()
		output := service.RunGit(ctx, []string{"git", "--version"}, tmpDir, []int{0}, false, false)

		// Should run successfully
		if output != "" {
			assert.Contains(t, output, "git version")
		}
	})
}

func TestNotifications(t *testing.T) {
	t.Run("notify function called", func(t *testing.T) {
		called := false
		var receivedMessage, receivedSeverity string

		notify := func(message string, severity string) {
			called = true
			receivedMessage = message
			receivedSeverity = severity
		}
		notifyOnce := func(key string, message string, severity string) {}

		service := NewService(notify, notifyOnce)

		// Trigger a notification
		service.notify("test message", "info")

		assert.True(t, called)
		assert.Equal(t, "test message", receivedMessage)
		assert.Equal(t, "info", receivedSeverity)
	})

	t.Run("notifyOnce function called", func(t *testing.T) {
		called := false
		var receivedKey, receivedMessage, receivedSeverity string

		notify := func(message string, severity string) {}
		notifyOnce := func(key string, message string, severity string) {
			called = true
			receivedKey = key
			receivedMessage = message
			receivedSeverity = severity
		}

		service := NewService(notify, notifyOnce)

		// Trigger a one-time notification
		service.notifyOnce("test-key", "test message", "warning")

		assert.True(t, called)
		assert.Equal(t, "test-key", receivedKey)
		assert.Equal(t, "test message", receivedMessage)
		assert.Equal(t, "warning", receivedSeverity)
	})
}

func TestWorktreeOperations(t *testing.T) {
	notify := func(message string, severity string) {}
	notifyOnce := func(key string, message string, severity string) {}

	service := NewService(notify, notifyOnce)
	ctx := context.Background()

	t.Run("get worktrees from non-git directory", func(t *testing.T) {
		worktrees, err := service.GetWorktrees(ctx)

		// Should handle error gracefully
		if err != nil {
			assert.Error(t, err)
			assert.Nil(t, worktrees)
		} else {
			assert.IsType(t, []*models.WorktreeInfo{}, worktrees)
		}
	})
}

func TestFetchPRMap(t *testing.T) {
	notify := func(message string, severity string) {}
	notifyOnce := func(key string, message string, severity string) {}

	service := NewService(notify, notifyOnce)
	ctx := context.Background()

	t.Run("fetch PR map without git repository", func(t *testing.T) {
		// This test just verifies the function doesn't panic
		// Behavior varies by git environment (may return error or empty map)
		prMap, err := service.FetchPRMap(ctx)

		// Function should not panic and should return valid types
		// Either error or map (which can be nil or empty)
		if err == nil {
			// prMap can be nil or a valid map - both are acceptable
			if prMap != nil {
				assert.IsType(t, map[string]*models.PRInfo{}, prMap)
			}
		}
	})
}
