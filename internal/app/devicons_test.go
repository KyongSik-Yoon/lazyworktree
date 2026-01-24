package app

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeviconForNameEmpty(t *testing.T) {
	SetIconProvider(&NerdFontV3Provider{})
	result := deviconForName("", false)
	assert.Empty(t, result)
}

func TestDeviconForNameFile(t *testing.T) {
	SetIconProvider(&NerdFontV3Provider{})
	tests := []struct {
		name     string
		isDir    bool
		fileName string
		expected string
	}{
		{"go file", false, "main.go", ""},
		{"markdown file", false, "README.md", "󰂺"},
		{"directory", true, "src", ""},
		{"unknown file", false, "file.unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deviconForName(tt.fileName, tt.isDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNerdFontV3ProviderFileIcons(t *testing.T) {
	provider := &NerdFontV3Provider{}
	assert.Equal(t, "", provider.GetFileIcon("main.vue", false))
}

func TestTextProviderFileIcons(t *testing.T) {
	provider := &TextProvider{}
	assert.Empty(t, provider.GetFileIcon("main.vue", false))
	assert.Equal(t, "/", provider.GetFileIcon("src", true))
}

func TestUIIconUsesProvider(t *testing.T) {
	SetIconProvider(&EmojiProvider{})
	assert.Equal(t, "🔍", uiIcon(UIIconSearch))

	SetIconProvider(&TextProvider{})
	assert.Equal(t, "/", uiIcon(UIIconSearch))

	SetIconProvider(&NerdFontV3Provider{})
	assert.Equal(t, nerdFontUIIconSearch, uiIcon(UIIconSearch))
}

func TestLabelWithIconToggle(t *testing.T) {
	SetIconProvider(&NerdFontV3Provider{})
	label := labelWithIcon(UIIconSearch, "Search", true)
	assert.Equal(t, nerdFontUIIconSearch+" Search", label)

	label = labelWithIcon(UIIconSearch, "Search", false)
	assert.Equal(t, "Search", label)
}

func TestStatusAndSyncIndicators(t *testing.T) {
	SetIconProvider(&EmojiProvider{})
	t.Cleanup(func() { SetIconProvider(&NerdFontV3Provider{}) })
	assert.Equal(t, "✅", statusIndicator(true, true))
	assert.Equal(t, "✏️", statusIndicator(false, true))
	assert.Equal(t, " ", statusIndicator(true, false))
	assert.Equal(t, "~", statusIndicator(false, false))
	assert.Equal(t, "✅", syncIndicator(true))
	assert.Equal(t, "-", syncIndicator(false))
}

func TestCIIconForConclusionSuccess(t *testing.T) {
	// Set default provider for testing
	SetIconProvider(&NerdFontV3Provider{})
	result := ciIconForConclusion("success")
	assert.Equal(t, "", result)
}

func TestCIIconForConclusionFailure(t *testing.T) {
	// Set default provider for testing
	SetIconProvider(&NerdFontV3Provider{})
	result := ciIconForConclusion("failure")
	assert.Equal(t, "", result)
}

func TestCIIconForConclusionSkipped(t *testing.T) {
	// Set default provider for testing
	SetIconProvider(&NerdFontV3Provider{})
	result := ciIconForConclusion("skipped")
	assert.Equal(t, "", result)
}

func TestCIIconForConclusionCancelled(t *testing.T) {
	// Set default provider for testing
	SetIconProvider(&NerdFontV3Provider{})
	result := ciIconForConclusion("cancelled")
	assert.Equal(t, "", result)
}

func TestCIIconForConclusionPending(t *testing.T) {
	// Set default provider for testing
	SetIconProvider(&NerdFontV3Provider{})
	result := ciIconForConclusion("pending")
	assert.Equal(t, "", result)
}

func TestCIIconForConclusionEmpty(t *testing.T) {
	// Set default provider for testing
	SetIconProvider(&NerdFontV3Provider{})
	result := ciIconForConclusion("")
	assert.Equal(t, "", result)
}

func TestCIIconForConclusionUnknown(t *testing.T) {
	// Set default provider for testing
	SetIconProvider(&NerdFontV3Provider{})
	result := ciIconForConclusion("unknown_status")
	assert.Equal(t, "", result)
}

func TestCIIconForConclusionAllStates(t *testing.T) {
	// Set default provider for testing
	SetIconProvider(&NerdFontV3Provider{})
	tests := []struct {
		conclusion string
		expected   string
	}{
		{"success", ""},
		{"failure", ""},
		{"skipped", ""},
		{"cancelled", ""},
		{"pending", ""},
		{"", ""},
		{"unknown", ""},
		{"random_value", ""},
	}

	for _, tt := range tests {
		t.Run("conclusion_"+tt.conclusion, func(t *testing.T) {
			result := ciIconForConclusion(tt.conclusion)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIconWithSpaceEmpty(t *testing.T) {
	result := iconWithSpace("")
	assert.Empty(t, result)
}

func TestIconWithSpaceWithIcon(t *testing.T) {
	// Test with a non-empty icon (use any non-empty string)
	result := iconWithSpace("test")
	assert.Equal(t, "test ", result)
}

func TestIconWithSpaceMultipleIcons(t *testing.T) {
	tests := []struct {
		icon     string
		expected string
	}{
		{"", ""},
		{"", " "},
		{"󰄱", "󰄱 "},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("icon_%d", i), func(t *testing.T) {
			result := iconWithSpace(tt.icon)
			// Empty icon returns empty string, non-empty returns icon with space
			if tt.icon == "" {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.icon+" ", result)
			}
		})
	}
}
