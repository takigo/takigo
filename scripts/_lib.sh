#!/usr/bin/env bash
# scripts/_lib.sh -- shared helpers for the demo-comparison scripts.
#
# Source this from any script that needs:
#   - PROJECT_DIR, SCRIPT_DIR, SS_DIR, LOGS_DIR, BIN_DIR path variables
#   - tcl_demo_for <go_demo>           — map a Go demo name to its Tcl counterpart
#   - run_compare <go_demo> [tcl_demo] — run demo_compare.sh and print just the score
#   - set_skip_if_exists <0|1>         — export SKIP_IF_EXISTS for screenshot reuse
#   - odiff_score <base> <cmp> [diff]  — run odiff, print diff % (0-100)
#   - find_windows_exact <title>       — list window IDs by exact WM_NAME
#   - maybe_xvfb "$@"                  — re-exec under xvfb-run when headless
#   - demotitle_bin                    — cached path to the built demotitle CLI
#
# Not executable on its own; sourced by other scripts.

# Resolve the absolute paths from this file's location. PROJECT_DIR is the
# takigo repo root; SCRIPT_DIR is the directory containing this _lib.sh.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SS_DIR="$PROJECT_DIR/tmp/screenshots"
LOGS_DIR="$PROJECT_DIR/tmp/logs"
BIN_DIR="$PROJECT_DIR/tmp/bin"
mkdir -p "$SS_DIR" "$LOGS_DIR" "$BIN_DIR"

# maybe_xvfb [ARGS...]
# Re-exec the calling script under xvfb-run when there is no usable display.
# Call it near the top of a script, before any DISPLAY-dependent work, passing
# the script's own arguments:  maybe_xvfb "$@"
# Force headless with HEADLESS=1; otherwise auto-detect an unset/unconnectable
# DISPLAY. The _XVFB_RUNNING guard prevents infinite re-exec. Xvfb runs at a
# fixed 1280x1024x24 / 96dpi so captures are deterministic.
maybe_xvfb() {
    [[ -n "${_XVFB_RUNNING:-}" ]] && return 0
    local need=0
    [[ "${HEADLESS:-0}" == "1" ]] && need=1
    if [[ -z "${DISPLAY:-}" ]] || ! xdpyinfo >/dev/null 2>&1; then
        need=1
    fi
    if [[ "$need" == "1" ]]; then
        command -v xvfb-run >/dev/null 2>&1 || {
            echo "HEADLESS/display unavailable but xvfb-run not found (install xorg-server-xvfb)" >&2
            exit 1
        }
        export _XVFB_RUNNING=1
        exec xvfb-run -a -s "-screen 0 1280x1024x24 -dpi 96" "$0" "$@"
    fi
}

# demotitle_bin
# Builds cmd/demotitle once into tmp/bin and echoes the path. The cached binary
# is rebuilt only when its source is newer, so the per-demo `go run` cost is
# paid once per source change instead of once per invocation.
demotitle_bin() {
    local bin="$BIN_DIR/demotitle"
    local src="$PROJECT_DIR/cmd/demotitle/main.go"
    if [[ ! -x "$bin" || "$src" -nt "$bin" ]]; then
        (cd "$PROJECT_DIR" && go build -o "$bin" ./cmd/demotitle) || return 1
    fi
    echo "$bin"
}

# find_windows_exact TITLE
# Prints window IDs whose WM_NAME exactly equals TITLE (fixed string, not
# regex). Window titles may contain regex metacharacters (e.g. parentheses),
# so xdotool search --name (which treats the pattern as a regex) cannot be
# used directly for exact matching.
find_windows_exact() {
    local title="$1"
    local id
    xdotool search --name "" 2>/dev/null | while read -r id; do
        [[ -z "$id" ]] && continue
        [[ "$(xdotool getwindowname "$id" 2>/dev/null)" == "$title" ]] && echo "$id"
    done || true
}

# odiff_score BASE COMPARE [DIFF_OUT]
# Runs odiff with anti-aliasing ignored and prints the diff percentage
# (0-100, lower = more similar). Exit 0 on success, 1 on error.
# odiff exit codes: 0 = match, 21 = layout diff, 22 = pixel diff.
odiff_score() {
    local base="$1" cmp="$2" diff_out="${3:-}"
    local out rc
    out=$(odiff "$base" "$cmp" ${diff_out:+"$diff_out"} --parsable-stdout --aa 2>/dev/null)
    rc=$?
    case "$rc" in
        0)  echo "0"; return 0 ;;
        22) echo "${out##*;}"; return 0 ;;
        21) echo "layout-diff" >&2; return 1 ;;
        *)  echo "odiff failed (exit $rc)" >&2; return 1 ;;
    esac
}

# tcl_demo_for GO_DEMO
# Prints the Tcl counterpart name. Prints "-" and returns 1 if Go-only or
# unknown. Sources demo_map.sh on first call so the mapping is loaded.
if [[ -z "${_DEMO_MAP_LOADED:-}" ]]; then
    # shellcheck disable=SC1091
    source "$SCRIPT_DIR/demo_map.sh"
    _DEMO_MAP_LOADED=1
fi
tcl_demo_for() {
    local go_demo="$1"
    if [[ -z "${DEMO_MAP[$go_demo]:-}" || "${DEMO_MAP[$go_demo]}" == "-" ]]; then
        echo "-"
        return 1
    fi
    echo "${DEMO_MAP[$go_demo]}"
}

# extract_demo_title GO_DEMO
# Reads the Title() option from the Go source via cmd/demotitle.
# Returns the empty string if the demo has no Title() call or the file
# can't be parsed. Use go/parser + go/ast (not grep) so escaped strings
# and comments are handled correctly.
extract_demo_title() {
    local demo="$1"
    local src="$PROJECT_DIR/demos/$demo/main.go"
    local bin
    bin=$(demotitle_bin) 2>/dev/null || true
    [[ -z "$bin" ]] && return 0
    "$bin" "$src" 2>/dev/null || true
}

# extract_demo_geometry GO_DEMO
# Reads the Geometry() option from the Go source so the Tcl wrapper can be
# positioned identically. Falls back to "+300+300" when the demo has no
# Geometry() call (matching the wrapper's default).
extract_demo_geometry() {
    local demo="$1"
    local src="$PROJECT_DIR/demos/$demo/main.go"
    local geom
    local bin
    bin=$(demotitle_bin) 2>/dev/null || true
    [[ -z "$bin" ]] && { echo "+300+300"; return 0; }
    geom=$("$bin" -geometry "$src") 2>/dev/null || geom=""
    [[ -z "$geom" ]] && geom="+300+300"
    echo "$geom"
}

# set_skip_if_exists FLAG (0 or 1)
# 0 = always retake screenshots; 1 = reuse existing files when present.
set_skip_if_exists() {
    if [[ "$1" == "0" ]]; then
        export SKIP_IF_EXISTS=0
    else
        export SKIP_IF_EXISTS=1
    fi
}

# run_compare GO_DEMO [TCL_DEMO]
# Runs demo_compare.sh for the given Go demo (and its Tcl counterpart, or
# the second argument if it differs). Captures the score and prints it on
# stdout. Stderr goes to a per-demo log file in tmp/logs/ so failures
# are diagnosable instead of silently discarded.
#
# Exit codes:
#   0  — comparison succeeded; the score is on stdout.
#   1  — demo_compare.sh failed (see the per-demo log).
run_compare() {
    local go_demo="$1"
    local tcl_demo="${2:-}"
    if [[ -z "$tcl_demo" ]]; then
        tcl_demo=$(tcl_demo_for "$go_demo") || {
            echo "no Tcl counterpart for $go_demo" >&2
            return 1
        }
    fi
    local err_log="$LOGS_DIR/${go_demo}.compare.err"
    local out
    out=$(bash "$SCRIPT_DIR/demo_compare.sh" "$go_demo" "$tcl_demo" 2>"$err_log") || {
        echo "compare failed for $go_demo (see $err_log)" >&2
        return 1
    }
    rm -f "$err_log"
    # Score line format: "Diff score: <number>  (odiff diff % ...)"
    echo "$out" | awk '/^Diff score:/ {print $3}'
}
