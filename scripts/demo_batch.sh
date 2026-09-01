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
# Animated demos (DEMO_ANIMATED in demo_map.sh) are skipped because their
# screenshots are nondeterministic. A score of 9999 marks a failed comparison.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/_lib.sh"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

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

SORTED_GO=$(echo "${!DEMO_MAP[@]}" | tr ' ' '\n' | sort)

# First pass: count demos that will actually be processed (after filter and
# animated skip), so the progress prefix is accurate.
TOTAL=0
for go_name in $SORTED_GO; do
    TCL_DEMO="${DEMO_MAP[$go_name]}"
    [[ "$TCL_DEMO" == "-" ]] && continue
    [[ -n "$FILTER" && "$go_name" != ${FILTER}* ]] && continue
    [[ " $DEMO_ANIMATED " == *" $go_name "* ]] && continue
    TOTAL=$(( TOTAL + 1 ))
done

COUNT=0
for go_name in $SORTED_GO; do
    TCL_DEMO="${DEMO_MAP[$go_name]}"
    [[ "$TCL_DEMO" == "-" ]] && continue

    # Apply filter
    if [[ -n "$FILTER" && "$go_name" != ${FILTER}* ]]; then
        continue
    fi

    # Skip animated demos (nondeterministic screenshots).
    if [[ " $DEMO_ANIMATED " == *" $go_name "* ]]; then
        echo "  SKIP: $go_name (animated)"
        continue
    fi

    COUNT=$(( COUNT + 1 ))
    echo ""
    echo "[$COUNT/$TOTAL] $go_name ..."

    if [[ ! -d "$PROJECT_DIR/demos/$go_name" ]]; then
        echo "  SKIP: no Go demo directory"
        continue
    fi

    # Run screenshot + compare, capture score. Stderr goes to a per-demo log so
    # failures are diagnosable instead of silently discarded.
    mkdir -p "$PROJECT_DIR/tmp/logs"
    ERR_LOG="$PROJECT_DIR/tmp/logs/${go_name}.compare.err"
    RESULT=$(bash "$SCRIPT_DIR/demo_compare.sh" "$go_name" "$TCL_DEMO" 2>"$ERR_LOG") || {
        echo "  FAILED: $go_name (see $ERR_LOG)" >&2
        echo "9999 $go_name $TCL_DEMO" >> "$SCORES_FILE"
        continue
    }
    rm -f "$ERR_LOG"   # clean up on success

    SCORE=$(echo "$RESULT" | grep "^Diff score:" | awk '{print $3}')
    [[ -z "$SCORE" ]] && SCORE="9999"

    echo "  Score: $SCORE (odiff diff %)"
    echo "$SCORE $go_name $TCL_DEMO" >> "$SCORES_FILE"
done

echo ""
echo "============================================================"
echo "Results sorted by diff % (highest = most different, 9999 = failed):"
echo "============================================================"
sort -rn "$SCORES_FILE" | tee "$SS_DIR/scores_sorted.txt"
echo ""
echo "Full scores: $SS_DIR/scores_sorted.txt"
echo "Screenshots: $SS_DIR/"
