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
#   <demo>_tree.txt     -- Structural diff (internal/cmd/demodiff), when both sides
#   <demo>_tree.json       wrote a widget-tree dump (DUMP_TREE=1, default)
#
# Prints a diff score (odiff diff percentage, lower = more similar) to stdout.
#
# Environment:
#   DISPLAY, SETTLE_SECS, TIMEOUT_SECS  (passed through to demo_screenshot.sh)
#   SKIP_IF_EXISTS=1                    (skip screenshot if file already exists)
#   HEADLESS=1                          (run screenshots under xvfb-run)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SS_DIR="$PROJECT_DIR/tmp/screenshots"

# shellcheck disable=SC1091
source "$SCRIPT_DIR/_lib.sh"

command -v odiff >/dev/null 2>&1 || { echo "odiff not found (required)" >&2; exit 1; }
command -v magick >/dev/null 2>&1 || { echo "magick (ImageMagick) not found (required)" >&2; exit 1; }

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

# ---------------------------------------------------------------------------
# assert_captured IMG SIDE
# Rejects missing/empty captures and essentially-black captures (an unmapped
# window / failed grab yields an all-black image).
# ---------------------------------------------------------------------------
assert_captured() {
    local img="$1" side="$2"
    [[ -s "$img" ]] || { echo "$side screenshot missing or empty: $img" >&2; exit 1; }
    local mean
    mean=$(magick identify -format "%[fx:mean]" "$img" 2>/dev/null) || mean=""
    if [[ -n "$mean" ]]; then
        local is_black
        is_black=$(awk -v m="$mean" 'BEGIN { print (m < 0.01) ? 1 : 0 }')
        if [[ "$is_black" == "1" ]]; then
            echo "$side screenshot is blank/black: $img (mean=$mean)" >&2
            exit 1
        fi
    fi
}

echo "=== Comparing demo: $DEMO (Go) vs $TCL_DEMO (Tcl) ===" >&2

echo "[1/2] Go screenshot..." >&2
take_screenshot "$DEMO" "go" "$DEMO" "$GO_IMG"
assert_captured "$GO_IMG" "Go"

echo "[2/2] Tcl screenshot..." >&2
take_screenshot "$DEMO" "tcl" "$TCL_DEMO" "$TCL_IMG"
assert_captured "$TCL_IMG" "Tcl"

# ---------------------------------------------------------------------------
# Normalize to the same canvas size (pad smaller image with white)
# ---------------------------------------------------------------------------
read -r W_GO H_GO <<<"$(magick identify -format "%w %h" "$GO_IMG")"
read -r W_TCL H_TCL <<<"$(magick identify -format "%w %h" "$TCL_IMG")"

MAXW=$(( W_GO > W_TCL ? W_GO : W_TCL ))
MAXH=$(( H_GO > H_TCL ? H_GO : H_TCL ))

PAD_GO="$SS_DIR/${DEMO}_go_pad.png"
PAD_TCL="$SS_DIR/${TCL_DEMO}_tcl_pad.png"

magick "$GO_IMG"  -gravity NorthWest -background white -extent "${MAXW}x${MAXH}" "$PAD_GO"
magick "$TCL_IMG" -gravity NorthWest -background white -extent "${MAXW}x${MAXH}" "$PAD_TCL"

# ---------------------------------------------------------------------------
# Compute pixel diff with odiff (anti-aliasing ignored)
# ---------------------------------------------------------------------------
SCORE=$(odiff_score "$PAD_GO" "$PAD_TCL" "$DIFF_IMG") || {
    echo "odiff failed for $DEMO" >&2
    rm -f "$PAD_GO" "$PAD_TCL"
    exit 1
}

# ---------------------------------------------------------------------------
# Side-by-side montage
# ---------------------------------------------------------------------------
montage \
    -label "Go: $DEMO"  "$GO_IMG" \
    -label "Tcl: $TCL_DEMO" "$TCL_IMG" \
    -label "Diff (%%=$SCORE)" "$DIFF_IMG" \
    -tile 3x1 -geometry +4+4 \
    -background gray80 \
    "$SIDE_IMG"

# Cleanup padded intermediates
rm -f "$PAD_GO" "$PAD_TCL"

# ---------------------------------------------------------------------------
# Structural diff of the widget trees
# ---------------------------------------------------------------------------
GO_TREE="$SS_DIR/${DEMO}_go.tree.json"
TCL_TREE="$SS_DIR/${TCL_DEMO}_tcl.tree.json"
TREE_TXT="$SS_DIR/${DEMO}_tree.txt"
TREE_JSON="$SS_DIR/${DEMO}_tree.json"
TREE_DIFFS="-"
if [[ -s "$GO_TREE" && -s "$TCL_TREE" ]] && DD=$(demodiff_bin); then
    "$DD" -go-png "$GO_IMG" -tcl-png "$TCL_IMG" "$GO_TREE" "$TCL_TREE" > "$TREE_TXT" || true
    "$DD" -json -go-png "$GO_IMG" -tcl-png "$TCL_IMG" "$GO_TREE" "$TCL_TREE" > "$TREE_JSON" || true
    TREE_DIFFS=$(sed -n 's/^summary: \([0-9]*\) diffs.*/\1/p' "$TREE_TXT")
    [[ -z "$TREE_DIFFS" ]] && TREE_DIFFS="-"
fi

# ---------------------------------------------------------------------------
# Report
# ---------------------------------------------------------------------------
echo ""
echo "Demo:       $DEMO"
echo "Go image:   $GO_IMG  (${W_GO}x${H_GO})"
echo "Tcl image:  $TCL_IMG  (${W_TCL}x${H_TCL})"
echo "Diff score: $SCORE  (odiff diff % 0-100 — lower is more similar)"
echo "Tree diffs: $TREE_DIFFS  (structural differences, see $TREE_TXT)"
echo "Side-by-side: $SIDE_IMG"
echo "Diff heatmap: $DIFF_IMG"
