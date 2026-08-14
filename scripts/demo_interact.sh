#!/usr/bin/env bash
# scripts/demo_interact.sh -- Drive a demo through xdotool events and
# compare before/after screenshots.
#
# Usage:
#   demo_interact.sh <go_demo> [event-flags...] [--diff] [--keep-running]
#
# Event flags (repeatable, run in given order):
#   --wait MS            Sleep MS milliseconds before next event (default 200)
#   --key "<sequence>"   xdotool key — e.g. "Tab", "Return", "ctrl+a"
#   --type "text"        xdotool type — literal text
#   --click X,Y          Left-click at (X,Y) relative to the window
#
# Output flags:
#   --before PATH        Override before-screenshot path
#   --after  PATH        Override after-screenshot path
#   --diff               Compute normalized MAE between before/after
#   --keep-running       Leave the demo window open after capture
#
# If no event flags are given, only the before/after screenshots are
# taken — useful for sanity-checking the window focus + capture path.
#
# Examples:
#   bash scripts/demo_interact.sh label
#   bash scripts/demo_interact.sh entry1 --type "hello" --key Return --diff
#   bash scripts/demo_interact.sh button --click 100,200 --diff

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/_lib.sh"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

DEMO="${1:-}"
if [[ -z "$DEMO" ]]; then
    echo "Usage: $0 <go_demo> [--wait MS] [--key <seq>] [--type <text>] [--click X,Y]" >&2
    echo "       [--before PATH] [--after PATH] [--diff] [--keep-running]" >&2
    exit 1
fi
shift

# Defaults.
BEFORE="$SS_DIR/${DEMO}_interact_before.png"
AFTER="$SS_DIR/${DEMO}_interact_after.png"
DIFF=0
KEEP_RUNNING=0
DEFAULT_WAIT_MS=200

# Events stored in order in the EVENTS array; each entry is
# "type|arg1|arg2|..." with type in {wait,key,type,click}.
EVENTS=()

add_event() {
    EVENTS+=("$1")
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --wait)
            [[ $# -ge 2 ]] || { echo "--wait requires MS" >&2; exit 1; }
            add_event "wait|$2"
            shift 2
            ;;
        --key)
            [[ $# -ge 2 ]] || { echo "--key requires <sequence>" >&2; exit 1; }
            add_event "key|$2"
            shift 2
            ;;
        --type)
            [[ $# -ge 2 ]] || { echo "--type requires <text>" >&2; exit 1; }
            add_event "type|$2"
            shift 2
            ;;
        --click)
            [[ $# -ge 2 ]] || { echo "--click requires X,Y" >&2; exit 1; }
            add_event "click|$2"
            shift 2
            ;;
        --before)  BEFORE="$2"; shift 2 ;;
        --after)   AFTER="$2";  shift 2 ;;
        --diff)    DIFF=1; shift ;;
        --keep-running) KEEP_RUNNING=1; shift ;;
        *) echo "Unknown flag: $1" >&2; exit 1 ;;
    esac
done

# ---------------------------------------------------------------------------
# Build the demo (reuse the same pattern as demo_screenshot.sh)
# ---------------------------------------------------------------------------
demo_dir="$PROJECT_DIR/demos/$DEMO"
if [[ ! -d "$demo_dir" ]]; then
    echo "Go demo directory not found: $demo_dir" >&2
    exit 1
fi

# Kill any stale instance of this demo.
TITLE=$(extract_demo_title "$DEMO")
if [[ -n "$TITLE" ]]; then
    wmctrl -c "$TITLE" 2>/dev/null || true
    sleep 0.3
fi

binary="$(mktemp /tmp/takigo_interact_XXXXXX)"
trap '[[ "$KEEP_RUNNING" == "0" ]] && { kill "$DEMO_PID" 2>/dev/null || true; wait "$DEMO_PID" 2>/dev/null || true; }; rm -f "$binary"' EXIT

echo "Building demos/$DEMO..." >&2
if ! (cd "$PROJECT_DIR" && go build -o "$binary" "./demos/$DEMO/" 2>&1); then
    rm -f "$binary"
    echo "Build failed for demos/$DEMO" >&2
    exit 1
fi
chmod +x "$binary"

DEMO_GEOMETRY=$(extract_demo_geometry "$DEMO")
export DEMO_GEOMETRY
if command -v xrdb >/dev/null 2>&1; then
    XFT_DPI_VAL=$(xrdb -query 2>/dev/null | awk -F':[[:space:]]*' '/^Xft\.dpi/ {print $2; exit}')
    [[ -n "$XFT_DPI_VAL" ]] && export XFT_DPI="$XFT_DPI_VAL"
fi

echo "Running demos/$DEMO (title: \"$TITLE\")..." >&2
nohup "$binary" > /tmp/takigo_interact_out.txt 2>&1 &
DEMO_PID=$!

# Wait for the window.
TIMEOUT_SECS="${TIMEOUT_SECS:-15}"
tries=$(( TIMEOUT_SECS * 2 ))
wid=""
for _ in $(seq 1 "$tries"); do
    if [[ -n "$TITLE" ]]; then
        wid=$(wmctrl -l 2>/dev/null | grep -F "$TITLE" | awk '{print $1}' | head -1 || true)
    fi
    [[ -n "$wid" ]] && break
    sleep 0.5
done
if [[ -z "$wid" ]]; then
    echo "Timeout waiting for window titled: $TITLE" >&2
    exit 1
fi

# Focus + capture the BEFORE screenshot.
echo "Capturing before..." >&2
xdotool windowraise "$wid" 2>/dev/null || true
xdotool windowfocus --sync "$wid" 2>/dev/null || true
sleep "${SETTLE_SECS:-1.5}"
import -window "$wid" "$BEFORE"

# ---------------------------------------------------------------------------
# Run the event sequence
# ---------------------------------------------------------------------------
echo "Running ${#EVENTS[@]} event(s)..." >&2
DEFAULT_WAIT_MS="${DEFAULT_WAIT_MS:-200}"
for entry in "${EVENTS[@]}"; do
    IFS='|' read -r kind arg <<<"$entry"
    case "$kind" in
        wait)
            ms="$arg"
            [[ -z "$ms" ]] && ms="$DEFAULT_WAIT_MS"
            sleep "$(awk -v ms="$ms" 'BEGIN { printf "%.3f", ms/1000 }')"
            ;;
        key)
            xdotool key --window "$wid" --clearmodifiers "$arg"
            ;;
        type)
            xdotool type --window "$wid" --delay 20 "$arg"
            ;;
        click)
            x="${arg%,*}"
            y="${arg#*,}"
            xdotool mousemove --window "$wid" "$x" "$y"
            xdotool click --window "$wid" 1
            ;;
    esac
done

# Default settle delay between last event and the AFTER capture.
sleep "${AFTER_SETTLE_SECS:-1.0}"

# ---------------------------------------------------------------------------
# Capture AFTER + optional diff
# ---------------------------------------------------------------------------
echo "Capturing after..." >&2
import -window "$wid" "$AFTER"

if [[ "$DIFF" == "1" ]]; then
    diff_img="${AFTER%.png}_diff.png"
    SCORE=$(compare -metric MAE -format "%[distortion]" "$BEFORE" "$AFTER" "$diff_img" 2>/dev/null) && true
    RC=$?
    if [[ $RC -ne 0 && $RC -ne 1 ]]; then
        echo "compare failed (exit $RC)" >&2
        exit 1
    fi
    [[ -z "$SCORE" ]] && SCORE="unknown"
    echo ""
    echo "Demo:       $DEMO"
    echo "Events:     ${#EVENTS[@]}"
    echo "Before:     $BEFORE"
    echo "After:      $AFTER"
    echo "Diff score: $SCORE  (normalized MAE 0-1, 0 = no change)"
    echo "Diff image: $diff_img"
else
    echo ""
    echo "Demo:   $DEMO"
    echo "Events: ${#EVENTS[@]}"
    echo "Before: $BEFORE"
    echo "After:  $AFTER"
fi

if [[ "$KEEP_RUNNING" == "1" ]]; then
    echo "Leaving demo running (window $wid). Clean up manually." >&2
    DEMO_PID=""   # tell trap not to kill
fi
