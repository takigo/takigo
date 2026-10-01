# Contributing

[AGENTS.md](AGENTS.md) is the maintained guide to the repository: layout,
coding conventions, how the Go sources map to the Tk sources, and the demo
comparison pipeline. Read it before a first change.

Before sending a change:

```bash
gofmt -l .                # prints nothing
go build ./...
go vet ./...
go fix -diff ./...        # prints nothing
CGO_ENABLED=0 GOOS=windows go vet ./...
xvfb-run -a -s "-screen 0 1280x1024x24 -noreset" go test -race -short ./...
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
