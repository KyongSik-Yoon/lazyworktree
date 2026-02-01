---
paths:
  - "**/*_test.go"
---

# Test Rules

When writing or modifying tests:

- **Table-driven tests**: Prefer table-driven tests for multiple input scenarios.
- **Descriptive names**: Use clear, descriptive test function and subtest names.
- **No mocks unless required**: Avoid mocks unless explicitly requested or required by existing patterns.
- **Coverage focus**: Aim for high coverage on new functionality.
- **Run specific tests**: Default to running only the tests you changed or added.
- **Testify assertions**: Use `github.com/stretchr/testify/assert` and `require` packages.
- **Golden files**: For complex output, consider golden file testing patterns.
