#!/usr/bin/env bash
# Run `go mod tidy` in every Go module. Used by Renovate postUpgradeTasks:
# modules linked via `replace` (core, core-contracts/*) change the requirements
# of their consumers, so every go.mod/go.sum must be refreshed together.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

# The Renovate image has no Go: install it with containerbase when missing.
if ! command -v go >/dev/null 2>&1; then
  install-tool golang "$(awk '/^go /{print $2}' core/go.mod)"
fi

find . -name go.mod -not -path '*/vendor/*' -print0 | sort -z |
  while IFS= read -r -d '' mod; do
    dir="$(dirname "$mod")"
    echo "go mod tidy $dir"
    (cd "$dir" && go mod tidy)
  done
