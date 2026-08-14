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

# ---------------------------------------------------------------------------
# extract_demo_title GO_DEMO_NAME
# Reads the Title() option from the Go source to know what window to expect.
# Uses cmd/demotitle (a small Go helper) instead of grep -oP so we don't
# depend on GNU PCRE and so escaped strings / comments are handled correctly.
# ---------------------------------------------------------------------------
extract_demo_title() {
    local demo="$1"
    local src="$PROJECT_DIR/demos/$demo/main.go"
    (cd "$PROJECT_DIR" && go run ./cmd/demotitle "$src") 2>/dev/null || true
}

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

# Get the expected window title from the Go source
TITLE=$(extract_demo_title "$GO_DEMO")
if [[ -z "$TITLE" ]]; then
    echo "Could not extract window title from demos/$GO_DEMO/main.go" >&2
    echo "Set TITLE manually or check the source file." >&2
    exit 1
fi

case "$TYPE" in
    go)  screenshot_go  "$GO_DEMO"  "$OUTPUT" "$TITLE" ;;
    tcl) screenshot_tcl "$TCL_DEMO" "$OUTPUT" "$TITLE" ;;
    *)
        echo "Unknown type: $TYPE (expected go or tcl)" >&2
        exit 1
        ;;
esac
