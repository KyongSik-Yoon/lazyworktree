---
name: test-finder
description: Find tests related to a feature using a cheaper model
model: haiku
tools:
  - Read
  - Grep
  - Glob
---

# Test Finder Agent

Find all test files and test functions related to: $ARGUMENTS

## Search Strategy

1. Use Glob to find `*_test.go` files
2. Use Grep to search for test function names matching the feature
3. Use Read to examine test file contents

## Output

Report:
- Test file paths
- Test function names (functions starting with `Test`)
- Brief description of what each test covers
- Any gaps in test coverage for the feature
