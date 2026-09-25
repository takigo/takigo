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
#   tmp/screenshots/scores.tsv        demo, tcl, score, sizes, stability
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

SCORES_TSV="$SS_DIR/scores.tsv"
SORTED_TXT="$SS_DIR/scores_sorted.txt"
BASELINE="$PROJECT_DIR/demos/parity.tsv"
printf 'demo\ttcl\tpixel_pct\tgo_size\ttcl_size\tselfdiff_go\tselfdiff_tcl\n' > "$SCORES_TSV"

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

TOTAL=${#DEMOS[@]}
COUNT=0
for go_name in "${DEMOS[@]}"; do
    TCL_DEMO="${DEMO_MAP[$go_name]}"
    COUNT=$(( COUNT + 1 ))
    echo ""
    echo "[$COUNT/$TOTAL] $go_name ..."

    ERR_LOG="$LOGS_DIR/${go_name}.compare.err"
    SCORE=9999 GO_SIZE=- TCL_SIZE=- SD_GO=- SD_TCL=-
    if RESULT=$(bash "$SCRIPT_DIR/demo_compare.sh" "$go_name" "$TCL_DEMO" 2>"$ERR_LOG"); then
        rm -f "$ERR_LOG"
        SCORE=$(awk '/^Diff score:/ {print $3}' <<<"$RESULT")
        [[ -z "$SCORE" ]] && SCORE=9999
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
    echo "  Score: $SCORE  go=$GO_SIZE tcl=$TCL_SIZE$STAB_MSG"
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$go_name" "$TCL_DEMO" "$SCORE" \
        "${GO_SIZE:--}" "${TCL_SIZE:--}" "$SD_GO" "$SD_TCL" >> "$SCORES_TSV"
done

tail -n +2 "$SCORES_TSV" | awk -F'\t' '{print $3, $1, $2}' | sort -rn > "$SORTED_TXT"

echo ""
echo "============================================================"
echo "Results sorted by diff % (highest = most different, 9999 = failed):"
echo "============================================================"
cat "$SORTED_TXT"

if (( STABILITY > 1 )); then
    FLAKY=$(tail -n +2 "$SCORES_TSV" | awk -F'\t' '$6 != "0" || $7 != "0" {print "  " $1 " go=" $6 " tcl=" $7}')
    echo ""
    if [[ -n "$FLAKY" ]]; then
        echo "Unstable captures (self-diff > 0 across $STABILITY runs):"
        echo "$FLAKY"
    else
        echo "All captures stable across $STABILITY runs."
    fi
fi

if [[ "$UPDATE_BASELINE" == "1" ]]; then
    TMP_BASE="$BASELINE.tmp"
    awk -F'\t' -v OFS='\t' '
        FNR == 1 && FILENAME == ARGV[1] { next }
        FILENAME == ARGV[1] {
            st = ($3 == "9999") ? "failed" : ($3 == "0" ? "exact" : "close")
            new[$1] = $1 OFS $2 OFS $3 OFS $4 OFS $5 OFS st
            next
        }
        /^#/ || NF < 6 { next }
        { old[$1] = $0; note[$1] = (NF >= 7) ? $7 : "" }
        END {
            for (d in old) if (!(d in new)) rows[d] = old[d]
            for (d in new) rows[d] = new[d] OFS note[d]
            n = asorti(rows, keys)
            for (i = 1; i <= n; i++) print rows[keys[i]]
        }' "$SCORES_TSV" "$( [[ -f "$BASELINE" ]] && echo "$BASELINE" || echo /dev/null )" > "$TMP_BASE.rows"
    SUMMARY=$(awk -F'\t' '{c[$6]++} END {printf "exact %d / close %d / failed %d", c["exact"], c["close"], c["failed"]}' "$TMP_BASE.rows")
    {
        echo "# Demo parity baseline -- written by: bash scripts/demo_batch.sh --retake --update-baseline"
        echo "# Headless Xvfb 96dpi, pinned DejaVu fonts, frozen timers. pixel_pct = odiff % (lower is better)."
        echo "# summary: $SUMMARY"
        printf '# demo\ttcl\tpixel_pct\tgo_size\ttcl_size\tstatus\tnote\n'
        cat "$TMP_BASE.rows"
    } > "$TMP_BASE"
    rm -f "$TMP_BASE.rows"
    mv "$TMP_BASE" "$BASELINE"
    echo ""
    echo "Baseline updated: $BASELINE ($SUMMARY)"
fi

echo ""
echo "Full scores: $SCORES_TSV"
echo "Screenshots: $SS_DIR/"
