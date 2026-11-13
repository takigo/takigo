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
# breaking change only asks for a minor version bump. Exit status: 0 when the
# comparison ran (whatever it found), non-zero when any step failed.

set -euo pipefail
cd "$(dirname "$0")/.."

incompatible=()
if [[ ${1:-} == --incompatible ]]; then
    incompatible=(-incompatible)
    shift
fi
base=${1:-$(git describe --tags --abbrev=0 --match 'v[0-9]*' HEAD 2>/dev/null || true)}
new=${2:-HEAD}
# The version of the tool is pinned: x/exp is a frozen module and
# @latest would make the CI job depend on whatever it is today.
APIDIFF_VERSION="${APIDIFF_VERSION:-v0.0.0-20261014143641-f0251eb4aa3e}"

if [[ -z $base ]]; then
    echo "apidiff: no v* tag to compare with"
    exit 0
fi

work=$PWD/tmp/apidiff
apidiff=$work/bin/apidiff
mkdir -p "$work"
[[ -x $apidiff ]] || GOBIN=$work/bin go install golang.org/x/exp/cmd/apidiff@"$APIDIFF_VERSION"

cleanup() {
    git worktree remove --force "$work/old" 2>/dev/null || true
    git worktree remove --force "$work/new" 2>/dev/null || true
}
trap cleanup EXIT
cleanup

# The tool's exit status is 0 whenever it produced a result, with or without
# changes, and non-zero on any error; callers (scripts/release.sh) rely on
# this script failing instead of printing an empty diff. Only grep's own
# "nothing left to print" status is ignored.
show() { # text: print it without the tool's noise and without a stray blank line
    [[ -z $1 ]] || grep -v '^Ignoring internal package' <<<"$1" || true
}

export_api() { # rev dir out
    rm -f "$3" # never fall back on the export of an earlier run
    git worktree add -q --detach "$2" "$1"
    local mod out
    mod=$(cd "$2" && go list -m)
    out=$(cd "$2" && "$apidiff" -m -w "$3" "$mod" 2>&1) || {
        show "$out" >&2
        echo "apidiff: exporting the API of $1 failed" >&2
        return 1
    }
    show "$out"
    [[ -s $3 ]] || { echo "apidiff: export data of $1 is empty" >&2; return 1; }
}

export_api "$base" "$work/old" "$work/old.exp"
export_api "$new" "$work/new" "$work/new.exp"
out=$("$apidiff" -m "${incompatible[@]}" "$work/old.exp" "$work/new.exp" 2>&1) || {
    show "$out" >&2
    echo "apidiff: comparing $base with $new failed" >&2
    exit 1
}
echo "API changes $base -> $new:"
show "$out"
