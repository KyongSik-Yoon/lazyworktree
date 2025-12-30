package security

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTrustManager(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "trusted.json")

	tm := &TrustManager{
		dbPath:        dbPath,
		trustedHashes: make(map[string]string),
	}

	assert.NotNil(t, tm)
	assert.Equal(t, dbPath, tm.dbPath)
	assert.NotNil(t, tm.trustedHashes)
	assert.Empty(t, tm.trustedHashes)
}

func TestCalculateHash(t *testing.T) {
	t.Run("calculate hash for existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		content := "test content"
		err := os.WriteFile(testFile, []byte(content), 0o644)
		require.NoError(t, err)

		tm := &TrustManager{
			dbPath:        filepath.Join(tmpDir, "trusted.json"),
			trustedHashes: make(map[string]string),
		}

		hash1 := tm.calculateHash(testFile)
		assert.NotEmpty(t, hash1)
		assert.Len(t, hash1, 64) // SHA256 hash is 64 hex chars

		// Same content should produce same hash
		hash2 := tm.calculateHash(testFile)
		assert.Equal(t, hash1, hash2)
	})

	t.Run("calculate hash for non-existent file", func(t *testing.T) {
		tm := &TrustManager{
			dbPath:        "/tmp/trusted.json",
			trustedHashes: make(map[string]string),
		}

		hash := tm.calculateHash("/nonexistent/file.txt")
		assert.Empty(t, hash)
	})

	t.Run("different content produces different hash", func(t *testing.T) {
		tmpDir := t.TempDir()

		file1 := filepath.Join(tmpDir, "file1.txt")
		err := os.WriteFile(file1, []byte("content1"), 0o644)
		require.NoError(t, err)

		file2 := filepath.Join(tmpDir, "file2.txt")
		err = os.WriteFile(file2, []byte("content2"), 0o644)
		require.NoError(t, err)

		tm := &TrustManager{
			dbPath:        filepath.Join(tmpDir, "trusted.json"),
			trustedHashes: make(map[string]string),
		}

		hash1 := tm.calculateHash(file1)
		hash2 := tm.calculateHash(file2)

		assert.NotEmpty(t, hash1)
		assert.NotEmpty(t, hash2)
		assert.NotEqual(t, hash1, hash2)
	})
}

func TestLoad(t *testing.T) {
	t.Run("load non-existent database", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "trusted.json")

		tm := &TrustManager{
			dbPath:        dbPath,
			trustedHashes: make(map[string]string),
		}

		tm.load()
		assert.Empty(t, tm.trustedHashes)
	})

	t.Run("load existing database", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "trusted.json")

		// Create a trust database with some entries
		trustedData := map[string]string{
			"/path/to/file1.txt": "hash1",
			"/path/to/file2.txt": "hash2",
		}
		data, err := json.MarshalIndent(trustedData, "", "  ")
		require.NoError(t, err)
		err = os.MkdirAll(filepath.Dir(dbPath), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(dbPath, data, 0o644)
		require.NoError(t, err)

		tm := &TrustManager{
			dbPath:        dbPath,
			trustedHashes: make(map[string]string),
		}

		tm.load()
		assert.Len(t, tm.trustedHashes, 2)
		assert.Equal(t, "hash1", tm.trustedHashes["/path/to/file1.txt"])
		assert.Equal(t, "hash2", tm.trustedHashes["/path/to/file2.txt"])
	})

	t.Run("load corrupt database", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "trusted.json")

		// Write invalid JSON
		err := os.MkdirAll(filepath.Dir(dbPath), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(dbPath, []byte("invalid json {{{"), 0o644)
		require.NoError(t, err)

		tm := &TrustManager{
			dbPath:        dbPath,
			trustedHashes: make(map[string]string),
		}

		tm.load()
		// Should start with empty map on corrupt data
		assert.Empty(t, tm.trustedHashes)
	})
}

func TestSave(t *testing.T) {
	t.Run("save trust database", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "trusted.json")

		tm := &TrustManager{
			dbPath:        dbPath,
			trustedHashes: make(map[string]string),
		}

		tm.trustedHashes["/path/to/file1.txt"] = "hash1"
		tm.trustedHashes["/path/to/file2.txt"] = "hash2"

		err := tm.save()
		assert.NoError(t, err)

		// Verify file was created
		_, err = os.Stat(dbPath)
		assert.NoError(t, err)

		// Verify content
		data, err := os.ReadFile(dbPath)
		require.NoError(t, err)

		var loaded map[string]string
		err = json.Unmarshal(data, &loaded)
		assert.NoError(t, err)
		assert.Len(t, loaded, 2)
		assert.Equal(t, "hash1", loaded["/path/to/file1.txt"])
		assert.Equal(t, "hash2", loaded["/path/to/file2.txt"])
	})

	t.Run("save creates directory if missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "nested", "dir", "trusted.json")

		tm := &TrustManager{
			dbPath:        dbPath,
			trustedHashes: make(map[string]string),
		}

		tm.trustedHashes["/path/to/file.txt"] = "hash"

		err := tm.save()
		assert.NoError(t, err)

		// Verify file and directory were created
		_, err = os.Stat(dbPath)
		assert.NoError(t, err)
	})
}

func TestCheckTrust(t *testing.T) {
	t.Run("check trust for non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tm := &TrustManager{
			dbPath:        filepath.Join(tmpDir, "trusted.json"),
			trustedHashes: make(map[string]string),
		}

		status := tm.CheckTrust("/nonexistent/file.txt")
		assert.Equal(t, TrustStatusNotFound, status)
	})

	t.Run("check trust for untrusted file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(testFile, []byte("content"), 0o644)
		require.NoError(t, err)

		tm := &TrustManager{
			dbPath:        filepath.Join(tmpDir, "trusted.json"),
			trustedHashes: make(map[string]string),
		}

		status := tm.CheckTrust(testFile)
		assert.Equal(t, TrustStatusUntrusted, status)
	})

	t.Run("check trust for trusted file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		content := "test content"
		err := os.WriteFile(testFile, []byte(content), 0o644)
		require.NoError(t, err)

		tm := &TrustManager{
			dbPath:        filepath.Join(tmpDir, "trusted.json"),
			trustedHashes: make(map[string]string),
		}

		// First trust the file
		err = tm.TrustFile(testFile)
		require.NoError(t, err)

		// Now check trust - should be trusted
		status := tm.CheckTrust(testFile)
		assert.Equal(t, TrustStatusTrusted, status)
	})

	t.Run("check trust for modified file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(testFile, []byte("original content"), 0o644)
		require.NoError(t, err)

		tm := &TrustManager{
			dbPath:        filepath.Join(tmpDir, "trusted.json"),
			trustedHashes: make(map[string]string),
		}

		// Trust the original file
		err = tm.TrustFile(testFile)
		require.NoError(t, err)

		// Modify the file
		err = os.WriteFile(testFile, []byte("modified content"), 0o644)
		require.NoError(t, err)

		// Check trust - should be untrusted due to content change
		status := tm.CheckTrust(testFile)
		assert.Equal(t, TrustStatusUntrusted, status)
	})

	t.Run("check trust with relative path", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(testFile, []byte("content"), 0o644)
		require.NoError(t, err)

		// Save current directory and change to tmpDir
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		defer func() {
			_ = os.Chdir(originalDir)
		}()

		tm := &TrustManager{
			dbPath:        filepath.Join(tmpDir, "trusted.json"),
			trustedHashes: make(map[string]string),
		}

		// Trust using relative path
		err = tm.TrustFile("test.txt")
		require.NoError(t, err)

		// Check using relative path
		status := tm.CheckTrust("test.txt")
		assert.Equal(t, TrustStatusTrusted, status)
	})
}

func TestTrustFile(t *testing.T) {
	t.Run("trust existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		content := "test content"
		err := os.WriteFile(testFile, []byte(content), 0o644)
		require.NoError(t, err)

		tm := &TrustManager{
			dbPath:        filepath.Join(tmpDir, "trusted.json"),
			trustedHashes: make(map[string]string),
		}

		err = tm.TrustFile(testFile)
		assert.NoError(t, err)

		// Verify hash was stored
		absPath, _ := filepath.Abs(testFile)
		hash, exists := tm.trustedHashes[absPath]
		assert.True(t, exists)
		assert.NotEmpty(t, hash)

		// Verify database was saved
		_, err = os.Stat(tm.dbPath)
		assert.NoError(t, err)
	})

	t.Run("trust non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tm := &TrustManager{
			dbPath:        filepath.Join(tmpDir, "trusted.json"),
			trustedHashes: make(map[string]string),
		}

		err := tm.TrustFile("/nonexistent/file.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("trust file updates hash on content change", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(testFile, []byte("original"), 0o644)
		require.NoError(t, err)

		tm := &TrustManager{
			dbPath:        filepath.Join(tmpDir, "trusted.json"),
			trustedHashes: make(map[string]string),
		}

		// Trust original
		err = tm.TrustFile(testFile)
		require.NoError(t, err)

		absPath, _ := filepath.Abs(testFile)
		hash1 := tm.trustedHashes[absPath]

		// Modify and trust again
		err = os.WriteFile(testFile, []byte("modified"), 0o644)
		require.NoError(t, err)

		err = tm.TrustFile(testFile)
		require.NoError(t, err)

		hash2 := tm.trustedHashes[absPath]

		// Hash should be different
		assert.NotEqual(t, hash1, hash2)
	})
}

func TestTrustStatus(t *testing.T) {
	tests := []struct {
		name     string
		expected TrustStatus
	}{
		{
			name:     "trusted",
			expected: TrustStatusTrusted,
		},
		{
			name:     "untrusted",
			expected: TrustStatusUntrusted,
		},
		{
			name:     "not found",
			expected: TrustStatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.GreaterOrEqual(t, int(tt.expected), 0)
		})
	}
}

func TestGetTrustDBPath(t *testing.T) {
	t.Run("uses XDG_DATA_HOME when set", func(t *testing.T) {
		originalXDG := os.Getenv("XDG_DATA_HOME")
		defer func() { _ = os.Setenv("XDG_DATA_HOME", originalXDG) }()

		_ = os.Setenv("XDG_DATA_HOME", "/custom/data")
		path := getTrustDBPath()
		assert.Equal(t, "/custom/data/lazyworktree/trusted.json", path)
	})

	t.Run("falls back to HOME/.local/share when XDG_DATA_HOME not set", func(t *testing.T) {
		originalXDG := os.Getenv("XDG_DATA_HOME")
		defer func() { _ = os.Setenv("XDG_DATA_HOME", originalXDG) }()

		_ = os.Unsetenv("XDG_DATA_HOME")
		path := getTrustDBPath()

		home, _ := os.UserHomeDir()
		expectedPath := filepath.Join(home, ".local", "share", "lazyworktree", "trusted.json")
		assert.Equal(t, expectedPath, path)
	})
}
