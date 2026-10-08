# Contributing

[AGENTS.md](AGENTS.md) is the maintained guide to the repository: layout,
build and test commands, the rules that apply everywhere, and an index of
the per-area guides (`widget/AGENTS.md`, `ttk/AGENTS.md`,
`platform/AGENTS.md`, `demos/AGENTS.md`). Coding agents load them by
themselves; read the root and the guide of the area you change before a
first change. `scripts/check_docs.sh` (part of `make check`) fails when one
of them names a path that no longer exists.

Before sending a change:

```bash
gofmt -l .                # prints nothing
go build ./...
go vet ./...
go fix -diff ./...        # prints nothing
CGO_ENABLED=0 GOOS=windows go vet ./...
bash scripts/check_docs.sh    # the agent guides name existing paths
go test -race -short ./...   # display tests start their own Xvfb
```

CI runs the same checks. `golangci-lint` reports only on lines a change
touches; do not mass-fix unrelated warnings.

Rules that are easy to miss:

- Standard library only. Do not add modules.
- When changing widget behaviour, check the Tk source it ports (`tk/generic`,
  `tk/library`) first.
- A change to a core package can shift demo screenshots: run
  `bash scripts/demo_batch.sh --retake` and `bash scripts/demo_gate.sh`.
- UI state belongs to the event-loop goroutine; see [THREADING.md](THREADING.md).
- A user-visible change gets an entry under `## Unreleased` in
  [CHANGELOG.md](CHANGELOG.md); a breaking one says what callers change.

## Releases

Versions are git tags (`vX.Y.Z`); nothing in the source carries the
number, and `takigo.Version()` reads it from the build info.

- Until the API settles, releases stay at v0: a breaking change bumps the
  minor version (v0.2 → v0.3), anything else the patch. v1.0.0 is a promise
  of compatibility; after it a breaking change needs a `/v2` module path.
- `make apidiff` (or `bash scripts/apidiff.sh [BASE]`) lists the exported
  API changes since the last tag; CI shows the same list in the `api` job.
- `make release VERSION=v0.2.0` checks the tree, runs `make check test`,
  turns the changelog's Unreleased section into the v0.2.0 section, commits
  it and tags it with that section as the message. `PUSH=1` also pushes the
  tag, registers it with proxy.golang.org and creates a GitHub release.
- A pushed tag is permanent (the module proxy caches it). Fix a bad release
  with a new version plus a `retract` line in `go.mod`; never move or delete
  a pushed tag.
