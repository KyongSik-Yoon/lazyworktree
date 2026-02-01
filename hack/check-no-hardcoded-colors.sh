#!/bin/bash
# Check for hardcoded colours in UI code
# This hook runs before Bash commands to catch colour violations

set -euo pipefail

# Only check Go files in the app directory
UI_DIR="internal/app"

if [[ ! -d "$UI_DIR" ]]; then
    exit 0
fi

# Patterns that indicate hardcoded colours (excluding theme definitions)
COLOUR_PATTERNS=(
    'lipgloss\.Color\("[^"]*"\)'
    '#[0-9a-fA-F]{6}'
    'lipgloss\.AdaptiveColor\{'
)

violations=0

for pattern in "${COLOUR_PATTERNS[@]}"; do
    # Search for hardcoded colours, excluding theme files and test files
    matches=$(grep -rE "$pattern" "$UI_DIR" --include="*.go" \
        --exclude="*_test.go" \
        2>/dev/null | grep -v "theme\." || true)

    if [[ -n "$matches" ]]; then
        if [[ $violations -eq 0 ]]; then
            echo "WARNING: Potential hardcoded colours found in UI code:"
            echo "Use theme fields instead (e.g., theme.Current().Primary)"
            echo ""
        fi
        echo "$matches"
        violations=$((violations + 1))
    fi
done

# This is a warning, not a blocker
exit 0
