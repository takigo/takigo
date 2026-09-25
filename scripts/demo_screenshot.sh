#!/usr/bin/env bash
# demo_screenshot.sh -- Screenshot a single Go or Tcl demo
#
# Usage:
#   demo_screenshot.sh <go_demoname> go  <output.png>
#   demo_screenshot.sh <go_demoname> tcl <output.png> [tcl_demo_name]
#
# Environment:
#   DISPLAY       : X display (default :0)
#   HEADLESS      : 1 = run under xvfb-run (also auto-fallback when DISPLAY unset)
#   SETTLE_SECS   : max wait for the window content to stop changing (default 5)
#   TIMEOUT_SECS  : max wait for window to appear (default 15)
#   PIN_FONTS     : 1 (default) = private DejaVu-only fontconfig, see pin_fonts
#   TAKIGO_FREEZE_TIMERS : 1 (default) = drop positive-delay timers on both sides

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/_lib.sh"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Re-exec under xvfb-run when headless (before touching DISPLAY).
maybe_xvfb "$@"

GO_DEMO="$1"
TYPE="$2"
OUTPUT="$3"
TCL_DEMO="${4:-$GO_DEMO}"

export DISPLAY="${DISPLAY:-:0}"
SETTLE_SECS="${SETTLE_SECS:-5}"
TIMEOUT_SECS="${TIMEOUT_SECS:-15}"
export TAKIGO_FREEZE_TIMERS="${TAKIGO_FREEZE_TIMERS:-1}"
pin_fonts

# Use the project's Tk 9.1 wish (system wish is 8.6 and incompatible).
# Matches: LD_LIBRARY_PATH="./tcl/unix:./tk/unix" ./tk/unix/wish
WISH="${WISH:-$PROJECT_DIR/tk/unix/wish}"
export LD_LIBRARY_PATH="$PROJECT_DIR/tcl/unix:$PROJECT_DIR/tk/unix${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"

# Make sure wish and the Go binary agree on Xft.dpi. The standalone wish
# reads X resources at startup; merging the same Xft.dpi value the Go
# side used (computed by screenunit.SetScreenDPI) keeps font metrics
# identical between the two screenshots. Under Xvfb there is no xrdb
# database, so fall back to the 96dpi the server was started with.
XFT_DPI_VAL=""
if command -v xrdb >/dev/null 2>&1; then
    XFT_DPI_VAL=$(xrdb -query 2>/dev/null | awk -F':[[:space:]]*' '/^Xft\.dpi/ {print $2; exit}')
fi
if [[ -n "$XFT_DPI_VAL" ]]; then
    export XFT_DPI="$XFT_DPI_VAL"
elif [[ -n "${_XVFB_RUNNING:-}" ]]; then
    export XFT_DPI=96
fi

# ---------------------------------------------------------------------------
# wait_for_titled_window TITLE
# Polls xdotool until a window with the exact given title appears. Uses
# xdotool search (walks the X tree, no window manager required) rather than
# wmctrl, so it works under a bare Xvfb display. Prints the window ID.
# ---------------------------------------------------------------------------
wait_for_titled_window() {
    local title="$1"
    local tries=$(( TIMEOUT_SECS * 2 ))
    for _ in $(seq 1 "$tries"); do
        local wid
        wid=$(find_windows_exact "$title" | head -1)
        if [[ -n "$wid" ]]; then
            echo "$wid"
            return 0
        fi
        sleep 0.5
    done
    echo "Timeout waiting for window titled: $title" >&2
    return 1
}

# ---------------------------------------------------------------------------
# capture_window HEX_WID OUTPUT
# Raises/focuses the window, then grabs it repeatedly until two consecutive
# grabs have the same pixel signature (or SETTLE_SECS elapses, in which case
# the last grab is kept and a warning is printed).
# ---------------------------------------------------------------------------
capture_window() {
    local wid="$1"
    local out="$2"
    local tmp="${out%.png}.settle.png"
    local prev="" sig="" deadline
    xdotool windowraise  "$wid" 2>/dev/null || true
    xdotool windowfocus --sync "$wid" 2>/dev/null || true
    sleep 0.3
    deadline=$(( $(date +%s%N) + $(awk -v s="$SETTLE_SECS" 'BEGIN { printf "%d", s * 1e9 }') ))
    while :; do
        import -window "$wid" "$tmp"
        sig=$(magick identify -format '%#' "$tmp" 2>/dev/null) || sig=""
        if [[ -n "$sig" && "$sig" == "$prev" ]]; then
            break
        fi
        if (( $(date +%s%N) >= deadline )); then
            echo "  WARNING: window did not settle within ${SETTLE_SECS}s" >&2
            break
        fi
        prev="$sig"
        sleep 0.25
    done
    mv "$tmp" "$out"
}

# ---------------------------------------------------------------------------
# screenshot_go GO_DEMO OUTPUT TITLE
# ---------------------------------------------------------------------------
screenshot_go() {
    local demo="$1"
    local out="$2"
    local title="$3"
    local demo_dir="$PROJECT_DIR/demos/$demo"
    local binary demo_pid wid

    if [[ ! -d "$demo_dir" ]]; then
        echo "Go demo directory not found: $demo_dir" >&2
        return 1
    fi

    # Kill any stale instance (windowkill works without a window manager).
    for sid in $(find_windows_exact "$title"); do xdotool windowkill "$sid" 2>/dev/null; done
    sleep 0.3

    binary="$(mktemp /tmp/takigo_demo_XXXXXX)"

    echo "  Building demos/$demo..." >&2
    if ! (cd "$PROJECT_DIR" && go build -o "$binary" "./demos/$demo/" 2>&1); then
        rm -f "$binary"
        echo "Build failed for demos/$demo" >&2
        return 1
    fi
    chmod +x "$binary"

    trap 'if [[ -n "${demo_pid:-}" ]]; then kill "$demo_pid" 2>/dev/null || true; wait "$demo_pid" 2>/dev/null || true; fi; rm -f "$binary"' RETURN

    echo "  Running Go demo: $demo  (expecting title: \"$title\")" >&2
    nohup "$binary" > /tmp/takigo_go_out.txt 2>&1 &
    demo_pid=$!

    if ! wid=$(wait_for_titled_window "$title"); then
        return 1
    fi

    # Confirm the demo is still alive before capture (crash-after-map would
    # otherwise yield a stale/blank grab).
    if ! kill -0 "$demo_pid" 2>/dev/null; then
        echo "  Demo exited before capture; log:" >&2
        tail -5 /tmp/takigo_go_out.txt >&2 || true
        return 1
    fi

    echo "  Capturing window $wid..." >&2
    capture_window "$wid" "$out"

    echo "  Saved: $out" >&2
}

# ---------------------------------------------------------------------------
# screenshot_tcl TCL_DEMO OUTPUT TITLE
# ---------------------------------------------------------------------------
screenshot_tcl() {
    local demo="$1"
    local out="$2"
    local title="$3"
    local tcl_file="$PROJECT_DIR/tk/library/demos/${demo}.tcl"
    local tcl_pid wid

    if [[ ! -f "$tcl_file" ]]; then
        echo "Tcl demo not found: $tcl_file" >&2
        return 1
    fi

    for sid in $(find_windows_exact "$title"); do xdotool windowkill "$sid" 2>/dev/null; done
    sleep 0.3

    trap 'if [[ -n "${tcl_pid:-}" ]]; then kill "$tcl_pid" 2>/dev/null || true; wait "$tcl_pid" 2>/dev/null || true; fi' RETURN

    echo "  Running Tcl demo: $demo  (expecting title: \"$title\")" >&2
    nohup "$WISH" "$SCRIPT_DIR/demo_wrapper.tcl" "$demo" > /tmp/takigo_tcl_out.txt 2>&1 &
    tcl_pid=$!

    if ! wid=$(wait_for_titled_window "$title"); then
        return 1
    fi

    if ! kill -0 "$tcl_pid" 2>/dev/null; then
        echo "  Tcl demo exited before capture; log:" >&2
        tail -5 /tmp/takigo_tcl_out.txt >&2 || true
        return 1
    fi

    echo "  Capturing window $wid..." >&2
    capture_window "$wid" "$out"
    echo "  Saved: $out" >&2
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
mkdir -p "$(dirname "$OUTPUT")"

# Get the expected window title and geometry from the Go source
TITLE=$(extract_demo_title "$GO_DEMO")
if [[ -z "$TITLE" ]]; then
    echo "Could not extract window title from demos/$GO_DEMO/main.go" >&2
    echo "Set TITLE manually or check the source file." >&2
    exit 1
fi
DEMO_GEOMETRY=$(extract_demo_geometry "$GO_DEMO")
export DEMO_GEOMETRY

case "$TYPE" in
    go)  screenshot_go  "$GO_DEMO"  "$OUTPUT" "$TITLE" ;;
    tcl) screenshot_tcl "$TCL_DEMO" "$OUTPUT" "$TITLE" ;;
    *)
        echo "Unknown type: $TYPE (expected go or tcl)" >&2
        exit 1
        ;;
esac
