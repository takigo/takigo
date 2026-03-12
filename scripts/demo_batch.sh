#!/usr/bin/env bash
# demo_batch.sh -- Screenshot and score all comparable demos
#
# Usage:
#   demo_batch.sh [--retake] [filter_prefix]
#
# Options:
#   --retake        : retake all screenshots even if files exist
#   filter_prefix   : only process demos starting with this string
#
# Output:
#   tmp/screenshots/<demo>_*.png  for each demo
#   tmp/screenshots/scores.txt    sorted by diff score (worst first)
#
# This is useful for identifying which demos need the most work.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SS_DIR="$PROJECT_DIR/tmp/screenshots"

RETAKE=0
FILTER=""

for arg in "$@"; do
    case "$arg" in
        --retake) RETAKE=1 ;;
        *)        FILTER="$arg" ;;
    esac
done

if [[ "$RETAKE" == "1" ]]; then
    export SKIP_IF_EXISTS=0
else
    export SKIP_IF_EXISTS=1
fi

mkdir -p "$SS_DIR"

SCORES_FILE="$SS_DIR/scores.txt"
> "$SCORES_FILE"

# Get all comparable demos
mapfile -t DEMOS < <(bash "$SCRIPT_DIR/demo_map.sh")

TOTAL=${#DEMOS[@]}
COUNT=0

for entry in "${DEMOS[@]}"; do
    GO_DEMO=$(echo "$entry" | awk '{print $1}')
    TCL_DEMO=$(echo "$entry" | awk '{print $2}')

    # Apply filter
    if [[ -n "$FILTER" && "$GO_DEMO" != ${FILTER}* ]]; then
        continue
    fi

    COUNT=$(( COUNT + 1 ))
    echo ""
    echo "[$COUNT/$TOTAL] $GO_DEMO ..."

    GO_IMG="$SS_DIR/${GO_DEMO}_go.png"
    TCL_IMG="$SS_DIR/${TCL_DEMO}_tcl.png"

    # Skip demos where Go directory doesn't exist
    if [[ ! -d "$PROJECT_DIR/demos/$GO_DEMO" ]]; then
        echo "  SKIP: no Go demo directory"
        continue
    fi

    # Run screenshot + compare, capture score
    RESULT=$(bash "$SCRIPT_DIR/demo_compare.sh" "$GO_DEMO" "$TCL_DEMO" 2>/dev/null) || {
        echo "  FAILED: $GO_DEMO" >&2
        echo "FAILED $GO_DEMO -" >> "$SCORES_FILE"
        continue
    }

    SCORE=$(echo "$RESULT" | grep "^Diff score:" | awk '{print $3}')
    [[ -z "$SCORE" ]] && SCORE="unknown"

    echo "  Score: $SCORE (MAE)"
    echo "$SCORE $GO_DEMO $TCL_DEMO" >> "$SCORES_FILE"
done

echo ""
echo "============================================================"
echo "Results sorted by diff score (highest = most different):"
echo "============================================================"
sort -rn "$SCORES_FILE" | tee "$SS_DIR/scores_sorted.txt"
echo ""
echo "Full scores: $SS_DIR/scores_sorted.txt"
echo "Screenshots: $SS_DIR/"
