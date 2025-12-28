from dataclasses import dataclass
from typing import Optional

@dataclass
class PRInfo:
    number: int
    state: str
    title: str
    url: str

@dataclass
class WorktreeInfo:
    path: str
    branch: str
    is_main: bool
    dirty: bool
    ahead: int = 0
    behind: int = 0
    last_active: str = ""
    last_active_ts: int = 0
    pr: Optional[PRInfo] = None
    untracked: int = 0
    modified: int = 0
    staged: int = 0
    divergence: str = ""

WORKTREE_DIR = "~/.local/share/worktrees"
LAST_SELECTED_FILENAME = ".last-selected"
CACHE_FILENAME = ".worktree-cache.json"
