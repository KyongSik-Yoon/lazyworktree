package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSymlinkPath(t *testing.T) {
	t.Run("source does not exist", func(t *testing.T) {
		mainDir := t.TempDir()
		worktreeDir := t.TempDir()

		err := symlinkPath(mainDir, worktreeDir, "nonexistent.txt")
		assert.NoError(t, err) // Should return nil without error
	})

	t.Run("create symlink for file", func(t *testing.T) {
		mainDir := t.TempDir()
		worktreeDir := t.TempDir()

		// Create source file in main
		srcFile := filepath.Join(mainDir, "test.txt")
		err := os.WriteFile(srcFile, []byte("test content"), 0o644)
		require.NoError(t, err)

		// Create symlink
		err = symlinkPath(mainDir, worktreeDir, "test.txt")
		assert.NoError(t, err)

		// Verify symlink exists and points to correct target
		dstFile := filepath.Join(worktreeDir, "test.txt")
		linkTarget, err := os.Readlink(dstFile)
		assert.NoError(t, err)
		assert.Equal(t, srcFile, linkTarget)
	})

	t.Run("create symlink for directory", func(t *testing.T) {
		mainDir := t.TempDir()
		worktreeDir := t.TempDir()

		// Create source directory in main
		srcDir := filepath.Join(mainDir, "testdir")
		err := os.MkdirAll(srcDir, 0o755)
		require.NoError(t, err)

		// Create symlink
		err = symlinkPath(mainDir, worktreeDir, "testdir")
		assert.NoError(t, err)

		// Verify symlink exists
		dstDir := filepath.Join(worktreeDir, "testdir")
		linkTarget, err := os.Readlink(dstDir)
		assert.NoError(t, err)
		assert.Equal(t, srcDir, linkTarget)
	})

	t.Run("symlink already exists", func(t *testing.T) {
		mainDir := t.TempDir()
		worktreeDir := t.TempDir()

		// Create source file
		srcFile := filepath.Join(mainDir, "test.txt")
		err := os.WriteFile(srcFile, []byte("test"), 0o644)
		require.NoError(t, err)

		// Create symlink first time
		err = symlinkPath(mainDir, worktreeDir, "test.txt")
		assert.NoError(t, err)

		// Try creating symlink again - should return nil without error
		err = symlinkPath(mainDir, worktreeDir, "test.txt")
		assert.NoError(t, err)
	})

	t.Run("create symlink with nested path", func(t *testing.T) {
		mainDir := t.TempDir()
		worktreeDir := t.TempDir()

		// Create source file in nested directory
		srcFile := filepath.Join(mainDir, "nested", "dir", "test.txt")
		err := os.MkdirAll(filepath.Dir(srcFile), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(srcFile, []byte("nested content"), 0o644)
		require.NoError(t, err)

		// Create symlink
		err = symlinkPath(mainDir, worktreeDir, "nested/dir/test.txt")
		assert.NoError(t, err)

		// Verify symlink exists
		dstFile := filepath.Join(worktreeDir, "nested/dir/test.txt")
		linkTarget, err := os.Readlink(dstFile)
		assert.NoError(t, err)
		assert.Equal(t, srcFile, linkTarget)

		// Verify parent directory was created
		dstDir := filepath.Join(worktreeDir, "nested/dir")
		info, err := os.Stat(dstDir)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

func TestLinkTopSymlinks(t *testing.T) {
	t.Run("missing main path", func(t *testing.T) {
		worktreeDir := t.TempDir()
		statusFunc := func(ctx context.Context, path string) string {
			return ""
		}

		err := LinkTopSymlinks(context.Background(), "", worktreeDir, statusFunc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing paths")
	})

	t.Run("missing worktree path", func(t *testing.T) {
		mainDir := t.TempDir()
		statusFunc := func(ctx context.Context, path string) string {
			return ""
		}

		err := LinkTopSymlinks(context.Background(), mainDir, "", statusFunc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing paths")
	})

	t.Run("symlink untracked files", func(t *testing.T) {
		mainDir := t.TempDir()
		worktreeDir := t.TempDir()

		// Create untracked files in main
		file1 := filepath.Join(mainDir, "untracked1.txt")
		err := os.WriteFile(file1, []byte("content1"), 0o644)
		require.NoError(t, err)

		file2 := filepath.Join(mainDir, "untracked2.txt")
		err = os.WriteFile(file2, []byte("content2"), 0o644)
		require.NoError(t, err)

		statusFunc := func(ctx context.Context, path string) string {
			return "?? untracked1.txt\n?? untracked2.txt\n M tracked.txt"
		}

		err = LinkTopSymlinks(context.Background(), mainDir, worktreeDir, statusFunc)
		assert.NoError(t, err)

		// Verify symlinks were created
		link1 := filepath.Join(worktreeDir, "untracked1.txt")
		target1, err := os.Readlink(link1)
		assert.NoError(t, err)
		assert.Equal(t, file1, target1)

		link2 := filepath.Join(worktreeDir, "untracked2.txt")
		target2, err := os.Readlink(link2)
		assert.NoError(t, err)
		assert.Equal(t, file2, target2)
	})

	t.Run("symlink ignored files", func(t *testing.T) {
		mainDir := t.TempDir()
		worktreeDir := t.TempDir()

		// Create ignored file in main
		ignoredFile := filepath.Join(mainDir, "ignored.log")
		err := os.WriteFile(ignoredFile, []byte("log content"), 0o644)
		require.NoError(t, err)

		statusFunc := func(ctx context.Context, path string) string {
			return "!! ignored.log"
		}

		err = LinkTopSymlinks(context.Background(), mainDir, worktreeDir, statusFunc)
		assert.NoError(t, err)

		// Verify symlink was created
		link := filepath.Join(worktreeDir, "ignored.log")
		target, err := os.Readlink(link)
		assert.NoError(t, err)
		assert.Equal(t, ignoredFile, target)
	})

	t.Run("symlink editor configs", func(t *testing.T) {
		mainDir := t.TempDir()
		worktreeDir := t.TempDir()

		// Create editor config directories
		editorDirs := []string{".vscode", ".idea", ".cursor", ".claude"}
		for _, dir := range editorDirs {
			dirPath := filepath.Join(mainDir, dir)
			err := os.MkdirAll(dirPath, 0o755)
			require.NoError(t, err)

			// Add a file in each directory
			configFile := filepath.Join(dirPath, "settings.json")
			err = os.WriteFile(configFile, []byte("{}"), 0o644)
			require.NoError(t, err)
		}

		statusFunc := func(ctx context.Context, path string) string {
			return ""
		}

		err := LinkTopSymlinks(context.Background(), mainDir, worktreeDir, statusFunc)
		assert.NoError(t, err)

		// Verify symlinks were created for all editor dirs
		for _, dir := range editorDirs {
			link := filepath.Join(worktreeDir, dir)
			target, err := os.Readlink(link)
			assert.NoError(t, err)
			assert.Equal(t, filepath.Join(mainDir, dir), target)
		}
	})

	t.Run("create tmp directory", func(t *testing.T) {
		mainDir := t.TempDir()
		worktreeDir := t.TempDir()

		statusFunc := func(ctx context.Context, path string) string {
			return ""
		}

		err := LinkTopSymlinks(context.Background(), mainDir, worktreeDir, statusFunc)
		assert.NoError(t, err)

		// Verify tmp directory was created
		tmpDir := filepath.Join(worktreeDir, "tmp")
		info, err := os.Stat(tmpDir)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("direnv allow with .envrc", func(t *testing.T) {
		mainDir := t.TempDir()
		worktreeDir := t.TempDir()

		// Create .envrc file
		envrcPath := filepath.Join(worktreeDir, ".envrc")
		err := os.WriteFile(envrcPath, []byte("export TEST=1"), 0o644)
		require.NoError(t, err)

		statusFunc := func(ctx context.Context, path string) string {
			return ""
		}

		// This will attempt to run direnv allow, which may not exist
		// The function doesn't return error if direnv fails (best-effort)
		err = LinkTopSymlinks(context.Background(), mainDir, worktreeDir, statusFunc)
		assert.NoError(t, err)
	})

	t.Run("handle empty status lines", func(t *testing.T) {
		mainDir := t.TempDir()
		worktreeDir := t.TempDir()

		statusFunc := func(ctx context.Context, path string) string {
			return "?? file.txt\n\n?? \n   \n"
		}

		// Create only the valid file
		file := filepath.Join(mainDir, "file.txt")
		err := os.WriteFile(file, []byte("content"), 0o644)
		require.NoError(t, err)

		err = LinkTopSymlinks(context.Background(), mainDir, worktreeDir, statusFunc)
		assert.NoError(t, err)

		// Only valid file should have symlink
		link := filepath.Join(worktreeDir, "file.txt")
		_, err = os.Readlink(link)
		assert.NoError(t, err)
	})

	t.Run("handle status with short lines", func(t *testing.T) {
		mainDir := t.TempDir()
		worktreeDir := t.TempDir()

		statusFunc := func(ctx context.Context, path string) string {
			return "?? file.txt\n??\nM\n   "
		}

		file := filepath.Join(mainDir, "file.txt")
		err := os.WriteFile(file, []byte("content"), 0o644)
		require.NoError(t, err)

		err = LinkTopSymlinks(context.Background(), mainDir, worktreeDir, statusFunc)
		assert.NoError(t, err)

		link := filepath.Join(worktreeDir, "file.txt")
		_, err = os.Readlink(link)
		assert.NoError(t, err)
	})

	t.Run("skip non-untracked files", func(t *testing.T) {
		mainDir := t.TempDir()
		worktreeDir := t.TempDir()

		// Create files
		tracked := filepath.Join(mainDir, "tracked.txt")
		err := os.WriteFile(tracked, []byte("tracked"), 0o644)
		require.NoError(t, err)

		modified := filepath.Join(mainDir, "modified.txt")
		err = os.WriteFile(modified, []byte("modified"), 0o644)
		require.NoError(t, err)

		statusFunc := func(ctx context.Context, path string) string {
			return " M tracked.txt\n M modified.txt\nA  added.txt"
		}

		err = LinkTopSymlinks(context.Background(), mainDir, worktreeDir, statusFunc)
		assert.NoError(t, err)

		// Verify no symlinks created for tracked/modified files
		_, err = os.Readlink(filepath.Join(worktreeDir, "tracked.txt"))
		assert.Error(t, err) // Should not exist

		_, err = os.Readlink(filepath.Join(worktreeDir, "modified.txt"))
		assert.Error(t, err) // Should not exist
	})
}
