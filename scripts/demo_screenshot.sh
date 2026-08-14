#!/usr/bin/env bash
# demo_screenshot.sh -- Screenshot a single Go or Tcl demo
#
# Usage:
#   demo_screenshot.sh <go_demoname> go  <output.png>
#   demo_screenshot.sh <go_demoname> tcl <output.png> [tcl_demo_name]
#
# Environment:
#   DISPLAY       : X display (default :0)
#   SETTLE_SECS   : wait after window appears before capture (default 1.5)
#   TIMEOUT_SECS  : max wait for window to appear (default 15)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/_lib.sh"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

GO_DEMO="$1"
TYPE="$2"
OUTPUT="$3"
TCL_DEMO="${4:-$GO_DEMO}"

export DISPLAY="${DISPLAY:-:0}"
SETTLE_SECS="${SETTLE_SECS:-1.5}"
TIMEOUT_SECS="${TIMEOUT_SECS:-15}"

# Use the project's Tk 9.1 wish (system wish is 8.6 and incompatible).
# Matches: LD_LIBRARY_PATH="./tcl/unix:./tk/unix" ./tk/unix/wish
WISH="${WISH:-$PROJECT_DIR/tk/unix/wish}"
export LD_LIBRARY_PATH="$PROJECT_DIR/tcl/unix:$PROJECT_DIR/tk/unix${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"

# Make sure wish and the Go binary agree on Xft.dpi. The standalone wish
# reads X resources at startup; merging the same Xft.dpi value the Go
# side used (computed by screenunit.SetScreenDPI) keeps font metrics
# identical between the two screenshots.
XFT_DPI_VAL=""
if command -v xrdb >/dev/null 2>&1; then
    XFT_DPI_VAL=$(xrdb -query 2>/dev/null | awk -F':[[:space:]]*' '/^Xft\.dpi/ {print $2; exit}')
fi
if [[ -n "$XFT_DPI_VAL" ]]; then
    export XFT_DPI="$XFT_DPI_VAL"
fi

# extract_demo_title GO_DEMO and extract_demo_geometry GO_DEMO are defined
# in _lib.sh (sourced above). No local copies needed.

# ---------------------------------------------------------------------------
# wait_for_titled_window TITLE
# Polls wmctrl until a window with the given title appears.
# Prints the hex window ID to stdout.
# ---------------------------------------------------------------------------
wait_for_titled_window() {
    local title="$1"
    local tries=$(( TIMEOUT_SECS * 2 ))
    for _ in $(seq 1 "$tries"); do
        local wid
        wid=$(wmctrl -l 2>/dev/null | grep -F "$title" | awk '{print $1}' | head -1 || true)
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
# Focuses the window, waits for it to settle, then screenshots it.
# ---------------------------------------------------------------------------
capture_window() {
    local wid="$1"
    local out="$2"
    xdotool windowraise  "$wid" 2>/dev/null || true
    xdotool windowfocus --sync "$wid" 2>/dev/null || true
    sleep "$SETTLE_SECS"
    import -window "$wid" "$out"
}

# ---------------------------------------------------------------------------
# screenshot_go GO_DEMO OUTPUT TITLE
# ---------------------------------------------------------------------------
screenshot_go() {
    local demo="$1"
    local out="$2"
    local title="$3"
    local demo_dir="$PROJECT_DIR/demos/$demo"

    if [[ ! -d "$demo_dir" ]]; then
        echo "Go demo directory not found: $demo_dir" >&2
        return 1
    fi

    # Kill any stale instance
    wmctrl -c "$title" 2>/dev/null || true
    sleep 0.3

    local binary
    binary="$(mktemp /tmp/takigo_demo_XXXXXX)"

    echo "  Building demos/$demo..." >&2
    if ! (cd "$PROJECT_DIR" && go build -o "$binary" "./demos/$demo/" 2>&1); then
        rm -f "$binary"
        echo "Build failed for demos/$demo" >&2
        return 1
    fi
    chmod +x "$binary"

    echo "  Running Go demo: $demo  (expecting title: \"$title\")" >&2
    nohup "$binary" > /tmp/takigo_go_out.txt 2>&1 &
    local demo_pid=$!

    local wid
    if ! wid=$(wait_for_titled_window "$title"); then
        kill "$demo_pid" 2>/dev/null || true
        rm -f "$binary"
        return 1
    fi

    echo "  Capturing window $wid..." >&2
    capture_window "$wid" "$out"

    kill "$demo_pid" 2>/dev/null || true
    wait "$demo_pid" 2>/dev/null || true
    rm -f "$binary"
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

    if [[ ! -f "$tcl_file" ]]; then
        echo "Tcl demo not found: $tcl_file" >&2
        return 1
    fi

    # Kill any stale instance
    wmctrl -c "$title" 2>/dev/null || true
    sleep 0.3

    echo "  Running Tcl demo: $demo  (expecting title: \"$title\")" >&2
    nohup "$WISH" "$SCRIPT_DIR/demo_wrapper.tcl" "$demo" > /tmp/takigo_tcl_out.txt 2>&1 &
    local tcl_pid=$!

    local wid
    if ! wid=$(wait_for_titled_window "$title"); then
        kill "$tcl_pid" 2>/dev/null || true
        return 1
    fi

    echo "  Capturing window $wid..." >&2
    capture_window "$wid" "$out"

    kill "$tcl_pid" 2>/dev/null || true
    wait "$tcl_pid" 2>/dev/null || true
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
