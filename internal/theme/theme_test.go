package theme

import "testing"

func TestGetTheme(t *testing.T) {
	for _, name := range AvailableThemes() {
		got := GetTheme(name)
		if got == nil {
			t.Fatalf("expected theme for %q", name)
		}
	}

	fallback := GetTheme("unknown")
	if fallback.Background != Dracula().Background {
		t.Fatalf("expected Dracula fallback, got %q", fallback.Background)
	}
}

func TestAvailableThemesIncludesDefaults(t *testing.T) {
	themes := AvailableThemes()
	required := map[string]bool{
		"dracula":          false,
		"narna":            false,
		"clean-light":      false,
		"catppuccin-latte": false,
		"rose-pine-dawn":   false,
		"one-light":        false,
		"everforest-light": false,
		"solarized-dark":   false,
		"solarized-light":  false,
		"gruvbox-dark":     false,
		"gruvbox-light":    false,
		"nord":             false,
		"monokai":          false,
		"catppuccin-mocha": false,
	}

	for _, name := range themes {
		if _, ok := required[name]; ok {
			required[name] = true
		}
	}

	for name, seen := range required {
		if !seen {
			t.Fatalf("expected theme %q to be available", name)
		}
	}
}

func TestIsLight(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{CleanLightName, true},
		{CatppuccinLatteName, true},
		{RosePineDawnName, true},
		{OneLightName, true},
		{EverforestLightName, true},
		{SolarizedLightName, true},
		{GruvboxLightName, true},
		{DraculaName, false},
		{NarnaName, false},
		{SolarizedDarkName, false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLight(tt.name); got != tt.want {
				t.Errorf("IsLight(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestDefaults(t *testing.T) {
	if got := DefaultDark(); got != DraculaName {
		t.Errorf("DefaultDark() = %q, want %q", got, DraculaName)
	}
	if got := DefaultLight(); got != CleanLightName {
		t.Errorf("DefaultLight() = %q, want %q", got, CleanLightName)
	}
}
