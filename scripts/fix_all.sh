#!/usr/bin/env bash
# fix_all.sh -- Fix all comparable demos one by one, tracking progress
#
# Supports interruption and resuming: progress is saved after each demo.
#
# Usage:
#   fix_all.sh [options]
#
# Options:
#   --status           Print current progress and exit
#   --reset <demo>     Reset a demo to pending (re-run it next time)
#   --skip  <demo>     Mark a demo as skipped (won't be processed)
#   --only  <demo>     Process only this demo (useful for testing)
#   --filter <prefix>  Process only demos whose name starts with prefix
#   --iterations N     Fix iterations per demo (default 1)
#   --no-retake        Reuse existing screenshots between demos
#
# Progress file: tmp/fix_all_progress.tsv
#   Columns: demo  tcl_demo  status  score_before  score_after  date
#   Status values: pending | done | failed | skipped

set -euo pipefail

# Claude cannot be invoked from inside an active Claude Code session.
if [[ -n "${CLAUDECODE:-}" ]]; then
    echo "ERROR: fix_all.sh must be run from a regular terminal, not from inside Claude Code." >&2
    echo "Open a new terminal in $(pwd) and run: bash scripts/fix_all.sh $*" >&2
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/_lib.sh"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PROGRESS_FILE="$PROJECT_DIR/tmp/fix_all_progress.tsv"

ITERATIONS=1
RETAKE_FLAG=""
ONLY_DEMO=""
FILTER=""

# Parse args
while [[ $# -gt 0 ]]; do
    case "$1" in
        --status)
            if [[ -f "$PROGRESS_FILE" ]]; then
                echo "Progress: $PROGRESS_FILE"
                echo ""
                printf "%-20s %-20s %-10s %-12s %-12s %s\n" \
                    "GO_DEMO" "TCL_DEMO" "STATUS" "SCORE_BEFORE" "SCORE_AFTER" "DATE"
                printf "%s\n" "$(printf '─%.0s' {1..80})"
                column -t -s $'\t' "$PROGRESS_FILE"
                echo ""
                echo "Summary:"
                awk -F'\t' '{counts[$3]++} END {for (s in counts) print "  " s ": " counts[s]}' \
                    "$PROGRESS_FILE" | sort
            else
                echo "No progress file yet: $PROGRESS_FILE"
            fi
            exit 0
            ;;
        --reset)
            shift; demo="$1"
            if [[ -f "$PROGRESS_FILE" ]]; then
                awk -F'\t' -v d="$demo" 'BEGIN{OFS="\t"} $1==d{$3="pending";$4="-";$5="-";$6="-"} {print}' \
                    "$PROGRESS_FILE" > "$PROGRESS_FILE.tmp" && mv "$PROGRESS_FILE.tmp" "$PROGRESS_FILE"
                echo "Reset '$demo' to pending."
            fi
            exit 0
            ;;
        --skip)
            shift; demo="$1"
            if [[ -f "$PROGRESS_FILE" ]]; then
                awk -F'\t' -v d="$demo" 'BEGIN{OFS="\t"} $1==d{$3="skipped"} {print}' \
                    "$PROGRESS_FILE" > "$PROGRESS_FILE.tmp" && mv "$PROGRESS_FILE.tmp" "$PROGRESS_FILE"
                echo "Marked '$demo' as skipped."
            fi
            exit 0
            ;;
        --only)     shift; ONLY_DEMO="$1" ;;
        --filter)   shift; FILTER="$1" ;;
        --iterations) shift; ITERATIONS="$1" ;;
        --no-retake) RETAKE_FLAG="--no-retake" ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
    shift
done

# ---------------------------------------------------------------------------
# Initialize progress file if it doesn't exist or is empty
# ---------------------------------------------------------------------------
if [[ ! -s "$PROGRESS_FILE" ]]; then
    echo "Initializing progress file from demo map..."
    # _lib.sh already sourced demo_map.sh.
    > "$PROGRESS_FILE"
    for go_name in $(echo "${!DEMO_MAP[@]}" | tr ' ' '\n' | sort); do
        tcl_name="${DEMO_MAP[$go_name]}"
        [[ "$tcl_name" == "-" ]] && continue
        # Skip if Go demo directory doesn't exist
        [[ ! -d "$PROJECT_DIR/demos/$go_name" ]] && continue
        printf "%s\t%s\t%s\t%s\t%s\t%s\n" \
            "$go_name" "$tcl_name" "pending" "-" "-" "-" >> "$PROGRESS_FILE"
    done
    echo "Initialized $(wc -l < "$PROGRESS_FILE") demos."
    echo ""
fi

# ---------------------------------------------------------------------------
# Auto-recover demos stuck in `in_progress` (left by a previous Ctrl+C)
# ---------------------------------------------------------------------------
if grep -qP '\tin_progress\t' "$PROGRESS_FILE" 2>/dev/null; then
    echo "Recovering in-progress demos to pending..."
    sed -i 's/\tin_progress\t/\tpending\t/' "$PROGRESS_FILE"
fi

# ---------------------------------------------------------------------------
# Helper: update one row in the progress file
# ---------------------------------------------------------------------------
update_progress() {
    local demo="$1" tcl="$2" status="$3" score_before="$4" score_after="$5"
    local date
    date=$(date +%Y-%m-%d)
    awk -F'\t' -v d="$demo" -v t="$tcl" -v s="$status" \
        -v sb="$score_before" -v sa="$score_after" -v dt="$date" \
        'BEGIN{OFS="\t"} $1==d{$2=t;$3=s;$4=sb;$5=sa;$6=dt} {print}' \
        "$PROGRESS_FILE" > "$PROGRESS_FILE.tmp" \
        && mv "$PROGRESS_FILE.tmp" "$PROGRESS_FILE"
}

# ---------------------------------------------------------------------------
# Process demos
# ---------------------------------------------------------------------------
TOTAL=$(grep -cP '\tpending\t' "$PROGRESS_FILE" 2>/dev/null || echo 0)
echo "Pending demos: $TOTAL"
echo ""

COUNT=0
while IFS=$'\t' read -r go_demo tcl_demo status score_before score_after _date; do
    # Filter
    [[ "$status" != "pending" ]] && continue
    [[ -n "$ONLY_DEMO"  && "$go_demo" != "$ONLY_DEMO"   ]] && continue
    [[ -n "$FILTER"     && "$go_demo" != ${FILTER}*      ]] && continue

    COUNT=$(( COUNT + 1 ))
    echo "┌──────────────────────────────────────────────────"
    echo "│ Demo $COUNT: $go_demo  (Tcl: $tcl_demo)"
    echo "└──────────────────────────────────────────────────"

    # Mark in-progress immediately so Ctrl+C leaves a trace
    update_progress "$go_demo" "$tcl_demo" "in_progress" "-" "-"

    # Run fix_demo.sh and capture output + exit code
    LOG="$LOGS_DIR/${go_demo}_fix_all.log"

    SCORE_B="-"
    SCORE_A="-"
    EXIT_CODE=0

    # Capture score lines from fix_demo.sh output
    set +e
    FIX_OUTPUT=$(bash "$SCRIPT_DIR/fix_demo.sh" "$go_demo" \
        $RETAKE_FLAG --iterations "$ITERATIONS" 2>&1 | tee "$LOG")
    EXIT_CODE=$?
    set -e

    SCORE_B=$(echo "$FIX_OUTPUT" | grep "^Score before:" | tail -1 | awk '{print $3}')
    SCORE_A=$(echo "$FIX_OUTPUT" | grep "^Score after:"  | tail -1 | awk '{print $3}')
    [[ -z "$SCORE_B" ]] && SCORE_B="-"
    [[ -z "$SCORE_A" ]] && SCORE_A="-"

    if [[ "$EXIT_CODE" -eq 0 ]]; then
        update_progress "$go_demo" "$tcl_demo" "done" "$SCORE_B" "$SCORE_A"
        echo "✓ Done: $go_demo  before=$SCORE_B after=$SCORE_A"
    else
        update_progress "$go_demo" "$tcl_demo" "failed" "$SCORE_B" "$SCORE_A"
        echo "✗ Failed: $go_demo  (exit $EXIT_CODE) — see $LOG"
    fi

    echo ""
done < "$PROGRESS_FILE"

echo "═══════════════════════════════════════════════════"
echo "Batch complete. Progress:"
awk -F'\t' '{counts[$3]++} END {for (s in counts) print "  " s ": " counts[s]}' \
    "$PROGRESS_FILE" | sort
echo ""
echo "Run with --status to see per-demo scores."
echo "═══════════════════════════════════════════════════"
