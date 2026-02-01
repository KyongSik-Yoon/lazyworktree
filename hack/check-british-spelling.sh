#!/bin/bash
# Check for American spellings in documentation
# Validates British spelling conventions

set -eo pipefail

# Files to check
DOC_FILES=(
    "README.md"
    "lazyworktree.1"
    "internal/app/screen/help.go"
)

# American -> British spelling pairs (space-separated: "american british")
SPELLINGS=(
    "color colour"
    "colors colours"
    "behavior behaviour"
    "behaviors behaviours"
    "initialize initialise"
    "initialized initialised"
    "organize organise"
    "organized organised"
    "customize customise"
    "customized customised"
    "favorite favourite"
    "favorites favourites"
    "center centre"
    "centers centres"
    "canceled cancelled"
    "traveling travelling"
    "dialog dialogue"
    "dialogs dialogues"
)

violations=0

for file in "${DOC_FILES[@]}"; do
    if [[ ! -f "$file" ]]; then
        continue
    fi

    # Strip fenced code blocks to avoid matching examples.
    cleaned=$(awk '
        /^```/ { in_block = !in_block; next }
        !in_block { print }
    ' "$file")

    for pair in "${SPELLINGS[@]}"; do
        american="${pair%% *}"
        british="${pair##* }"

        # Case-insensitive search, excluding code blocks and URLs
        matches=$(printf '%s\n' "$cleaned" | grep -inE "\b${american}\b" 2>/dev/null \
            | grep -v 'http' \
            | grep -v '//' || true)

        if [[ -n "$matches" ]]; then
            if [[ $violations -eq 0 ]]; then
                echo "WARNING: American spellings found (use British spelling):"
                echo ""
            fi
            echo "In $file: '$american' should be '$british'"
            echo "$matches" | head -3
            echo ""
            violations=$((violations + 1))
        fi
    done
done

# This is a warning, not a blocker
exit 0
