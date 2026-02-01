---
name: quick-explore
description: Fast read-only codebase exploration using a cheaper model
model: haiku
tools:
  - Read
  - Grep
  - Glob
  - LS
---

# Quick Explore Agent

Search the lazyworktree codebase for: $ARGUMENTS

## Instructions

1. Use Glob to find relevant files by pattern
2. Use Grep to search for keywords and patterns
3. Use Read to examine file contents
4. Use LS to explore directory structures

## Output

Return:
- File paths matching the search criteria
- Relevant code snippets
- Brief context about what each match represents

This is a read-only exploration. Do not suggest or make any modifications.
