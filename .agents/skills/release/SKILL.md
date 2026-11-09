---
name: release
description: Use when cutting, preparing or checking a takigo release — "release", "cut v0.5.0", "tag a version", "what version would the next release be", "the release failed in CI", "retract a version". Covers the devel→master merge, make release, RELEASE_DRY, apidiff, the changelog, CI waiting and why a pushed tag is permanent.
---

# Release takigo

Versions are git tags only (`takigo.Version()` reads build info). The
procedure lives in `scripts/release.sh`; `CONTRIBUTING.md` ("Releases")
explains the policy. Never run a release, push or tag without the user
asking for it in this conversation.

## Before

1. Work is on `devel`. `git status` must be clean on both branches.
2. `## Unreleased` in `CHANGELOG.md` lists every user-visible change;
   breaking ones say what callers change. An empty section means there is
   nothing to release.
3. `make apidiff` lists exported API changes since the last tag. Any
   incompatible change makes the release a minor bump (v0.x stays v0).
4. `RELEASE_DRY=1 make release` prints the version it would pick.
   `VERSION=vX.Y.Z` overrides it.

## Release

```bash
git switch master && git merge --no-ff devel   # subject "Merge branch 'devel'"
make release                                   # needs gh logged in
git switch devel && git merge --ff-only master # bring the release commit back
```

`make release` runs `make check test` on a clean checkout, turns
`## Unreleased` into `## vX.Y.Z (date)` and commits, pushes master and
waits for CI. Only when CI passes does it create and push the annotated
tag, ask proxy.golang.org to fetch it and create the GitHub release.

## When it fails

- CI failed: nothing was tagged. Fix on master (or devel and merge), run
  `make release` again; it reuses the untagged changelog section and its
  version.
- A pushed tag is permanent (the module proxy caches it). Never move or
  delete one. A bad release is fixed by a new version plus a `retract`
  directive in `go.mod`.
- Stale local tags that were never pushed (`v0.1.0` with the old
  `github.com/msorc/takigo` module path, `v0.3.5`) must stay unpushed:
  never `git push --tags`.
