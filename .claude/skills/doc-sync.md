---
name: doc-sync
description: Check documentation consistency across README, man page, and help screen
---

# Documentation Sync Check

Compare documentation sources for consistency.

## README keybindings section
!`sed -n '/## Key/,/^##/p' README.md | head -50`

## Man page keybindings
!`sed -n '/KEYBINDINGS/,/^$/p' lazyworktree.1 | head -50`

## Help screen text
!`grep -A 100 'helpText' internal/app/screen/help.go | head -60`

---

Compare these three sources and identify:
1. Missing keybindings in any source
2. Inconsistent descriptions between sources
3. Outdated or incorrect information
4. British spelling violations
5. Specific recommendations to bring all sources into sync
