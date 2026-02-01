---
paths:
  - "internal/theme/**/*.go"
---

# Theme Rules

When working with theme code:

- **Complete field sets**: Every theme must define all required colour fields.
- **Contrast compliance**: Ensure sufficient contrast between foreground and background colours.
- **Lipgloss integration**: Theme fields must be compatible with lipgloss styling.
- **No hardcoded colours**: All colour references in UI code must use theme fields.
- **Test all themes**: Verify each theme renders correctly in tests.
- **Catppuccin base**: Default themes follow the Catppuccin colour palette conventions.
