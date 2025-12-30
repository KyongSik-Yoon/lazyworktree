package models

type PRInfo struct {
	Number int
	State  string
	Title  string
	URL    string
}

type WorktreeInfo struct {
	Path         string
	Branch       string
	IsMain       bool
	Dirty        bool
	Ahead        int
	Behind       int
	LastActive   string
	LastActiveTS int64
	PR           *PRInfo
	Untracked    int
	Modified     int
	Staged       int
	Divergence   string
}

const (
	LastSelectedFilename = ".last-selected"
	CacheFilename        = ".worktree-cache.json"
)
