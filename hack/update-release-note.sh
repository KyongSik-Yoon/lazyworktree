#!/usr/bin/env bash
# Author: Chmouel Boudjnah <chmouel@chmouel.com>
set -euxfo pipefail

last_two_tags=($(git tag -l | tail -2))
old="${last_two_tags[0]}"
new="${last_two_tags[1]}"

git log --patch ${old}..${new} | aichat -m gemini:gemini-3-pro-preview -r gh-release | tee /tmp/gh-release.md

gh release edit -F/tmp/gh-release.md ${new}
