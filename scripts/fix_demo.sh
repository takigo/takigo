#!/usr/bin/env bash
# fix_demo.sh -- Compare one demo against Tcl original and auto-fix with an LLM
#
# Usage:
#   fix_demo.sh <demoname> [options]
#
# Options:
#   --no-retake    Reuse existing screenshots (faster when iterating)
#   --iterations N Run up to N fix+compare cycles (default 1)
#
# The LLM tool is configurable via the LLM_CMD env var (default: claude). The
# prompt is fed on stdin; the tool must write its response to stdout.
#
# Output:
#   tmp/screenshots/<demo>_*.png  updated after each fix
#   tmp/logs/<demo>.log           the LLM's output

set -euo pipefail

# shellcheck disable=SC1091
source "$(cd "$(dirname "$0")" && pwd)/_lib.sh"

llm_guard

DEMO="${1:-}"
if [[ -z "$DEMO" ]]; then
    echo "Usage: $0 <demoname> [--no-retake] [--iterations N]" >&2
    exit 1
fi

RETAKE=1
ITERATIONS=1
shift
while [[ $# -gt 0 ]]; do
    case "$1" in
        --no-retake)   RETAKE=0 ;;
        --iterations)  shift; ITERATIONS="$1" ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
    shift
done

if ! TCL_DEMO=$(tcl_demo_for "$DEMO"); then
    echo "Demo '$DEMO' has no Tcl counterpart — nothing to compare." >&2
    exit 1
fi

# ---------------------------------------------------------------------------
# run_compare -- take screenshots + compute diff score, print score
# ---------------------------------------------------------------------------
run_compare() {
    set_skip_if_exists "$1"
    local out
    out=$(bash "$SCRIPT_DIR/demo_compare.sh" "$DEMO" "$TCL_DEMO")
    echo "$out" >&2
    echo "$out" | awk '/^Diff score:/ {print $3}'
}

# ---------------------------------------------------------------------------
# run_llm -- call the configured LLM non-interactively to analyze and fix the demo
# ---------------------------------------------------------------------------
run_llm() {
    local log="$1"

    eval "$LLM_CMD" 2>&1 <<EOF | tee "$log"
Fix the Go demo '$DEMO' to visually match the Tcl/Tk original.

Screenshot files — Read ALL THREE before editing anything:
  Go version:        $SS_DIR/${DEMO}_go.png
  Tcl/Tk original:   $SS_DIR/${TCL_DEMO}_tcl.png
  Side-by-side diff: $SS_DIR/${DEMO}_side.png

Go source to fix: $PROJECT_DIR/demos/$DEMO/main.go

Instructions:
1. Read all three images. Carefully note every visual difference:
   widget sizes, fonts, padding/spacing, colors, label text, geometry,
   number and arrangement of widgets, border/relief styles.
2. Read the Go source file.
3. Edit the Go source to fix ALL identified visual differences.
   Use Bash to run: cd $PROJECT_DIR && go build ./demos/$DEMO/
   to confirm the fix compiles.
4. Only fix appearance — do not change behavior or add new features.

Focus areas in order of importance:
- Missing or extra widgets vs the Tcl version
- Wrong padding/margin (pack options: padx/pady, ipadx/ipady)
- Wrong font (size, family, weight)
- Wrong widget dimensions (width/height options)
- Wrong colors or relief styles
- Incorrect geometry string

After editing, print a brief summary of changes made.
EOF
}

# ---------------------------------------------------------------------------
# Main loop
# ---------------------------------------------------------------------------
echo "╔══════════════════════════════════════════════════╗"
echo "║  fix_demo: $DEMO (Tcl: $TCL_DEMO)"
echo "╚══════════════════════════════════════════════════╝"
echo ""

echo "── Initial comparison ──────────────────────────────"
SCORE=$(run_compare "$RETAKE")
echo "Score before: $SCORE"
echo ""

for i in $(seq 1 "$ITERATIONS"); do
    echo "── Iteration $i / $ITERATIONS ─────────────────────────"
    LOG="$LOGS_DIR/${DEMO}_iter${i}.log"
    echo "  Calling $LLM_NAME... (log: $LOG)"
    echo ""
    run_llm "$LOG"

    echo ""
    echo "── Re-comparing after fix ──────────────────────────"
    NEW_SCORE=$(run_compare 1)
    echo ""
    echo "Score before: $SCORE"
    echo "Score after:  $NEW_SCORE"

    SCORE="$NEW_SCORE"
    echo ""
done

echo "═══════════════════════════════════════════════════"
echo "Final score: $SCORE  (lower = more similar to Tcl)"
echo "Screenshots: $SS_DIR/${DEMO}_side.png"
echo "═══════════════════════════════════════════════════"
