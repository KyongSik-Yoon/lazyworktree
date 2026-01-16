# Shell integration

Lazyworktree provides shell integration helpers to enhance your workflow when
working with Git worktrees. These helpers allow you to quickly jump to
worktrees and enable shell completion for the `lazyworktree` command.

It includes shell completion scripts for bash, zsh, and fish shells.

## Shell integration with Zsh

The "jump" helper changes your current directory to the selected worktree on exit. It uses `--output-selection` to write the selected path to a temporary file.

Option A, source the helper from a local clone:

```bash
# Add to .zshrc
source /path/to/lazyworktree/shell/functions.shell

# Create an alias for a specific repository
# worktree storage key is derived from the origin remote (e.g. github.com:owner/repo)
# and falls back to the directory basename when no remote is set.
jt() { worktree_jump ~/path/to/your/main/repo "$@"; }
```

Option B, download the helper and source it:

```bash
# Download the helper functions
mkdir -p ~/.shell/functions
curl -sL https://raw.githubusercontent.com/chmouel/lazyworktree/refs/heads/main/shell/functions.shell -o ~/.shell/functions/lazyworktree.shell

# Review and customise the functions if needed
# nano ~/.shell/functions/lazyworktree.shell

# Add to .zshrc
source ~/.shell/functions/lazyworktree.shell

# Create an alias for a specific repository
jt() { worktree_jump ~/path/to/your/main/repo "$@"; }
```

You can now run `jt` to open the Terminal User Interface, select a worktree, and upon pressing `Enter`, your shell will change directory to that location.

To jump directly to a worktree by name with shell completion enabled, use the following:

```bash
jt() { worktree_jump ~/path/to/your/main/repo "$@"; }
_jt() { _worktree_jump ~/path/to/your/main/repo; }
compdef _jt jt
```

Should you require a shortcut to the last-selected worktree, use the built-in `worktree_go_last` helper, which reads the `.last-selected` file:

```bash
alias pl='worktree_go_last ~/path/to/your/main/repo'
```

## Shell Completion

Generate completion scripts for bash, zsh, or fish:

```bash
# Bash
eval "$(lazyworktree completion bash --code)"

# Zsh
eval "$(lazyworktree completion zsh --code)"

# Fish
lazyworktree completion fish --code > ~/.config/fish/completions/lazyworktree.fish
```

Or simply run `lazyworktree completion` to see instructions for your shell.

Package manager installations (deb, rpm, AUR) include completions automatically.
