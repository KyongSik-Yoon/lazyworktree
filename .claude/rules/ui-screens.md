---
paths:
  - "internal/app/screen/**/*.go"
  - "internal/app/app_screens*.go"
---

# Screen Component Rules

When working with UI screen components:

- **Theme colours only**: Use theme fields for all colours, never hardcode. Access via `theme.Current()`.
- **tea.Model interface**: Implement `Init()`, `Update()`, and `View()` methods.
- **Screen Manager**: Add screens to the Manager stack via `Push()` and `Pop()`.
- **Test coverage**: Create a matching `*_test.go` file for new screens.
- **Keybinding consistency**: Register keybindings and ensure they appear in help screen, README, and man page.
- **Lipgloss styles**: Define styles using theme fields, not direct colour values.
