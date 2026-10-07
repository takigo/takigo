#!/usr/bin/env bash
# release.sh -- Tag a release from CHANGELOG.md
#
# Usage:
#   release.sh vX.Y.Z [--push]
#
# Checks that master is clean and up to date with origin, that the tag is
# new and that `make check test` passes on a clean checkout of HEAD
# (SKIP_CHECKS=1 skips that). If CHANGELOG.md has no "## vX.Y.Z" heading
# yet, the "## Unreleased" heading becomes "## vX.Y.Z (date)", a fresh
# empty Unreleased section goes above it, and the change is committed. The
# section's text is the annotated tag's message.
#
# Without --push the tag stays local and the push commands are printed.
# --push pushes master and the tag, asks proxy.golang.org to fetch the
# version (so pkg.go.dev lists it), and creates a GitHub release with the
# same notes when gh is installed.
#
# A pushed tag is permanent: the module proxy caches it. Fix a bad release
# with a new version and a retract directive in go.mod, never by moving
# the tag.

set -euo pipefail
cd "$(dirname "$0")/.."

die() { echo "release: $*" >&2; exit 1; }

ver=${1:-}
push=false
[[ ${2:-} == --push ]] && push=true
[[ $ver =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?$ ]] ||
    die "usage: release.sh vX.Y.Z [--push]"

mod=$(go list -m)
major=${BASH_REMATCH[1]}
if ((major >= 2)) && [[ $mod != */v$major ]]; then
    die "$ver needs module path .../v$major, go.mod has $mod"
fi

[[ $(git rev-parse --abbrev-ref HEAD) == master ]] || die "not on master"
[[ -z $(git status --porcelain) ]] || die "working tree is not clean"
git fetch -q origin master --tags
[[ -z $(git rev-list HEAD..origin/master) ]] || die "master is behind origin/master"
git rev-parse -q --verify "refs/tags/$ver" >/dev/null && die "tag $ver exists"

if [[ ${SKIP_CHECKS:-} != 1 ]]; then
    # Check the commit being tagged in a clean worktree: ./... in the working
    # tree also picks up gitignored scratch programs under tmp/.
    check=$PWD/tmp/release-check
    git worktree remove --force "$check" 2>/dev/null || true
    git worktree add -q --detach "$check" HEAD
    trap 'git worktree remove --force "$check"' EXIT
    make -C "$check" check test
    git worktree remove --force "$check"
    trap - EXIT
fi

if ! grep -q "^## $ver\b" CHANGELOG.md; then
    grep -q '^## Unreleased' CHANGELOG.md || die "CHANGELOG.md has neither '## $ver' nor '## Unreleased'"
    awk -v ver="$ver" -v date="$(date +%Y-%m-%d)" '
        !done && /^## Unreleased/ { print "## Unreleased\n"; print "## " ver " (" date ")"; done=1; next }
        { print }' CHANGELOG.md >CHANGELOG.md.new
    mv CHANGELOG.md.new CHANGELOG.md
    git commit -q -m "Release $ver" CHANGELOG.md
fi

notes=$(mktemp)
trap 'rm -f "$notes"' EXIT
awk -v ver="$ver" '
    $0 ~ "^## " ver "( |$)" { on=1; next }
    on && /^## / { exit }
    on { print }' CHANGELOG.md | sed '/./,$!d' >"$notes"
[[ -s $notes ]] || die "the CHANGELOG.md section for $ver is empty"

{ echo "$ver"; echo; cat "$notes"; } | git tag -a --cleanup=whitespace "$ver" -F -
echo "Tagged $ver at $(git rev-parse --short HEAD)."

if ! $push; then
    echo "Push it with: git push origin master $ver"
    echo "or rerun with --push (also registers it with proxy.golang.org)."
    exit 0
fi

git push origin master "$ver"
(cd "$(mktemp -d)" && GOPROXY=https://proxy.golang.org GOFLAGS= go list -m "$mod@$ver")
if command -v gh >/dev/null; then
    gh release create "$ver" --title "$ver" --notes-file "$notes"
fi
