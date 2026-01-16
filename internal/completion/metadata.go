package completion

import "github.com/chmouel/lazyworktree/internal/theme"

// FlagInfo contains metadata about a command-line flag for completion generation.
type FlagInfo struct {
	Name        string   // Flag name without dashes
	Description string   // Human-readable description
	HasValue    bool     // true for string flags, false for bool flags
	ValueHint   string   // Hint for value type (e.g., "DIR", "PATH", "NAME")
	Values      []string // Enumerated values for completion (e.g., theme names)
}

// SubcommandInfo contains metadata about a subcommand for completion generation.
type SubcommandInfo struct {
	Name        string     // Subcommand name
	Description string     // Human-readable description
	Flags       []FlagInfo // Flags specific to this subcommand
}

// GetFlags returns metadata for all lazyworktree command-line flags.
// This is the single source of truth for shell completion generation.
func GetFlags() []FlagInfo {
	return []FlagInfo{
		{
			Name:        "worktree-dir",
			Description: "Override default worktree root directory",
			HasValue:    true,
			ValueHint:   "DIR",
		},
		{
			Name:        "debug-log",
			Description: "Path to debug log file",
			HasValue:    true,
			ValueHint:   "PATH",
		},
		{
			Name:        "output-selection",
			Description: "Write selected path to file",
			HasValue:    true,
			ValueHint:   "FILE",
		},
		{
			Name:        "theme",
			Description: "Override UI theme",
			HasValue:    true,
			ValueHint:   "NAME",
			Values:      theme.AvailableThemes(),
		},
		{
			Name:        "search-auto-select",
			Description: "Start with filter focused",
			HasValue:    false,
		},
		{
			Name:        "version",
			Description: "Print version information",
			HasValue:    false,
		},
		{
			Name:        "show-syntax-themes",
			Description: "List available delta syntax themes",
			HasValue:    false,
		},
		{
			Name:        "completion",
			Description: "Generate shell completion script",
			HasValue:    true,
			ValueHint:   "SHELL",
			Values:      []string{"bash", "zsh", "fish"},
		},
		{
			Name:        "config",
			Description: "Path to configuration file",
			HasValue:    true,
			ValueHint:   "FILE",
		},
	}
}

// GetSubcommands returns metadata for all lazyworktree subcommands.
func GetSubcommands() []SubcommandInfo {
	return []SubcommandInfo{
		{
			Name:        "wt-create",
			Description: "Create a new worktree",
			Flags: []FlagInfo{
				{
					Name:        "from-branch",
					Description: "Create from branch",
					HasValue:    true,
					ValueHint:   "BRANCH",
				},
				{
					Name:        "from-pr",
					Description: "Create from PR number",
					HasValue:    true,
					ValueHint:   "NUMBER",
				},
				{
					Name:        "with-change",
					Description: "Carry over uncommitted changes (only with --from-branch)",
					HasValue:    false,
				},
				{
					Name:        "silent",
					Description: "Suppress progress messages",
					HasValue:    false,
				},
			},
		},
		{
			Name:        "wt-delete",
			Description: "Delete a worktree",
			Flags: []FlagInfo{
				{
					Name:        "no-branch",
					Description: "Skip branch deletion",
					HasValue:    false,
				},
				{
					Name:        "silent",
					Description: "Suppress progress messages",
					HasValue:    false,
				},
			},
		},
	}
}
