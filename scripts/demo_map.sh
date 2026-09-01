#!/usr/bin/env bash
# demo_map.sh -- Mapping of Go demo names to Tcl demo names
#
# The mapping is derived from the filesystem (demos/*/main.go vs
# tk/library/demos/*.tcl), with explicit overrides for names that differ or
# demos that are deliberately not comparable. Adding a new demo no longer
# requires editing this file.
#
# Source this file or run it directly.
#
# Usage (direct):
#   demo_map.sh                    -- print all comparable demos (go_name tcl_name)
#   demo_map.sh <go_demo_name>     -- print the Tcl name for a Go demo
#
# Each printed line: GO_NAME TCL_NAME
# "-" means the Go demo has no Tcl counterpart (Go-only).

declare -A DEMO_MAP

# --- Explicit overrides ------------------------------------------------------
# Different names
DEMO_MAP[windowicons]=icon
# Go-only despite a .tcl existing (the system-tray icon isn't captured by a
# plain window screenshot, so there is nothing useful to compare).
DEMO_MAP[systray]=-

# --- Animated demos (nondeterministic screenshots; skipped in batch) --------
DEMO_ANIMATED="anilabel aniwave pendulum knightstour twind"

# --- Auto-derive the rest from the filesystem --------------------------------
DEMOS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/demos"
TCL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/tk/library/demos"

for go_dir in "$DEMOS_DIR"/*/; do
    go_name=$(basename "$go_dir")
    [[ -f "$go_dir/main.go" ]] || continue
    [[ -v DEMO_MAP[$go_name] ]] && continue   # explicit override wins
    if [[ -f "$TCL_DIR/${go_name}.tcl" ]]; then
        DEMO_MAP[$go_name]=$go_name
    else
        DEMO_MAP[$go_name]=-
    fi
done

# ---------------------------------------------------------------------------
# CLI interface
# ---------------------------------------------------------------------------
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ $# -eq 0 ]]; then
        # Print all comparable demos (go_name tcl_name)
        for go_name in $(echo "${!DEMO_MAP[@]}" | tr ' ' '\n' | sort); do
            tcl_name="${DEMO_MAP[$go_name]}"
            [[ "$tcl_name" == "-" ]] && continue
            echo "$go_name $tcl_name"
        done
    else
        go_name="$1"
        if [[ -v DEMO_MAP[$go_name] ]]; then
            echo "${DEMO_MAP[$go_name]}"
        else
            echo "-"
            exit 1
        fi
    fi
fi
