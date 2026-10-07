#!/usr/bin/env bash
# apidiff.sh -- Report exported API changes between two revisions
#
# Usage:
#   apidiff.sh [--incompatible] [BASE [NEW]]
#
# BASE defaults to the latest v* tag reachable from HEAD, NEW to HEAD. Both
# are checked out into clean worktrees under tmp/apidiff (so gitignored
# scratch programs don't break package loading), their module APIs are
# written as export data with golang.org/x/exp/cmd/apidiff and compared.
# Internal packages are left out. --incompatible lists only the breaking
# changes. Exits 0 when there is no base tag. Informational: under v0 a
# breaking change only asks for a minor version bump.

set -euo pipefail
cd "$(dirname "$0")/.."

incompatible=()
if [[ ${1:-} == --incompatible ]]; then
    incompatible=(-incompatible)
    shift
fi
base=${1:-$(git describe --tags --abbrev=0 --match 'v[0-9]*' HEAD 2>/dev/null || true)}
new=${2:-HEAD}
if [[ -z $base ]]; then
    echo "apidiff: no v* tag to compare with"
    exit 0
fi

work=$PWD/tmp/apidiff
apidiff=$work/bin/apidiff
mkdir -p "$work"
[[ -x $apidiff ]] || GOBIN=$work/bin go install golang.org/x/exp/cmd/apidiff@latest

cleanup() {
    git worktree remove --force "$work/old" 2>/dev/null || true
    git worktree remove --force "$work/new" 2>/dev/null || true
}
trap cleanup EXIT
cleanup

export_api() { # rev dir out
    git worktree add -q --detach "$2" "$1"
    local mod
    mod=$(cd "$2" && go list -m)
    (cd "$2" && "$apidiff" -m -w "$3" "$mod") 2>&1 | grep -v '^Ignoring internal package' || true
    [[ -s $3 ]]
}

export_api "$base" "$work/old" "$work/old.exp"
export_api "$new" "$work/new" "$work/new.exp"
echo "API changes $base -> $new:"
"$apidiff" -m "${incompatible[@]}" "$work/old.exp" "$work/new.exp" 2>&1 |
    grep -v '^Ignoring internal package' || true
