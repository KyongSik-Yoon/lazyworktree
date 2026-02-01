package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chmouel/lazyworktree/internal/config"
	"github.com/chmouel/lazyworktree/internal/models"
)

// CommandRunner is a function type for creating exec.Cmd instances.
type CommandRunner func(ctx context.Context, name string, args ...string) *exec.Cmd

// TerminalTabLauncher launches commands in new terminal tabs.
type TerminalTabLauncher interface {
	// Name returns the terminal name for display.
	Name() string
	// IsAvailable checks if running inside this terminal.
	IsAvailable() bool
	// Launch opens a new tab with the given command.
	// Returns the tab title on success.
	Launch(ctx context.Context, cmd, cwd, title string, env map[string]string) (string, error)
}

// KittyLauncher implements TerminalTabLauncher for Kitty terminal.
type KittyLauncher struct {
	commandRunner CommandRunner
}

// Name returns "Kitty".
func (k *KittyLauncher) Name() string { return "Kitty" }

// IsAvailable checks if running inside Kitty terminal.
func (k *KittyLauncher) IsAvailable() bool {
	return os.Getenv("KITTY_WINDOW_ID") != ""
}

// Launch opens a new Kitty tab with the given command.
func (k *KittyLauncher) Launch(ctx context.Context, cmd, cwd, title string, env map[string]string) (string, error) {
	args := []string{"@", "launch", "--type=tab", "--cwd=" + cwd, "--tab-title=" + title}
	for key, val := range env {
		args = append(args, "--env="+key+"="+val)
	}
	args = append(args, "--", "bash", "-lc", cmd)

	c := k.commandRunner(ctx, "kitty", args...)
	output, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to launch Kitty tab: %w (%s)", err, string(output))
	}
	return title, nil
}

// detectTerminalLauncher returns the first available terminal launcher.
func detectTerminalLauncher(runner CommandRunner) TerminalTabLauncher {
	launchers := []TerminalTabLauncher{
		&KittyLauncher{commandRunner: runner},
		// Future: &ITermLauncher{}, &WezTermLauncher{}, etc.
	}
	for _, l := range launchers {
		if l.IsAvailable() {
			return l
		}
	}
	return nil
}

const terminalTabLabel = "terminal tab"

type terminalTabReadyMsg struct {
	terminalName string
	tabTitle     string
	err          error
}

func buildTerminalTabInfoMessage(terminal, title string) string {
	return fmt.Sprintf("Command launched in new %s tab: %s", terminal, title)
}

func (m *Model) openTerminalTab(customCmd *config.CustomCommand, wt *models.WorktreeInfo) tea.Cmd {
	if customCmd == nil || customCmd.Command == "" {
		return nil
	}

	launcher := detectTerminalLauncher(m.commandRunner)
	if launcher == nil {
		return func() tea.Msg {
			return terminalTabReadyMsg{err: fmt.Errorf("no supported terminal detected (Kitty required)")}
		}
	}

	env := m.buildCommandEnv(wt.Branch, wt.Path)
	title := customCmd.Description
	if title == "" {
		title = filepath.Base(wt.Path)
	}

	return func() tea.Msg {
		tabTitle, err := launcher.Launch(m.ctx, customCmd.Command, wt.Path, title, env)
		return terminalTabReadyMsg{
			terminalName: launcher.Name(),
			tabTitle:     tabTitle,
			err:          err,
		}
	}
}
