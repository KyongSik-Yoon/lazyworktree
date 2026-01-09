package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = writer

	fn()

	_ = writer.Close()
	os.Stdout = orig

	out, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	return string(out)
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to read home dir: %v", err)
	}

	result, err := expandPath("~/worktrees")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(home, "worktrees")
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}

	t.Setenv("LW_TEST_DIR", "/tmp/lw")
	result, err = expandPath("$LW_TEST_DIR/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "/tmp/lw/path" {
		t.Fatalf("expected env expansion, got %q", result)
	}
}

func TestPrintSyntaxThemes(t *testing.T) {
	out := captureStdout(t, func() {
		printSyntaxThemes()
	})

	if !strings.Contains(out, "Available syntax themes") {
		t.Fatalf("expected header to be printed, got %q", out)
	}
	if !strings.Contains(out, "dracula") {
		t.Fatalf("expected theme list to include dracula, got %q", out)
	}
}

func TestPrintCompletion(t *testing.T) {
	shells := []string{"bash", "zsh", "fish"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			out := captureStdout(t, func() {
				if err := printCompletion(shell); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})

			if out == "" {
				t.Error("expected non-empty output")
			}

			// Verify it contains lazyworktree
			if !strings.Contains(out, "lazyworktree") {
				t.Error("output missing program name")
			}

			// Verify shell-specific structure
			switch shell {
			case "bash":
				if !strings.Contains(out, "_lazyworktree_completion") {
					t.Error("bash completion missing expected function")
				}
			case "zsh":
				if !strings.Contains(out, "#compdef") {
					t.Error("zsh completion missing compdef directive")
				}
			case "fish":
				if !strings.Contains(out, "complete -c") {
					t.Error("fish completion missing complete command")
				}
			}
		})
	}
}

func TestPrintCompletionInvalidShell(t *testing.T) {
	err := printCompletion("invalid")
	if err == nil {
		t.Error("expected error for invalid shell")
	}

	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("unexpected error message: %v", err)
	}
}
