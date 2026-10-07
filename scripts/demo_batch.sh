#!/usr/bin/env bash
# demo_batch.sh -- Screenshot and score all comparable demos
#
# Usage:
#   demo_batch.sh [--retake] [--stability N] [--update-baseline] [filter_prefix]
#
# Options:
#   --retake          : retake all screenshots even if files exist
#   --stability N     : capture each side N times in total and report the max
#                       self-diff % per side (anything > 0 is a flaky capture)
#   --update-baseline : write the results into demos/parity.tsv (rows for
#                       demos not processed in this run are kept; notes are kept)
#   filter_prefix     : only process demos starting with this string
#
# Runs headless (xvfb-run) unless HEADLESS=0, so results don't depend on the
# desktop's DPI, fonts or window manager. Timers are frozen on both sides
# (TAKIGO_FREEZE_TIMERS=1), so animated demos are compared on their first frame.
#
# Output:
#   tmp/screenshots/<demo>_*.png      for each demo
#   tmp/screenshots/scores.tsv        demo, tcl, score, tree diffs, sizes, stability
#   tmp/screenshots/<demo>_tree.txt   structural diff per demo (internal/cmd/demodiff)
#   tmp/screenshots/scores_sorted.txt "score go tcl", worst first
# A score of 9999 marks a failed comparison.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/_lib.sh"

export HEADLESS="${HEADLESS:-1}"
maybe_xvfb "$@"

RETAKE=0
FILTER=""
STABILITY=1
UPDATE_BASELINE=0

while [[ $# -gt 0 ]]; do
    case "$1" in
        --retake)          RETAKE=1 ;;
        --stability)       shift; STABILITY="$1" ;;
        --update-baseline) UPDATE_BASELINE=1 ;;
        -*)                echo "Unknown option: $1" >&2; exit 1 ;;
        *)                 FILTER="$1" ;;
    esac
    shift
done

set_skip_if_exists $(( 1 - RETAKE ))
export TAKIGO_FREEZE_TIMERS="${TAKIGO_FREEZE_TIMERS:-1}"
# The comparison is against Tk, so the Go side draws exactly as Tk does.
export TAKIGO_CLASSIC="${TAKIGO_CLASSIC:-1}"

SCORES_TSV="$SS_DIR/scores.tsv"
SORTED_TXT="$SS_DIR/scores_sorted.txt"
BASELINE="$PROJECT_DIR/demos/parity.tsv"
printf 'demo\ttcl\tpixel_pct\ttree_diffs\tgo_size\ttcl_size\tselfdiff_go\tselfdiff_tcl\n' > "$SCORES_TSV"

# selfdiff SIDE DEMO TCL PRIMARY
# Recaptures one side STABILITY-1 times and prints the max odiff % against
# PRIMARY ("size" if a recapture has different dimensions, "err" on failure).
selfdiff() {
    local side="$1" demo="$2" tcl="$3" primary="$4"
    local max=0 i tmp s
    for (( i = 2; i <= STABILITY; i++ )); do
        tmp="$SS_DIR/${demo}_${side}.stab${i}.png"
        if ! bash "$SCRIPT_DIR/demo_screenshot.sh" "$demo" "$side" "$tmp" "$tcl" 2>/dev/null; then
            echo "err"; return
        fi
        if [[ "$(magick identify -format '%wx%h' "$tmp")" != "$(magick identify -format '%wx%h' "$primary")" ]]; then
            rm -f "$tmp"; echo "size"; return
        fi
        s=$(odiff_score "$primary" "$tmp" 2>/dev/null) || s="err"
        rm -f "$tmp"
        [[ "$s" == "err" ]] && { echo "err"; return; }
        max=$(awk -v a="$max" -v b="$s" 'BEGIN { print (b > a) ? b : a }')
    done
    echo "$max"
}

DEMOS=()
for go_name in $(printf '%s\n' "${!DEMO_MAP[@]}" | sort); do
    [[ "${DEMO_MAP[$go_name]}" == "-" ]] && continue
    [[ -n "$FILTER" && "$go_name" != ${FILTER}* ]] && continue
    if [[ "$TAKIGO_FREEZE_TIMERS" != "1" && " $DEMO_ANIMATED " == *" $go_name "* ]]; then
        echo "  SKIP: $go_name (animated; timers not frozen)"
        continue
    fi
    [[ -d "$PROJECT_DIR/demos/$go_name" ]] || { echo "  SKIP: $go_name (no Go demo directory)"; continue; }
    DEMOS+=("$go_name")
done

# Build every demo up front so the whole run measures one snapshot of the
# source, even if files are edited while it runs.
export DEMO_BIN_DIR="$BIN_DIR/demos"
rm -rf "$DEMO_BIN_DIR"
mkdir -p "$DEMO_BIN_DIR"
echo "Building ${#DEMOS[@]} demos into $DEMO_BIN_DIR ..."
(cd "$PROJECT_DIR" && go build -o "$DEMO_BIN_DIR/" $(printf './demos/%s ' "${DEMOS[@]}"))

TOTAL=${#DEMOS[@]}
COUNT=0
for go_name in "${DEMOS[@]}"; do
    TCL_DEMO="${DEMO_MAP[$go_name]}"
    COUNT=$(( COUNT + 1 ))
    echo ""
    echo "[$COUNT/$TOTAL] $go_name ..."

    ERR_LOG="$LOGS_DIR/${go_name}.compare.err"
    SCORE=9999 TREE=- GO_SIZE=- TCL_SIZE=- SD_GO=- SD_TCL=-
    # One retry: a window occasionally fails to appear in time under load.
    if RESULT=$(bash "$SCRIPT_DIR/demo_compare.sh" "$go_name" "$TCL_DEMO" 2>"$ERR_LOG") ||
        RESULT=$(bash "$SCRIPT_DIR/demo_compare.sh" "$go_name" "$TCL_DEMO" 2>"$ERR_LOG"); then
        rm -f "$ERR_LOG"
        SCORE=$(awk '/^Diff score:/ {print $3}' <<<"$RESULT")
        [[ -z "$SCORE" ]] && SCORE=9999
        TREE=$(awk '/^Tree diffs:/ {print $3}' <<<"$RESULT")
        GO_SIZE=$(sed -n 's/^Go image:.*(\([0-9]*x[0-9]*\)).*/\1/p' <<<"$RESULT")
        TCL_SIZE=$(sed -n 's/^Tcl image:.*(\([0-9]*x[0-9]*\)).*/\1/p' <<<"$RESULT")
        if (( STABILITY > 1 )); then
            SD_GO=$(selfdiff go "$go_name" "$TCL_DEMO" "$SS_DIR/${go_name}_go.png")
            SD_TCL=$(selfdiff tcl "$go_name" "$TCL_DEMO" "$SS_DIR/${TCL_DEMO}_tcl.png")
        fi
    else
        echo "  FAILED: $go_name (see $ERR_LOG)" >&2
    fi

    STAB_MSG=""
    (( STABILITY > 1 )) && STAB_MSG="  selfdiff go=$SD_GO tcl=$SD_TCL"
    echo "  Score: $SCORE  tree=${TREE:--}  go=$GO_SIZE tcl=$TCL_SIZE$STAB_MSG"
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$go_name" "$TCL_DEMO" "$SCORE" "${TREE:--}" \
        "${GO_SIZE:--}" "${TCL_SIZE:--}" "$SD_GO" "$SD_TCL" >> "$SCORES_TSV"
done

tail -n +2 "$SCORES_TSV" | awk -F'\t' '{print $3, $1, $2, "tree=" $4}' | sort -rn > "$SORTED_TXT"

echo ""
echo "============================================================"
echo "Results sorted by diff % (highest = most different, 9999 = failed):"
echo "============================================================"
cat "$SORTED_TXT"

if (( STABILITY > 1 )); then
    FLAKY=$(tail -n +2 "$SCORES_TSV" | awk -F'\t' '$7 != "0" || $8 != "0" {print "  " $1 " go=" $7 " tcl=" $8}')
    echo ""
    if [[ -n "$FLAKY" ]]; then
        echo "Unstable captures (self-diff > 0 across $STABILITY runs):"
        echo "$FLAKY"
    else
        echo "All captures stable across $STABILITY runs."
    fi
fi

if [[ "$UPDATE_BASELINE" == "1" ]]; then
    echo ""
    update_baseline "$SCORES_TSV"
fi

echo ""
echo "Full scores: $SCORES_TSV"
echo "Screenshots: $SS_DIR/"
