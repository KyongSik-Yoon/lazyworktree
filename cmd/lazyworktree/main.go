// Package main is the entry point for the lazyworktree application.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/chmouel/lazyworktree/internal/app"
	"github.com/chmouel/lazyworktree/internal/config"
	"github.com/chmouel/lazyworktree/internal/log"
	"github.com/chmouel/lazyworktree/internal/theme"
	"github.com/chmouel/lazyworktree/internal/utils"
	kongcompletion "github.com/jotaen/kong-completion"
	"github.com/posener/complete"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

// CLI represents the main command-line interface structure.
type CLI struct {
	WorktreeDir      string   `help:"Override the default worktree root directory" short:"w"`
	DebugLog         string   `help:"Path to debug log file"`
	OutputSelection  string   `help:"Write selected worktree path to a file"`
	Theme            string   `help:"Override the UI theme" short:"t" predictor:"theme"`
	SearchAutoSelect bool     `help:"Start with filter focused"`
	Version          bool     `help:"Print version information" short:"v"`
	ShowSyntaxThemes bool     `help:"List available delta syntax themes"`
	ConfigFile       string   `help:"Path to configuration file"`
	Config           []string `help:"Override config values (repeatable): --config=lw.key=value" short:"C" predictor:"config" completion-shell-default:"false"`

	WtCreate   *WtCreateCmd              `cmd:"" help:"Create a new worktree"`
	WtDelete   *WtDeleteCmd              `cmd:"" help:"Delete a worktree"`
	Completion kongcompletion.Completion `cmd:"" help:"Generate or run shell completions"`
}

// WtCreateCmd represents the wt-create subcommand.
type WtCreateCmd struct {
	FromBranch string `help:"Create worktree from branch" xor:"source"`
	FromPR     int    `help:"Create worktree from PR number" xor:"source"`
	WithChange bool   `help:"Carry over uncommitted changes to the new worktree (only with --from-branch)"`
	Silent     bool   `help:"Suppress progress messages"`
}

// WtDeleteCmd represents the wt-delete subcommand.
type WtDeleteCmd struct {
	NoBranch     bool   `help:"Skip branch deletion"`
	Silent       bool   `help:"Suppress progress messages"`
	WorktreePath string `arg:"" optional:"" help:"Worktree path/name"`
}

func main() {
	// Handle special flags that exit early before Kong parsing
	// This is needed because Kong requires a subcommand when subcommands are defined
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "--version" || arg == "-v" {
			printVersion()
			return
		}
		if arg == "--show-syntax-themes" {
			printSyntaxThemes()
			return
		}
	}

	cli := &CLI{}
	parser, err := kong.New(cli,
		kong.Name("lazyworktree"),
		kong.Description("A TUI tool to manage git worktrees"),
		kong.UsageOnError(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating parser: %v\n", err)
		os.Exit(1)
	}

	// Set up kong-completion with custom predictors
	// This must happen before parsing so tab completion can be intercepted
	kongcompletion.Register(parser,
		kongcompletion.WithPredictor("theme", complete.PredictSet(theme.AvailableThemes()...)),
		kongcompletion.WithPredictor("config", configPredictor()),
	)

	// Check if a subcommand is provided in args
	hasSubcommand := false
	for _, arg := range args {
		if arg == "wt-create" || arg == "wt-delete" || arg == "completion" {
			hasSubcommand = true
			break
		}
	}

	// Extract potential filter args (non-flag, non-subcommand args) before Kong parsing
	// This is a simple heuristic: args that don't start with '-' and aren't subcommands
	var filterArgs []string
	skipNext := false
	for i, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		// Skip flags
		if strings.HasPrefix(arg, "-") {
			// Check if this flag takes a value (has = in it)
			if strings.Contains(arg, "=") {
				// Flag with =value, already handled
				continue
			}
			// Check if next arg is a value (doesn't start with -)
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				// Next arg might be a flag value, skip it
				skipNext = true
			}
			continue
		}
		// Check if it's a subcommand
		if arg == "wt-create" || arg == "wt-delete" || arg == "completion" {
			// Skip subcommand and let Kong handle it
			break
		}
		// This is a potential filter arg
		filterArgs = append(filterArgs, arg)
	}

	ctx, err := parser.Parse(args)
	var cmd string
	if err != nil {
		// If no subcommand was provided and we get a "missing command" error,
		// treat it as valid and proceed to TUI mode
		errStr := err.Error()
		if !hasSubcommand && strings.Contains(errStr, "expected one of") {
			// This is the "missing command" error - it's OK, we'll launch TUI
			// The flags should still be parsed in the cli struct
			// ctx will be nil, so we skip subcommand handling
			cmd = ""
		} else {
			// Some other error occurred, show it and exit
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	} else {
		// No error, get the command from context
		cmd = ctx.Command()
		// If completion command was selected, run it immediately and exit
		if cmd == "completion" || cmd == "completion <shell>" {
			if err := ctx.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			// ctx.Run() should have called ctx.Exit(0), but ensure we exit
			os.Exit(0)
		}
	}

	// Handle special flags that exit early (in case they weren't caught above)
	if cli.Version {
		printVersion()
		return
	}
	if cli.ShowSyntaxThemes {
		printSyntaxThemes()
		return
	}

	// Handle subcommands
	if cmd != "" {
		switch cmd {
		case "wt-create":
			handleWtCreate(cli.WtCreate, cli.WorktreeDir, cli.ConfigFile, cli.Config)
			return
		case "wt-delete":
			handleWtDelete(cli.WtDelete, cli.WorktreeDir, cli.ConfigFile, cli.Config)
			return
		case "completion":
			// This should have been handled above, but just in case
			if ctx != nil {
				if err := ctx.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
			}
			return
		}
	}

	// If no subcommand, use extracted filter args as initial filter for TUI
	initialFilter := strings.Join(filterArgs, " ")

	// Set up debug logging before loading config, so debug output is captured
	if cli.DebugLog != "" {
		expanded, err := utils.ExpandPath(cli.DebugLog)
		if err == nil {
			if err := log.SetFile(expanded); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening debug log file %q: %v\n", expanded, err)
			}
		} else {
			if err := log.SetFile(cli.DebugLog); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening debug log file %q: %v\n", cli.DebugLog, err)
			}
		}
	}

	cfg, err := config.LoadConfig(cli.ConfigFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// If debug log wasn't set via flag, check if it's in the config
	// If it is, enable logging. If not, disable logging and discard buffer.
	if cli.DebugLog == "" {
		if cfg.DebugLog != "" {
			expanded, err := utils.ExpandPath(cfg.DebugLog)
			path := cfg.DebugLog
			if err == nil {
				path = expanded
			}
			if err := log.SetFile(path); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening debug log file from config %q: %v\n", path, err)
			}
		} else {
			// No debug log configured, discard any buffered logs
			_ = log.SetFile("")
		}
	}

	if err := applyThemeConfig(cfg, cli.Theme); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		_ = log.Close()
		os.Exit(1)
	}
	if cli.SearchAutoSelect {
		cfg.SearchAutoSelect = true
	}

	if err := applyWorktreeDirConfig(cfg, cli.WorktreeDir); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		_ = log.Close()
		os.Exit(1)
	}

	if cli.DebugLog != "" {
		expanded, err := utils.ExpandPath(cli.DebugLog)
		if err == nil {
			cfg.DebugLog = expanded
		} else {
			cfg.DebugLog = cli.DebugLog
		}
	}

	// Apply CLI config overrides (highest precedence)
	if len(cli.Config) > 0 {
		if err := cfg.ApplyCLIOverrides(cli.Config); err != nil {
			fmt.Fprintf(os.Stderr, "Error applying config overrides: %v\n", err)
			_ = log.Close()
			os.Exit(1)
		}
	}

	model := app.NewModel(cfg, initialFilter)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	_, err = p.Run()
	model.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		_ = log.Close()
		os.Exit(1)
	}

	selectedPath := model.GetSelectedPath()
	if cli.OutputSelection != "" {
		expanded, err := utils.ExpandPath(cli.OutputSelection)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error expanding output-selection: %v\n", err)
			_ = log.Close()
			os.Exit(1)
		}
		const defaultDirPerms = 0o750
		if err := os.MkdirAll(filepath.Dir(expanded), defaultDirPerms); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output-selection dir: %v\n", err)
			_ = log.Close()
			os.Exit(1)
		}
		data := ""
		if selectedPath != "" {
			data = selectedPath + "\n"
		}
		const defaultFilePerms = 0o600
		if err := os.WriteFile(expanded, []byte(data), defaultFilePerms); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output-selection: %v\n", err)
			_ = log.Close()
			os.Exit(1)
		}
		return
	}
	if selectedPath != "" {
		fmt.Println(selectedPath)
	}
	if err := log.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error closing debug log: %v\n", err)
	}
}

// applyWorktreeDirConfig applies the worktree directory configuration.
// This ensures the same path expansion logic is used in both TUI and CLI modes.
func applyWorktreeDirConfig(cfg *config.AppConfig, worktreeDirFlag string) error {
	switch {
	case worktreeDirFlag != "":
		expanded, err := utils.ExpandPath(worktreeDirFlag)
		if err != nil {
			return fmt.Errorf("error expanding worktree-dir: %w", err)
		}
		cfg.WorktreeDir = expanded
	case cfg.WorktreeDir != "":
		expanded, err := utils.ExpandPath(cfg.WorktreeDir)
		if err == nil {
			cfg.WorktreeDir = expanded
		}
	default:
		home, _ := os.UserHomeDir()
		cfg.WorktreeDir = filepath.Join(home, ".local", "share", "worktrees")
	}
	return nil
}

func printSyntaxThemes() {
	names := theme.AvailableThemes()
	sort.Strings(names)
	fmt.Println("Available syntax themes (delta --syntax-theme defaults):")
	for _, name := range names {
		fmt.Printf("  %-16s -> %s\n", name, config.SyntaxThemeForUITheme(name))
	}
}

// configPredictor returns a predictor function for --config flag completion.
// It suggests config keys in the format "lw.key=value" and appropriate values for each key.
func configPredictor() complete.Predictor {
	return complete.PredictFunc(func(args complete.Args) []string {
		last := args.Last

		// If empty, suggest starting with "lw."
		if last == "" {
			return []string{"lw."}
		}

		// If it doesn't start with "lw.", suggest "lw."
		if !strings.HasPrefix(last, "lw.") {
			return []string{"lw."}
		}

		// Check if there's an "=" sign
		parts := strings.SplitN(last, "=", 2)
		if len(parts) == 1 {
			// No "=" yet, suggest config keys with full "lw.key=" format
			keyPrefix := strings.TrimPrefix(parts[0], "lw.")
			return suggestConfigKeys(keyPrefix)
		}

		// There's an "=", suggest values for the key
		key := strings.TrimPrefix(parts[0], "lw.")
		return suggestConfigValues(key)
	})
}

// suggestConfigKeys returns config key suggestions matching the prefix.
// Returns suggestions in the format "lw.key=" for completion.
func suggestConfigKeys(prefix string) []string {
	allKeys := []string{
		"theme", "worktree_dir", "sort_mode", "auto_fetch_prs", "auto_refresh",
		"refresh_interval", "search_auto_select", "fuzzy_finder_input", "show_icons",
		"max_untracked_diffs", "max_diff_chars", "max_name_length", "git_pager",
		"git_pager_args", "git_pager_interactive", "pager", "editor", "trust_mode",
		"debug_log", "init_commands", "terminate_commands", "merge_method",
		"issue_branch_name_template", "pr_branch_name_template", "branch_name_script",
		"session_prefix", "palette_mru", "palette_mru_limit",
	}

	var matches []string
	for _, key := range allKeys {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			// Return full format "lw.key=" for completion
			matches = append(matches, "lw."+key+"=")
		}
	}
	return matches
}

// suggestConfigValues returns value suggestions for a given config key.
func suggestConfigValues(key string) []string {
	switch key {
	case "theme":
		return theme.AvailableThemes()
	case "sort_mode":
		return []string{"switched", "active", "path"}
	case "merge_method":
		return []string{"rebase", "merge"}
	case "trust_mode":
		return []string{"tofu", "never", "always"}
	case "auto_fetch_prs", "auto_refresh", "search_auto_select", "fuzzy_finder_input",
		"show_icons", "git_pager_interactive", "palette_mru":
		return []string{"true", "false"}
	default:
		// For other keys, return empty to let shell handle file/path completion
		return nil
	}
}

// printVersion prints version information.
func printVersion() {
	v := version
	c := commit
	d := date
	b := builtBy

	if c == "none" || b == "unknown" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if c == "none" {
				for _, setting := range info.Settings {
					if setting.Key == "vcs.revision" {
						c = setting.Value
					}
				}
			}
			if b == "unknown" {
				b = info.GoVersion
			}
		}
	}

	fmt.Printf("lazyworktree version %s\ncommit: %s\nbuilt at: %s\nbuilt by: %s\n", v, c, d, b)
}

// applyThemeConfig applies theme configuration from command line flag.
func applyThemeConfig(cfg *config.AppConfig, themeName string) error {
	if themeName == "" {
		return nil
	}

	normalized := config.NormalizeThemeName(themeName)
	if normalized == "" {
		return fmt.Errorf("unknown theme %q", themeName)
	}

	cfg.Theme = normalized
	if !cfg.GitPagerArgsSet && filepath.Base(cfg.GitPager) == "delta" {
		cfg.GitPagerArgs = config.DefaultDeltaArgsForTheme(normalized)
	}

	return nil
}
