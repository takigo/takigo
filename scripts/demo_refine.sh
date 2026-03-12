#!/usr/bin/env bash
# demo_refine.sh -- Full refinement workflow for one demo
#
# Usage:
#   demo_refine.sh <demoname> [--retake]
#
# 1. Looks up the Tcl counterpart name from demo_map.sh
# 2. Takes screenshots of both Go and Tcl versions
# 3. Generates side-by-side comparison and diff heatmap
# 4. Prints paths so Claude Code can Read the images and analyze them
#
# --retake : force new screenshots even if they already exist
#
# After running this script, use Claude Code's Read tool to view:
#   tmp/screenshots/<demo>_go.png
#   tmp/screenshots/<demo>_tcl.png
#   tmp/screenshots/<demo>_side.png    (side-by-side with diff)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

DEMO="${1:-}"
RETAKE=0
[[ "${2:-}" == "--retake" ]] && RETAKE=1

if [[ -z "$DEMO" ]]; then
    echo "Usage: $0 <demoname> [--retake]"
    echo ""
    echo "Available demos (with Tcl counterparts):"
    bash "$SCRIPT_DIR/demo_map.sh"
    exit 1
fi

# Look up Tcl name
source "$SCRIPT_DIR/demo_map.sh"
TCL_DEMO="${DEMO_MAP[$DEMO]:-$DEMO}"

if [[ "$TCL_DEMO" == "-" ]]; then
    echo "Demo '$DEMO' has no Tcl counterpart — Go-only demo, nothing to compare." >&2
    exit 1
fi

echo "Demo: $DEMO  (Tcl: $TCL_DEMO)"
echo "Project: $PROJECT_DIR"
echo ""

if [[ "$RETAKE" == "1" ]]; then
    export SKIP_IF_EXISTS=0
else
    export SKIP_IF_EXISTS=1
fi

bash "$SCRIPT_DIR/demo_compare.sh" "$DEMO" "$TCL_DEMO"

SS_DIR="$PROJECT_DIR/tmp/screenshots"
echo ""
echo "============================================================"
echo "Ready for Claude Code analysis. Read these image files:"
echo "  Go:          $SS_DIR/${DEMO}_go.png"
echo "  Tcl:         $SS_DIR/${TCL_DEMO}_tcl.png"
echo "  Side-by-side: $SS_DIR/${DEMO}_side.png"
echo "  Diff heatmap: $SS_DIR/${DEMO}_diff.png"
echo "============================================================"
echo ""
echo "Go source:   $PROJECT_DIR/demos/$DEMO/main.go"
