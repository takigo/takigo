# takigo

A Go port of the [Tk 9.1](https://www.tcl-lang.org/) GUI toolkit: classic and
themed (ttk) widgets, the pack/grid/place geometry managers, canvas, text,
bindings, dialogs and the window-manager interface, behind a Go API of typed
constructors and functional options.

- **No third-party dependencies**: the module needs only the Go standard library.
- **Three backends**: X11 (Linux/BSD, cgo over Xlib and Xft), macOS (cgo over
  AppKit) and Windows (pure Go).
- **Tracks Tk**: the sources mirror Tk's, and the 67 demos under
  `demos/`, most of them ports of Tk's own, are compared against `wish`
  screenshot by screenshot.

The API is not stable yet; there is no tagged release.

## Hello, world

```go
package main

import (
	"log"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/button"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Hello"))
	if err != nil {
		log.Fatal(err)
	}
	b := button.New(app, "hello",
		button.Text("Hello, world"),
		button.Command(app.Quit),
	)
	pack.Pack(b, pack.PadX(20), pack.PadY(20))
	app.Run()
}
```

## Requirements

Go 1.27 (`GOTOOLCHAIN=auto` fetches it with an older Go).

| Platform | Needs |
|---|---|
| Linux / BSD | a C compiler and `libx11-dev libxft-dev libfontconfig1-dev` |
| macOS | Xcode command line tools |
| Windows | nothing; builds with `CGO_ENABLED=0` |

## Try the demos

```bash
go run ./demos/widget_demo     # launcher for all demos
go run ./demos/button          # a single demo
```

## Documentation

- [docs/tutorial.md](docs/tutorial.md): from an empty directory to a complete application.
- [THREADING.md](THREADING.md): which types are safe to use from which goroutine.
- [AGENTS.md](AGENTS.md): repository layout, conventions, build and test commands.
- [docs/architecture-review.md](docs/architecture-review.md): known design debt.

## Development

```bash
go build ./...
go vet ./...
xvfb-run -a -s "-screen 0 1280x1024x24 -noreset" go test -short ./...
```

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

takigo is distributed under the same BSD-style terms as Tk; see [LICENSE](LICENSE).
