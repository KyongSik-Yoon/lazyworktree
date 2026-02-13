package screen

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chmouel/lazyworktree/internal/theme"
)

func TestTextareaScreenType(t *testing.T) {
	s := NewTextareaScreen("Prompt", "Placeholder", "", 120, 40, theme.Dracula(), false)
	if s.Type() != TypeTextarea {
		t.Fatalf("expected TypeTextarea, got %v", s.Type())
	}
}

func TestTextareaScreenCtrlSSubmit(t *testing.T) {
	s := NewTextareaScreen("Prompt", "Placeholder", "hello", 120, 40, theme.Dracula(), false)
	called := false
	var gotValue string
	var gotChecked bool
	s.OnSubmit = func(value string, checked bool) tea.Cmd {
		called = true
		gotValue = value
		gotChecked = checked
		return nil
	}
	s.SetCheckbox("Pinned", true)

	next, _ := s.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if next != nil {
		t.Fatal("expected screen to close on Ctrl+S")
	}
	if !called {
		t.Fatal("expected submit callback to be called")
	}
	if gotValue != "hello" {
		t.Fatalf("expected value %q, got %q", "hello", gotValue)
	}
	if !gotChecked {
		t.Fatal("expected checkbox value to be forwarded")
	}
}

func TestTextareaScreenEnterAddsNewLine(t *testing.T) {
	s := NewTextareaScreen("Prompt", "Placeholder", "hello", 120, 40, theme.Dracula(), false)

	next, _ := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if next == nil {
		t.Fatal("expected screen to stay open on Enter")
	}

	updated := next.(*TextareaScreen)
	if updated.Input.Value() != "hello\n" {
		t.Fatalf("expected newline to be inserted, got %q", updated.Input.Value())
	}
}

func TestTextareaScreenCheckboxToggle(t *testing.T) {
	s := NewTextareaScreen("Prompt", "Placeholder", "", 120, 40, theme.Dracula(), false)
	s.SetCheckbox("Pinned", false)

	// Focus checkbox then toggle it.
	next, _ := s.Update(tea.KeyMsg{Type: tea.KeyTab})
	if next == nil {
		t.Fatal("expected screen to stay open after Tab")
	}
	next, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if next == nil {
		t.Fatal("expected screen to stay open after Space")
	}
	if !s.CheckboxChecked {
		t.Fatal("expected checkbox to be toggled on")
	}
}
