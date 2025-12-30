package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPRInfo(t *testing.T) {
	tests := []struct {
		name   string
		prInfo *PRInfo
	}{
		{
			name: "valid PR info",
			prInfo: &PRInfo{
				Number: 123,
				State:  "OPEN",
				Title:  "Test PR",
				URL:    "https://github.com/user/repo/pull/123",
			},
		},
		{
			name: "merged PR",
			prInfo: &PRInfo{
				Number: 456,
				State:  "MERGED",
				Title:  "Merged PR",
				URL:    "https://github.com/user/repo/pull/456",
			},
		},
		{
			name: "closed PR",
			prInfo: &PRInfo{
				Number: 789,
				State:  "CLOSED",
				Title:  "Closed PR",
				URL:    "https://github.com/user/repo/pull/789",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.prInfo)
			assert.Greater(t, tt.prInfo.Number, 0)
			assert.NotEmpty(t, tt.prInfo.State)
			assert.NotEmpty(t, tt.prInfo.Title)
			assert.NotEmpty(t, tt.prInfo.URL)
		})
	}
}

func TestWorktreeInfo(t *testing.T) {
	tests := []struct {
		name string
		wt   *WorktreeInfo
	}{
		{
			name: "main worktree clean",
			wt: &WorktreeInfo{
				Path:         "/repo/main",
				Branch:       "main",
				IsMain:       true,
				Dirty:        false,
				Ahead:        0,
				Behind:       0,
				LastActive:   "2024-01-01",
				LastActiveTS: 1704067200,
				PR:           nil,
				Untracked:    0,
				Modified:     0,
				Staged:       0,
				Divergence:   "",
			},
		},
		{
			name: "feature worktree with changes",
			wt: &WorktreeInfo{
				Path:         "/repo/feature",
				Branch:       "feature/test",
				IsMain:       false,
				Dirty:        true,
				Ahead:        3,
				Behind:       1,
				LastActive:   "2024-01-02",
				LastActiveTS: 1704153600,
				PR: &PRInfo{
					Number: 100,
					State:  "OPEN",
					Title:  "Feature PR",
					URL:    "https://github.com/user/repo/pull/100",
				},
				Untracked:  2,
				Modified:   5,
				Staged:     3,
				Divergence: "↑3 ↓1",
			},
		},
		{
			name: "worktree ahead of remote",
			wt: &WorktreeInfo{
				Path:         "/repo/ahead",
				Branch:       "ahead-branch",
				IsMain:       false,
				Dirty:        false,
				Ahead:        5,
				Behind:       0,
				LastActive:   "2024-01-03",
				LastActiveTS: 1704240000,
				Divergence:   "↑5",
			},
		},
		{
			name: "worktree behind remote",
			wt: &WorktreeInfo{
				Path:         "/repo/behind",
				Branch:       "behind-branch",
				IsMain:       false,
				Dirty:        false,
				Ahead:        0,
				Behind:       2,
				LastActive:   "2024-01-04",
				LastActiveTS: 1704326400,
				Divergence:   "↓2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.wt)
			assert.NotEmpty(t, tt.wt.Path)
			assert.NotEmpty(t, tt.wt.Branch)
			assert.GreaterOrEqual(t, tt.wt.Ahead, 0)
			assert.GreaterOrEqual(t, tt.wt.Behind, 0)
			assert.GreaterOrEqual(t, tt.wt.Untracked, 0)
			assert.GreaterOrEqual(t, tt.wt.Modified, 0)
			assert.GreaterOrEqual(t, tt.wt.Staged, 0)

			// Verify dirty flag is consistent with file changes
			if tt.wt.Untracked > 0 || tt.wt.Modified > 0 || tt.wt.Staged > 0 {
				assert.True(t, tt.wt.Dirty, "worktree should be dirty when it has changes")
			}
		})
	}
}

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "last selected filename",
			constant: LastSelectedFilename,
			expected: ".last-selected",
		},
		{
			name:     "cache filename",
			constant: CacheFilename,
			expected: ".worktree-cache.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
			assert.NotEmpty(t, tt.constant)
		})
	}
}
