#!/usr/bin/env bash
# demo_compare.sh -- Compare Go vs Tcl versions of a demo visually
#
# Usage:
#   demo_compare.sh <demoname> [tcl_demo_name]
#
# Outputs (in tmp/screenshots/):
#   <demo>_go.png       -- Go screenshot
#   <demo>_tcl.png      -- Tcl screenshot
#   <demo>_diff.png     -- Pixel-level difference heatmap
#   <demo>_side.png     -- Side-by-side montage for easy viewing
#
# Prints a diff score (mean absolute error, lower = more similar) to stdout.
#
# Environment:
#   DISPLAY, SETTLE_SECS, TIMEOUT_SECS  (passed through to demo_screenshot.sh)
#   SKIP_IF_EXISTS=1                    (skip screenshot if file already exists)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SS_DIR="$PROJECT_DIR/tmp/screenshots"

DEMO="$1"
TCL_DEMO="${2:-$DEMO}"

GO_IMG="$SS_DIR/${DEMO}_go.png"
TCL_IMG="$SS_DIR/${TCL_DEMO}_tcl.png"
DIFF_IMG="$SS_DIR/${DEMO}_diff.png"
SIDE_IMG="$SS_DIR/${DEMO}_side.png"

mkdir -p "$SS_DIR"

# ---------------------------------------------------------------------------
# Take screenshots
# ---------------------------------------------------------------------------
take_screenshot() {
    local demo="$1"
    local type="$2"
    local tcl_name="$3"
    local out="$4"

    if [[ "${SKIP_IF_EXISTS:-0}" == "1" && -f "$out" ]]; then
        echo "  Skipping (exists): $out" >&2
        return 0
    fi

    bash "$SCRIPT_DIR/demo_screenshot.sh" "$demo" "$type" "$out" "$tcl_name"
}

echo "=== Comparing demo: $DEMO (Go) vs $TCL_DEMO (Tcl) ===" >&2

echo "[1/2] Go screenshot..." >&2
take_screenshot "$DEMO" "go" "$DEMO" "$GO_IMG"

echo "[2/2] Tcl screenshot..." >&2
take_screenshot "$DEMO" "tcl" "$TCL_DEMO" "$TCL_IMG"

# ---------------------------------------------------------------------------
# Normalize to the same canvas size (pad smaller image with white)
# ---------------------------------------------------------------------------
W_GO=$(identify -format "%w" "$GO_IMG")
H_GO=$(identify -format "%h" "$GO_IMG")
W_TCL=$(identify -format "%w" "$TCL_IMG")
H_TCL=$(identify -format "%h" "$TCL_IMG")

MAXW=$(( W_GO > W_TCL ? W_GO : W_TCL ))
MAXH=$(( H_GO > H_TCL ? H_GO : H_TCL ))

PAD_GO="$SS_DIR/${DEMO}_go_pad.png"
PAD_TCL="$SS_DIR/${TCL_DEMO}_tcl_pad.png"

magick "$GO_IMG"  -gravity NorthWest -background white -extent "${MAXW}x${MAXH}" "$PAD_GO"
magick "$TCL_IMG" -gravity NorthWest -background white -extent "${MAXW}x${MAXH}" "$PAD_TCL"

# ---------------------------------------------------------------------------
# Compute pixel diff
# ---------------------------------------------------------------------------
# compare returns exit code 1 when images differ (normal), 2 on error
SCORE=$(compare -metric MAE "$PAD_GO" "$PAD_TCL" "$DIFF_IMG" 2>&1 | awk '{print $1}') || true

# ---------------------------------------------------------------------------
# Side-by-side montage
# ---------------------------------------------------------------------------
montage \
    -label "Go: $DEMO"  "$GO_IMG" \
    -label "Tcl: $TCL_DEMO" "$TCL_IMG" \
    -label "Diff (MAE=$SCORE)" "$DIFF_IMG" \
    -tile 3x1 -geometry +4+4 \
    -background gray80 \
    "$SIDE_IMG"

# Cleanup padded intermediates
rm -f "$PAD_GO" "$PAD_TCL"

# ---------------------------------------------------------------------------
# Report
# ---------------------------------------------------------------------------
echo ""
echo "Demo:       $DEMO"
echo "Go image:   $GO_IMG  (${W_GO}x${H_GO})"
echo "Tcl image:  $TCL_IMG  (${W_TCL}x${H_TCL})"
echo "Diff score: $SCORE  (MAE — lower is more similar)"
echo "Side-by-side: $SIDE_IMG"
echo "Diff heatmap: $DIFF_IMG"
