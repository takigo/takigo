# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the **Tk 9.1a1** source distribution — a cross-platform GUI toolkit implemented with the Tcl scripting language. It requires Tcl 9.0+. The source lives under `tk/`.

## Build Commands (Linux/Unix)

Tk requires a built Tcl installation. Build Tcl first, then:

```bash
cd tk/unix
./configure --with-tcl=<path-to-tcl-build>
make binaries libraries    # build shared library and wish binary
make tktest                # build test harness
make install
```

Key configure options: `--enable-symbols[=mem|all]`, `--enable-shared/--disable-shared`, `--enable-xft`, `--enable-aqua` (macOS).

## Running Tests

```bash
cd tk/unix
make test              # runs both classic and TTK tests
make test-classic      # classic widget tests only
make test-ttk          # TTK themed widget tests only
```

Tests require a display server. On headless Linux, use `xvfb-run`. Set `ERROR_ON_FAILURES=1` for non-zero exit on failure.

Tests use Tcl's `tcltest` 2.2 framework. Test files are in `tk/tests/` (classic) and `tk/tests/ttk/` (themed widgets). Entry point: `tk/tests/all.tcl`.

## Architecture

### Platform Abstraction

The codebase is split into cross-platform and platform-specific layers:

- **`generic/`** — Cross-platform C implementation (~84 .c files, ~30 headers). All widget logic, event binding, image handling, canvas, text engine, and configuration live here.
- **`generic/ttk/`** — TTK (Themed Tk) subsystem with modern widget implementations and theme engines (clam, classic, default, alt). Separate from classic widgets.
- **`unix/`** — X11 windowing system implementation
- **`win/`** — Win32 native API implementation
- **`macosx/`** — Cocoa/Aqua native implementation
- **`xlib/`** — X11 compatibility stubs for non-X11 platforms

### Key Subsystems

| Subsystem | Core Files | Description |
|-----------|-----------|-------------|
| Widget core | `tkWindow.c`, `tkCmds.c`, `tkConfig.c` | Window management, Tk commands, option configuration |
| Event/Binding | `tkBind.c` (largest file ~160KB), `tkEvent.c` | Event loop integration, event-to-script binding |
| Geometry | `tkPack.c`, `tkGrid.c`, `tkPlace.c` | Layout managers |
| Canvas | `tkCanvas.c`, `tkCanv*.c`, `tkRectOval.c` | Vector drawing widget with arc/line/polygon/text/image items |
| Text | `tkText.c`, `tkTextDisp.c`, `tkTextBTree.c`, `tkTextIndex.c`, `tkTextMark.c`, `tkTextTag.c` | Multi-line text widget with B-tree storage, tags, marks |
| Image | `tkImgPhoto.c`, `tkImgPNG.c`, `tkImgGIF.c`, `tkImgSVGnano.c` | Built-in image format handlers (PNG, GIF, PPM, SVG via nanosvg) |
| TTK themes | `generic/ttk/ttkClamTheme.c`, `ttkClassicTheme.c`, `ttkDefaultTheme.c` | Theme engine implementations |

### Library Scripts

`library/` contains Tcl-side implementations: widget bindings (`button.tcl`, `entry.tcl`, `text.tcl`, etc.), common dialogs (`msgbox.tcl`, `clrpick.tcl`, `fontchooser.tcl`), TTK themes and widgets (`library/ttk/`), demos, images, and i18n message catalogs.

### Build Systems

Three parallel build systems exist:
- **Autoconf** (`unix/configure.ac` + `Makefile.in`) — Linux, macOS prefix builds, MinGW
- **NMake** (`win/makefile.vc` + `rules.vc`) — Windows MSVC
- **GNUmakefile** (`macosx/GNUmakefile`) — macOS framework builds

### Stub Library

`tkStubInit.c` / `tkStubLib.c` provide the stubs mechanism for binary-compatible extension loading across Tk versions.
