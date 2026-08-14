#!/usr/bin/env bash
# scripts/_lib.sh -- shared helpers for the demo-comparison scripts.
#
# Source this from any script that needs:
#   - PROJECT_DIR, SCRIPT_DIR, SS_DIR, LOGS_DIR path variables
#   - tcl_demo_for <go_demo>           — map a Go demo name to its Tcl counterpart
#   - run_compare <go_demo> [tcl_demo] — run demo_compare.sh and print just the score
#   - set_skip_if_exists <0|1>         — export SKIP_IF_EXISTS for screenshot reuse
#
# Not executable on its own; sourced by other scripts.

# Resolve the absolute paths from this file's location. PROJECT_DIR is the
# takigo repo root; SCRIPT_DIR is the directory containing this _lib.sh.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SS_DIR="$PROJECT_DIR/tmp/screenshots"
LOGS_DIR="$PROJECT_DIR/tmp/logs"
mkdir -p "$SS_DIR" "$LOGS_DIR"

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
    (cd "$PROJECT_DIR" && go run ./cmd/demotitle "$src") 2>/dev/null || true
}

# extract_demo_geometry GO_DEMO
# Reads the Geometry() option from the Go source so the Tcl wrapper can be
# positioned identically. Falls back to "+300+300" when the demo has no
# Geometry() call (matching the wrapper's default).
extract_demo_geometry() {
    local demo="$1"
    local src="$PROJECT_DIR/demos/$demo/main.go"
    local geom
    geom=$(cd "$PROJECT_DIR" && go run ./cmd/demotitle -geometry "$src") 2>/dev/null || geom=""
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
    # Score line format: "Diff score: <number>  (normalized MAE ...)"
    echo "$out" | awk '/^Diff score:/ {print $3}'
}
