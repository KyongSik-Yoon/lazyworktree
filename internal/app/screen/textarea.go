package screen

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chmouel/lazyworktree/internal/theme"
)

// TextareaScreen displays a modal multiline input with optional checkbox support.
type TextareaScreen struct {
	Prompt      string
	Placeholder string
	Input       textarea.Model
	ErrorMsg    string
	Thm         *theme.Theme
	ShowIcons   bool

	// Validation
	Validate func(string) string

	// Callbacks
	OnSubmit         func(value string, checked bool) tea.Cmd
	OnCancel         func() tea.Cmd
	OnCheckboxToggle func(checked bool) tea.Cmd

	// Checkbox support
	CheckboxEnabled bool
	CheckboxChecked bool
	CheckboxFocused bool
	CheckboxLabel   string

	boxWidth  int
	boxHeight int
}

// NewTextareaScreen creates a multiline input modal sized relative to the terminal.
func NewTextareaScreen(prompt, placeholder, value string, maxWidth, maxHeight int, thm *theme.Theme, showIcons bool) *TextareaScreen {
	width := 90
	height := 22
	if maxWidth > 0 {
		width = clampInt(int(float64(maxWidth)*0.75), 70, 110)
	}
	if maxHeight > 0 {
		height = clampInt(int(float64(maxHeight)*0.75), 16, 36)
	}

	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.SetValue(value)
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetWidth(width - 8)
	ta.SetHeight(clampInt(height-11, 6, 24))
	ta.Focus()

	focused, _ := textarea.DefaultStyles()
	focused.Base = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(thm.Border).
		Padding(0, 1)
	focused.Text = lipgloss.NewStyle().Foreground(thm.TextFg)
	focused.Prompt = lipgloss.NewStyle().Foreground(thm.Accent)
	focused.Placeholder = lipgloss.NewStyle().Foreground(thm.MutedFg).Italic(true)
	focused.CursorLine = lipgloss.NewStyle().Foreground(thm.TextFg)
	focused.EndOfBuffer = lipgloss.NewStyle().Foreground(thm.MutedFg)
	blurred := focused
	blurred.Base = blurred.Base.BorderForeground(thm.BorderDim)
	ta.FocusedStyle = focused
	ta.BlurredStyle = blurred

	return &TextareaScreen{
		Prompt:      prompt,
		Placeholder: placeholder,
		Input:       ta,
		Thm:         thm,
		ShowIcons:   showIcons,
		boxWidth:    width,
		boxHeight:   height,
	}
}

// SetValidation sets a validation function that returns an error message.
func (s *TextareaScreen) SetValidation(fn func(string) string) {
	s.Validate = fn
}

// SetCheckbox enables a checkbox with the given label and default state.
func (s *TextareaScreen) SetCheckbox(label string, defaultChecked bool) {
	s.CheckboxEnabled = true
	s.CheckboxLabel = label
	s.CheckboxChecked = defaultChecked
	s.CheckboxFocused = false
}

// Type returns the screen type.
func (s *TextareaScreen) Type() Type {
	return TypeTextarea
}

// Update handles keyboard input for the textarea screen.
// Returns nil to signal the screen should be closed.
func (s *TextareaScreen) Update(msg tea.KeyMsg) (Screen, tea.Cmd) {
	var cmd tea.Cmd
	keyStr := msg.String()

	switch keyStr {
	case keyTab, keyShiftTab:
		if s.CheckboxEnabled {
			s.CheckboxFocused = !s.CheckboxFocused
			if s.CheckboxFocused {
				s.Input.Blur()
			} else {
				cmd = s.Input.Focus()
			}
		}
		return s, cmd

	case " ":
		if s.CheckboxEnabled && s.CheckboxFocused {
			s.CheckboxChecked = !s.CheckboxChecked
			if s.OnCheckboxToggle != nil {
				return s, s.OnCheckboxToggle(s.CheckboxChecked)
			}
			return s, nil
		}

	case "ctrl+s":
		value := s.Input.Value()
		if s.Validate != nil {
			if errMsg := strings.TrimSpace(s.Validate(value)); errMsg != "" {
				s.ErrorMsg = errMsg
				return s, nil
			}
		}
		s.ErrorMsg = ""
		if s.OnSubmit != nil {
			cmd = s.OnSubmit(value, s.CheckboxChecked)
			if s.ErrorMsg != "" {
				return s, cmd
			}
		}
		return nil, cmd

	case keyEsc, keyCtrlC:
		if s.OnCancel != nil {
			return nil, s.OnCancel()
		}
		return nil, nil
	}

	if s.CheckboxFocused {
		return s, nil
	}

	s.Input, cmd = s.Input.Update(msg)
	return s, cmd
}

// View renders the multiline input screen.
func (s *TextareaScreen) View() string {
	width := s.boxWidth
	height := s.boxHeight

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.Thm.Accent).
		Padding(1, 2).
		Width(width).
		Height(height)

	promptStyle := lipgloss.NewStyle().
		Foreground(s.Thm.Accent).
		Bold(true).
		Width(width - 6).
		Align(lipgloss.Center)

	footerStyle := lipgloss.NewStyle().
		Foreground(s.Thm.MutedFg).
		Width(width - 6).
		Align(lipgloss.Center)

	contentLines := []string{
		promptStyle.Render(s.Prompt),
	}

	if s.CheckboxEnabled {
		checkbox := "[ ] "
		if s.CheckboxChecked {
			checkbox = "[x] "
		}
		checkboxStyle := lipgloss.NewStyle().
			Width(width - 6)
		if s.CheckboxFocused {
			checkboxStyle = checkboxStyle.
				Background(s.Thm.Accent).
				Foreground(s.Thm.AccentFg).
				Padding(0, 1).
				Bold(true)
		} else {
			checkboxStyle = checkboxStyle.Foreground(s.Thm.Accent)
		}
		contentLines = append(contentLines, checkboxStyle.Render(checkbox+s.CheckboxLabel))
	}

	contentLines = append(contentLines, s.Input.View())

	if s.ErrorMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(s.Thm.ErrorFg).
			Width(width - 6).
			Align(lipgloss.Center)
		contentLines = append(contentLines, errorStyle.Render(s.ErrorMsg))
	}

	footerText := "Ctrl+S save • Esc cancel • Enter newline"
	if s.CheckboxEnabled {
		footerText = "Tab switch focus • Space toggle • Ctrl+S save • Esc cancel"
	}
	contentLines = append(contentLines, footerStyle.Render(footerText))

	return boxStyle.Render(strings.Join(contentLines, "\n\n"))
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
