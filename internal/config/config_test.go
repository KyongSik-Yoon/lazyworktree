package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.True(t, cfg.SortByActive)
	assert.False(t, cfg.AutoFetchPRs)
	assert.Equal(t, 10, cfg.MaxUntrackedDiffs)
	assert.Equal(t, 200000, cfg.MaxDiffChars)
	assert.Equal(t, "tofu", cfg.TrustMode)
	assert.Empty(t, cfg.WorktreeDir)
	assert.Empty(t, cfg.InitCommands)
	assert.Empty(t, cfg.TerminateCommands)
	assert.Empty(t, cfg.DebugLog)
}

func TestNormalizeCommandList(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "whitespace only string",
			input:    "   ",
			expected: []string{},
		},
		{
			name:     "single command string",
			input:    "echo hello",
			expected: []string{"echo hello"},
		},
		{
			name:     "trimmed string",
			input:    "  echo hello  ",
			expected: []string{"echo hello"},
		},
		{
			name:     "empty list",
			input:    []interface{}{},
			expected: []string{},
		},
		{
			name:     "list with single command",
			input:    []interface{}{"echo hello"},
			expected: []string{"echo hello"},
		},
		{
			name:     "list with multiple commands",
			input:    []interface{}{"echo hello", "ls -la", "pwd"},
			expected: []string{"echo hello", "ls -la", "pwd"},
		},
		{
			name:     "list with nil elements",
			input:    []interface{}{"echo hello", nil, "pwd"},
			expected: []string{"echo hello", "pwd"},
		},
		{
			name:     "list with empty strings",
			input:    []interface{}{"echo hello", "", "pwd"},
			expected: []string{"echo hello", "pwd"},
		},
		{
			name:     "list with whitespace strings",
			input:    []interface{}{"echo hello", "   ", "pwd"},
			expected: []string{"echo hello", "pwd"},
		},
		{
			name:     "list with trimmed strings",
			input:    []interface{}{"  echo hello  ", "  pwd  "},
			expected: []string{"echo hello", "pwd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeCommandList(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCoerceBool(t *testing.T) {
	tests := []struct {
		name       string
		input      interface{}
		defaultVal bool
		expected   bool
	}{
		{
			name:       "nil with default true",
			input:      nil,
			defaultVal: true,
			expected:   true,
		},
		{
			name:       "nil with default false",
			input:      nil,
			defaultVal: false,
			expected:   false,
		},
		{
			name:       "bool true",
			input:      true,
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "bool false",
			input:      false,
			defaultVal: true,
			expected:   false,
		},
		{
			name:       "int 1",
			input:      1,
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "int 0",
			input:      0,
			defaultVal: true,
			expected:   false,
		},
		{
			name:       "int non-zero",
			input:      42,
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "string true",
			input:      "true",
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "string false",
			input:      "false",
			defaultVal: true,
			expected:   false,
		},
		{
			name:       "string 1",
			input:      "1",
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "string 0",
			input:      "0",
			defaultVal: true,
			expected:   false,
		},
		{
			name:       "string yes",
			input:      "yes",
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "string no",
			input:      "no",
			defaultVal: true,
			expected:   false,
		},
		{
			name:       "string y",
			input:      "y",
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "string n",
			input:      "n",
			defaultVal: true,
			expected:   false,
		},
		{
			name:       "string on",
			input:      "on",
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "string off",
			input:      "off",
			defaultVal: true,
			expected:   false,
		},
		{
			name:       "string with whitespace",
			input:      "  true  ",
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "string uppercase",
			input:      "TRUE",
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "invalid string",
			input:      "invalid",
			defaultVal: true,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := coerceBool(tt.input, tt.defaultVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCoerceInt(t *testing.T) {
	tests := []struct {
		name       string
		input      interface{}
		defaultVal int
		expected   int
	}{
		{
			name:       "nil with default",
			input:      nil,
			defaultVal: 42,
			expected:   42,
		},
		{
			name:       "int value",
			input:      123,
			defaultVal: 42,
			expected:   123,
		},
		{
			name:       "bool (should return default)",
			input:      true,
			defaultVal: 42,
			expected:   42,
		},
		{
			name:       "string number",
			input:      "123",
			defaultVal: 42,
			expected:   123,
		},
		{
			name:       "string with whitespace",
			input:      "  456  ",
			defaultVal: 42,
			expected:   456,
		},
		{
			name:       "empty string",
			input:      "",
			defaultVal: 42,
			expected:   42,
		},
		{
			name:       "invalid string",
			input:      "abc",
			defaultVal: 42,
			expected:   42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := coerceInt(tt.input, tt.defaultVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		validate func(*testing.T, *AppConfig)
	}{
		{
			name: "empty config uses defaults",
			data: map[string]interface{}{},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.True(t, cfg.SortByActive)
				assert.False(t, cfg.AutoFetchPRs)
				assert.Equal(t, 10, cfg.MaxUntrackedDiffs)
				assert.Equal(t, 200000, cfg.MaxDiffChars)
				assert.Equal(t, "tofu", cfg.TrustMode)
			},
		},
		{
			name: "worktree_dir",
			data: map[string]interface{}{
				"worktree_dir": "/custom/path",
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.Equal(t, "/custom/path", cfg.WorktreeDir)
			},
		},
		{
			name: "debug_log",
			data: map[string]interface{}{
				"debug_log": "/tmp/debug.log",
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.Equal(t, "/tmp/debug.log", cfg.DebugLog)
			},
		},
		{
			name: "init_commands string",
			data: map[string]interface{}{
				"init_commands": "echo hello",
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.Equal(t, []string{"echo hello"}, cfg.InitCommands)
			},
		},
		{
			name: "init_commands list",
			data: map[string]interface{}{
				"init_commands": []interface{}{"echo hello", "pwd"},
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.Equal(t, []string{"echo hello", "pwd"}, cfg.InitCommands)
			},
		},
		{
			name: "terminate_commands",
			data: map[string]interface{}{
				"terminate_commands": []interface{}{"cleanup", "exit"},
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.Equal(t, []string{"cleanup", "exit"}, cfg.TerminateCommands)
			},
		},
		{
			name: "sort_by_active false",
			data: map[string]interface{}{
				"sort_by_active": false,
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.False(t, cfg.SortByActive)
			},
		},
		{
			name: "auto_fetch_prs true",
			data: map[string]interface{}{
				"auto_fetch_prs": true,
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.True(t, cfg.AutoFetchPRs)
			},
		},
		{
			name: "max_untracked_diffs",
			data: map[string]interface{}{
				"max_untracked_diffs": 20,
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.Equal(t, 20, cfg.MaxUntrackedDiffs)
			},
		},
		{
			name: "max_diff_chars",
			data: map[string]interface{}{
				"max_diff_chars": 100000,
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.Equal(t, 100000, cfg.MaxDiffChars)
			},
		},
		{
			name: "negative max_untracked_diffs becomes 0",
			data: map[string]interface{}{
				"max_untracked_diffs": -5,
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.Equal(t, 0, cfg.MaxUntrackedDiffs)
			},
		},
		{
			name: "negative max_diff_chars becomes 0",
			data: map[string]interface{}{
				"max_diff_chars": -1000,
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.Equal(t, 0, cfg.MaxDiffChars)
			},
		},
		{
			name: "trust_mode tofu",
			data: map[string]interface{}{
				"trust_mode": "tofu",
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.Equal(t, "tofu", cfg.TrustMode)
			},
		},
		{
			name: "trust_mode never",
			data: map[string]interface{}{
				"trust_mode": "never",
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.Equal(t, "never", cfg.TrustMode)
			},
		},
		{
			name: "trust_mode always",
			data: map[string]interface{}{
				"trust_mode": "always",
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.Equal(t, "always", cfg.TrustMode)
			},
		},
		{
			name: "trust_mode uppercase converted to lowercase",
			data: map[string]interface{}{
				"trust_mode": "TOFU",
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.Equal(t, "tofu", cfg.TrustMode)
			},
		},
		{
			name: "invalid trust_mode uses default",
			data: map[string]interface{}{
				"trust_mode": "invalid",
			},
			validate: func(t *testing.T, cfg *AppConfig) {
				assert.Equal(t, "tofu", cfg.TrustMode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := parseConfig(tt.data)
			assert.NotNil(t, cfg)
			tt.validate(t, cfg)
		})
	}
}

func TestLoadRepoConfig(t *testing.T) {
	t.Run("empty repo path", func(t *testing.T) {
		cfg, path, err := LoadRepoConfig("")
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Empty(t, path)
	})

	t.Run("non-existent .wt file", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, path, err := LoadRepoConfig(tmpDir)
		assert.NoError(t, err)
		assert.Nil(t, cfg)
		assert.Equal(t, filepath.Join(tmpDir, ".wt"), path)
	})

	t.Run("valid .wt file", func(t *testing.T) {
		tmpDir := t.TempDir()
		wtPath := filepath.Join(tmpDir, ".wt")

		yamlContent := `init_commands:
  - echo "init"
  - pwd
terminate_commands:
  - echo "terminate"
`
		err := os.WriteFile(wtPath, []byte(yamlContent), 0o644)
		require.NoError(t, err)

		cfg, path, err := LoadRepoConfig(tmpDir)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, wtPath, path)
		assert.Equal(t, wtPath, cfg.Path)
		assert.Equal(t, []string{"echo \"init\"", "pwd"}, cfg.InitCommands)
		assert.Equal(t, []string{"echo \"terminate\""}, cfg.TerminateCommands)
	})

	t.Run("invalid YAML in .wt file", func(t *testing.T) {
		tmpDir := t.TempDir()
		wtPath := filepath.Join(tmpDir, ".wt")

		err := os.WriteFile(wtPath, []byte("invalid: yaml: content: [[["), 0o644)
		require.NoError(t, err)

		cfg, path, err := LoadRepoConfig(tmpDir)
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Equal(t, wtPath, path)
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("no config file returns defaults", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "nonexistent.yaml")

		cfg, err := LoadConfig(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, DefaultConfig().SortByActive, cfg.SortByActive)
		assert.Equal(t, DefaultConfig().MaxUntrackedDiffs, cfg.MaxUntrackedDiffs)
	})

	t.Run("valid config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		yamlContent := `worktree_dir: /custom/worktrees
sort_by_active: false
auto_fetch_prs: true
max_untracked_diffs: 20
max_diff_chars: 100000
trust_mode: always
init_commands:
  - echo "init"
terminate_commands:
  - echo "cleanup"
`
		err := os.WriteFile(configPath, []byte(yamlContent), 0o644)
		require.NoError(t, err)

		cfg, err := LoadConfig(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "/custom/worktrees", cfg.WorktreeDir)
		assert.False(t, cfg.SortByActive)
		assert.True(t, cfg.AutoFetchPRs)
		assert.Equal(t, 20, cfg.MaxUntrackedDiffs)
		assert.Equal(t, 100000, cfg.MaxDiffChars)
		assert.Equal(t, "always", cfg.TrustMode)
		assert.Equal(t, []string{"echo \"init\""}, cfg.InitCommands)
		assert.Equal(t, []string{"echo \"cleanup\""}, cfg.TerminateCommands)
	})

	t.Run("invalid YAML returns defaults", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		err := os.WriteFile(configPath, []byte("invalid: [[["), 0o644)
		require.NoError(t, err)

		cfg, err := LoadConfig(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, DefaultConfig().SortByActive, cfg.SortByActive)
	})
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		setup    func()
		cleanup  func()
		validate func(*testing.T, string)
	}{
		{
			name:    "path without tilde",
			input:   "/absolute/path",
			setup:   func() {},
			cleanup: func() {},
			validate: func(t *testing.T, result string) {
				assert.Equal(t, "/absolute/path", result)
			},
		},
		{
			name:    "path with tilde",
			input:   "~/test/path",
			setup:   func() {},
			cleanup: func() {},
			validate: func(t *testing.T, result string) {
				home, _ := os.UserHomeDir()
				assert.Equal(t, filepath.Join(home, "test/path"), result)
			},
		},
		{
			name:    "path with environment variable",
			input:   "$HOME/test",
			setup:   func() {},
			cleanup: func() {},
			validate: func(t *testing.T, result string) {
				home := os.Getenv("HOME")
				assert.Equal(t, filepath.Join(home, "test"), result)
			},
		},
		{
			name:  "path with custom env var",
			input: "$CUSTOM_VAR/test",
			setup: func() {
				_ = os.Setenv("CUSTOM_VAR", "/custom")
			},
			cleanup: func() {
				_ = os.Unsetenv("CUSTOM_VAR")
			},
			validate: func(t *testing.T, result string) {
				assert.Equal(t, "/custom/test", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			result, err := expandPath(tt.input)
			assert.NoError(t, err)
			tt.validate(t, result)
		})
	}
}
