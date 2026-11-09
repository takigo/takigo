---
name: windows-wine-check
description: Use when checking takigo's Windows backend (platform/windows, font/gdi, internal/win32) from Linux — "check it on Windows", "does the demo work under wine", "screenshot the Windows build", "compare Windows and Linux rendering". Cross-builds the demos with CGO_ENABLED=0 GOOS=windows, runs them under wine in Xvfb, grabs screenshots and widget-tree dumps, and lists the distortions that are wine's, not takigo's.
---

# Check the Windows backend under wine

Read `platform/AGENTS.md` first. The Windows backend is pure Go and is
vetted from Linux with `CGO_ENABLED=0 GOOS=windows go vet ./...`; this
skill is for seeing it run.

## Setup

- Reuse the existing wine install and its default prefix (`~/.wine`); do
  not create a new prefix. Check with `wine --version`. If wine is
  missing, stop and tell the user.
- Tools: `xvfb-run`, `xdotool`, ImageMagick `import`, as for the demo
  pipeline (`demos/AGENTS.md`).

## Build and run

```bash
out=tmp/windows; mkdir -p $out
CGO_ENABLED=0 GOOS=windows go build -o $out/ ./demos/...   # systray is Unix-only
xvfb-run -a -s "-screen 0 4000x3000x24 -dpi 96 -noreset" bash -c '
  WINEDEBUG=-all TAKIGO_FREEZE_TIMERS=1 TAKIGO_DUMP_TREE=$PWD/tmp/windows/button_tree.json \
    wine tmp/windows/button.exe & sleep 5
  id=$(xdotool search --name "Button Demonstration" | head -1)   # the demo wm title
  import -window "$id" tmp/windows/button.png
  kill %1'
```

`TAKIGO_DUMP_TREE` takes a Unix path under wine too, so
`go run ./internal/cmd/demodiff` can compare the Windows tree with a Linux
dump of the same demo (`tmp/screenshots/<demo>_go.tree.json` after a
`scripts/demo_compare.sh` run).

To try a backend fix without editing the repo, write an overlay file and
build with `go build -overlay overlay.json`.

## Not takigo bugs

- The wine prefix has no Segoe UI or Consolas: every face is a fallback,
  and `TkFixedFont` text renders proportional.
- No CJK, Devanagari or emoji fallback fonts (boxes in `unicodeout`).
- No font anti-aliasing under Xvfb.

Report a difference as a backend bug only when it is none of these and the
Linux build of the same demo is right.
