package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LinkTopSymlinks creates symlinks for untracked/ignored files and editor configs from main to target worktree.
func LinkTopSymlinks(ctx context.Context, mainPath, worktreePath string, statusFunc func(context.Context, string) string) error {
	if mainPath == "" || worktreePath == "" {
		return fmt.Errorf("missing paths for link_topsymlinks")
	}

	// Symlink untracked/ignored files
	status := statusFunc(ctx, mainPath)
	for _, line := range strings.Split(status, "\n") {
		if len(line) < 4 {
			continue
		}
		if !strings.HasPrefix(line, "?? ") && !strings.HasPrefix(line, "!! ") {
			continue
		}
		rel := strings.TrimSpace(line[3:])
		if rel == "" {
			continue
		}
		if err := symlinkPath(mainPath, worktreePath, rel); err != nil {
			return err
		}
	}

	// Symlink editor config dirs
	for _, name := range []string{".vscode", ".idea", ".cursor", ".claude"} {
		if err := symlinkPath(mainPath, worktreePath, name); err != nil {
			return err
		}
	}

	// Create tmp/
	if err := os.MkdirAll(filepath.Join(worktreePath, "tmp"), 0o755); err != nil {
		return err
	}

	// direnv allow if .envrc exists
	envrcPath := filepath.Join(worktreePath, ".envrc")
	if _, err := os.Stat(envrcPath); err == nil {
		cmd := exec.CommandContext(ctx, "direnv", "allow")
		cmd.Dir = worktreePath
		_ = cmd.Run() // best-effort
	}

	return nil
}

func symlinkPath(mainPath, worktreePath, rel string) error {
	src := filepath.Join(mainPath, rel)
	if _, err := os.Stat(src); err != nil {
		return nil
	}

	dst := filepath.Join(worktreePath, rel)
	if _, err := os.Lstat(dst); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	_ = os.Remove(dst)
	return os.Symlink(src, dst)
}
