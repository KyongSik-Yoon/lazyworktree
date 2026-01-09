package completion

import (
	"strings"
	"testing"

	"github.com/chmouel/lazyworktree/internal/theme"
)

func TestGenerateBash(t *testing.T) {
	script, err := Generate("bash")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify structure
	if !strings.Contains(script, "_lazyworktree_completion") {
		t.Error("missing bash completion function")
	}

	if !strings.Contains(script, "complete -F") {
		t.Error("missing complete command")
	}

	// Verify all flags are present
	for _, flag := range GetFlags() {
		if !strings.Contains(script, "--"+flag.Name) {
			t.Errorf("missing flag --%s", flag.Name)
		}
	}

	// Verify theme values are included
	for _, themeName := range theme.AvailableThemes() {
		if !strings.Contains(script, themeName) {
			t.Errorf("missing theme %s in bash completion", themeName)
		}
	}

	// Verify completion shell values
	for _, shell := range []string{"bash", "zsh", "fish"} {
		if !strings.Contains(script, shell) {
			t.Errorf("missing shell %s in completion flag values", shell)
		}
	}
}

func TestGenerateZsh(t *testing.T) {
	script, err := Generate("zsh")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify structure
	if !strings.Contains(script, "#compdef lazyworktree") {
		t.Error("missing compdef directive")
	}

	if !strings.Contains(script, "_lazyworktree()") {
		t.Error("missing zsh completion function")
	}

	if !strings.Contains(script, "_arguments") {
		t.Error("missing _arguments call")
	}

	// Verify all flags are present
	for _, flag := range GetFlags() {
		if !strings.Contains(script, "--"+flag.Name) {
			t.Errorf("missing flag --%s", flag.Name)
		}
	}

	// Verify theme values are included
	for _, themeName := range theme.AvailableThemes() {
		if !strings.Contains(script, themeName) {
			t.Errorf("missing theme %s in zsh completion", themeName)
		}
	}

	// Verify file completion for paths
	if !strings.Contains(script, "_files") {
		t.Error("missing _files completion for path flags")
	}
}

func TestGenerateFish(t *testing.T) {
	script, err := Generate("fish")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify structure
	if !strings.Contains(script, "complete -c lazyworktree") {
		t.Error("missing fish complete commands")
	}

	// Verify all flags are present
	for _, flag := range GetFlags() {
		if !strings.Contains(script, "-l "+flag.Name) {
			t.Errorf("missing flag -l %s", flag.Name)
		}
	}

	// Verify theme values are included
	for _, themeName := range theme.AvailableThemes() {
		if !strings.Contains(script, themeName) {
			t.Errorf("missing theme %s in fish completion", themeName)
		}
	}

	// Verify shell values for completion flag
	for _, shell := range []string{"bash", "zsh", "fish"} {
		if !strings.Contains(script, shell) {
			t.Errorf("missing shell %s in completion flag values", shell)
		}
	}
}

func TestGenerateUnsupportedShell(t *testing.T) {
	_, err := Generate("powershell")
	if err == nil {
		t.Error("expected error for unsupported shell")
	}

	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("unexpected error message: %v", err)
	}

	if !strings.Contains(err.Error(), "powershell") {
		t.Errorf("error message should include the invalid shell name: %v", err)
	}
}

func TestMetadataThemeSync(t *testing.T) {
	// Ensure theme values are dynamically populated
	flags := GetFlags()
	var themeFlag *FlagInfo
	for i := range flags {
		if flags[i].Name == "theme" {
			themeFlag = &flags[i]
			break
		}
	}

	if themeFlag == nil {
		t.Fatal("theme flag not found in metadata")
	}

	expectedThemes := theme.AvailableThemes()
	if len(themeFlag.Values) != len(expectedThemes) {
		t.Errorf("theme count mismatch: got %d, want %d",
			len(themeFlag.Values), len(expectedThemes))
	}

	// Verify all themes are present
	for _, expected := range expectedThemes {
		found := false
		for _, actual := range themeFlag.Values {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("theme %s not found in metadata", expected)
		}
	}
}

func TestAllFlagsHaveDescriptions(t *testing.T) {
	flags := GetFlags()
	for _, flag := range flags {
		if flag.Description == "" {
			t.Errorf("flag %s missing description", flag.Name)
		}
	}
}

func TestValueFlagsHaveHints(t *testing.T) {
	flags := GetFlags()
	for _, flag := range flags {
		if flag.HasValue && flag.ValueHint == "" {
			t.Errorf("flag %s has value but missing ValueHint", flag.Name)
		}
	}
}

func TestCompletionFlagMetadata(t *testing.T) {
	flags := GetFlags()
	var completionFlag *FlagInfo
	for i := range flags {
		if flags[i].Name == "completion" {
			completionFlag = &flags[i]
			break
		}
	}

	if completionFlag == nil {
		t.Fatal("completion flag not found in metadata")
	}

	// Verify completion flag has correct shell values
	expectedShells := []string{"bash", "zsh", "fish"}
	if len(completionFlag.Values) != len(expectedShells) {
		t.Errorf("completion flag shells count mismatch: got %d, want %d",
			len(completionFlag.Values), len(expectedShells))
	}

	for _, expected := range expectedShells {
		found := false
		for _, actual := range completionFlag.Values {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("shell %s not found in completion flag values", expected)
		}
	}
}
