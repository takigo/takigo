# Takigo

## Project Overview

**Takigo** is a pure Go port of the Tk 9.1 GUI toolkit. It provides a functional-options API for building cross-platform graphical applications, replacing Tk's traditional Tcl command string layer with Go's type system. Currently, it targets X11/Linux via low-level `cgo` Xlib bindings.

The project reimplements Tk's core functionality, geometry managers (pack, grid, place), classic widgets, TTK themed widgets, and a 2D canvas drawing surface.

## Building and Running

**Build:**

*   Build all packages:
    ```bash
    go build ./...
    ```
*   Build the main widget demo launcher:
    ```bash
    go build ./demos/widget_demo
    ```
*   Build a specific demo program:
    ```bash
    go build ./demos/<demo_name>
    ```

**Testing:**

*   Run unit tests (pure Go + cgo):
    ```bash
    go test ./color/ ./option/ ./ttk/ ./bind/ ./widget/text/ ./event/ -v
    ```
*   Run integration tests:
    ```bash
    go test . -v
    ```
    *Note: Integration tests require an active X11 display (e.g., via `DISPLAY` environment variable) or running via `xvfb-run`.* Tests requiring X11 utilize `internal/testutil.RequireDisplay(t)` and `NewTestApp(t)` helpers.

## Development Conventions & Architecture

*   **API Design:** The library relies heavily on functional options (e.g., `AppOption`, `WidgetOption`) for configuration rather than method chaining or command strings.
*   **X11 Interoperability:** Cgo bindings to Xlib are isolated within the `internal/xlib` package, with platform-neutral abstractions defined in `platform/` and `platform/x11/`.
*   **Event Handling:** Takigo features a custom event dispatcher (`event/`). The event loop runs on a select loop in the main goroutine, while a dedicated goroutine reads `XNextEvent` from the X server and channels it to the application.
*   **Dependency Management:** To prevent circular imports, key interfaces like `GeomManager`, `AppContext`, `BindEngine`, etc., are defined in top-level packages such as `window/` and `widget/`.
*   **Theming:** TTK themes are loaded by registering them via blank-import `init()` functions (e.g., `_ "github.com/msorc/takigo/ttk/defaulttheme"`).
*   **Rendering:** Double buffering is utilized for complex widgets (canvas, text, TTK widgets) by rendering to an offscreen pixmap before copying to the window.
