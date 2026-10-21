#!/usr/bin/env bash
# release.sh -- Release master: push, wait for CI, tag and publish
#
# Usage:
#   release.sh [vX.Y.Z]
#
# Without a version, the next one follows from the latest v* tag reachable
# from HEAD: the minor goes up when scripts/apidiff.sh reports breaking API
# changes since that tag, the patch otherwise. RELEASE_DRY=1 prints the
# version that would be released and stops.
#
# Checks that master is clean and up to date with origin, that the tag is
# new (locally and on origin), that gh is logged in and that `make check
# test` passes on a clean checkout of HEAD (SKIP_CHECKS=1 skips that). If
# CHANGELOG.md has no "## vX.Y.Z" heading yet, the "## Unreleased" heading
# becomes "## vX.Y.Z (date)", a fresh empty Unreleased section goes above
# it, and the change is committed.
#
# master is then pushed and the script waits for the CI run of that commit.
# Only when CI passes is the annotated tag (the changelog section is its
# message) created and pushed, proxy.golang.org asked to fetch the version
# (so pkg.go.dev lists it) and a GitHub release created with the same
# notes. When CI fails nothing is tagged: fix it on master and run the
# script again; it finds the untagged changelog section, keeps its version
# and tags the new HEAD.
#
# A pushed tag is permanent: the module proxy caches it. Fix a bad release
# with a new version and a retract directive in go.mod, never by moving
# the tag.

set -euo pipefail
cd "$(dirname "$0")/.."

die() { echo "release: $*" >&2; exit 1; }

semver='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?$'
dry=false
[[ ${RELEASE_DRY:-} == 1 ]] && dry=true
ver=${1:-}
[[ -z $ver || $ver =~ $semver ]] || die "usage: release.sh [vX.Y.Z]"

tag_exists() { git rev-parse -q --verify "refs/tags/$1" >/dev/null; }

if ! $dry; then
    [[ $(git rev-parse --abbrev-ref HEAD) == master ]] || die "not on master"
    [[ -z $(git status --porcelain) ]] || die "working tree is not clean"
    command -v gh >/dev/null || die "gh is needed to wait for CI"
    gh auth status >/dev/null 2>&1 || die "gh is not logged in (gh auth login)"
    git fetch -q origin master --tags
    [[ -z $(git rev-list HEAD..origin/master) ]] || die "master is behind origin/master"
fi

if [[ -z $ver ]]; then
    base=$(git describe --tags --abbrev=0 --match 'v[0-9]*' HEAD 2>/dev/null) ||
        die "no v* tag to count from; pass the version"
    # An untagged section above the base tag's is a release whose CI failed.
    top=$(grep -m1 -oE '^## v[^ ]+' CHANGELOG.md | cut -c4- || true)
    if [[ -n $top && $top != "$base" && $top =~ $semver ]] && ! tag_exists "$top"; then
        ver=$top
        echo "release: $ver (CHANGELOG.md section without a tag)"
    else
        [[ $base =~ $semver ]] || die "cannot parse tag $base"
        maj=${BASH_REMATCH[1]} min=${BASH_REMATCH[2]} pat=${BASH_REMATCH[3]}
        if bash scripts/apidiff.sh --incompatible "$base" | grep -q '^- '; then
            ver=v$maj.$((min + 1)).0
            echo "release: $ver (breaking API changes since $base)"
        else
            ver=v$maj.$min.$((pat + 1))
            echo "release: $ver (no breaking API changes since $base)"
        fi
    fi
fi
[[ $ver =~ $semver ]]

mod=$(go list -m)
major=${BASH_REMATCH[1]}
if ((major >= 2)) && [[ $mod != */v$major ]]; then
    die "$ver needs module path .../v$major, go.mod has $mod"
fi
tag_exists "$ver" && die "tag $ver exists"
$dry && exit 0
[[ -z $(git ls-remote --tags origin "refs/tags/$ver") ]] || die "tag $ver exists on origin"

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

start=$(git rev-parse HEAD)
pushed=false
# Undo the changelog commit when the release fails before master is pushed.
rollback() {
    local rc=$?
    ((rc == 0)) && return
    $pushed && return
    [[ $(git rev-parse HEAD) == "$start" ]] && return
    echo "release: failed; resetting to $(git rev-parse --short "$start")" >&2
    git reset -q --hard "$start"
}

notes=$(mktemp)
trap 'rc=$?; rm -f "$notes"; (exit $rc); rollback' EXIT

if ! grep -q "^## $ver\b" CHANGELOG.md; then
    grep -q '^## Unreleased' CHANGELOG.md || die "CHANGELOG.md has neither '## $ver' nor '## Unreleased'"
    awk -v ver="$ver" -v date="$(date +%Y-%m-%d)" '
        !done && /^## Unreleased/ { print "## Unreleased\n"; print "## " ver " (" date ")"; done=1; next }
        { print }' CHANGELOG.md >CHANGELOG.md.new
    mv CHANGELOG.md.new CHANGELOG.md
    git commit -q -m "Release $ver" CHANGELOG.md
fi

awk -v ver="$ver" '
    $0 ~ "^## " ver "( |$)" { on=1; next }
    on && /^## / { exit }
    on { print }' CHANGELOG.md | sed '/./,$!d' >"$notes"
[[ -s $notes ]] || die "the CHANGELOG.md section for $ver is empty"

sha=$(git rev-parse HEAD)
git push origin master
pushed=true

echo "release: waiting for the CI run of ${sha:0:7}"
run=
for _ in $(seq 60); do
    run=$(gh run list --workflow ci.yml --commit "$sha" --json databaseId -q '.[0].databaseId')
    [[ -n $run ]] && break
    sleep 2
done
[[ -n $run ]] || die "no CI run for $sha appeared; rerun when it has passed"
gh run watch "$run" --exit-status ||
    die "CI failed for ${sha:0:7} (gh run view $run); fix it on master and rerun make release"

{ echo "$ver"; echo; cat "$notes"; } | git tag -a --cleanup=whitespace "$ver" -F -
git push origin "$ver"
echo "Tagged and pushed $ver at ${sha:0:7}."
(cd "$(mktemp -d)" && GOPROXY=https://proxy.golang.org GOFLAGS= go list -m "$mod@$ver")
gh release create "$ver" --title "$ver" --notes-file "$notes"
