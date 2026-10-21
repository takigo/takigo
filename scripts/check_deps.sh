#!/usr/bin/env bash
# Check that the repository uses the standard library only: no go.sum and
# no module requirement. A dependency is a decision, not an accident, so CI
# fails when one appears without it having been discussed.
set -euo pipefail
cd "$(dirname "$0")/.."

fail=0

if [[ -f go.sum ]]; then
    echo "go.sum exists: the repository uses the standard library only" >&2
    fail=1
fi

if grep -qE '^[[:space:]]*require' go.mod; then
    echo "go.mod requires a module: the repository uses the standard library only" >&2
    fail=1
fi

modules=$(go list -m all 2>/dev/null | tail -n +2)
if [[ -n $modules ]]; then
    echo "go.mod pulls in a module:" >&2
    echo "$modules" >&2
    fail=1
fi

if (( fail )); then
    exit 1
fi
echo "deps: the standard library only"
