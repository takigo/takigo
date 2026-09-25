#!/usr/bin/env bash
# demo_gate.sh -- Compare the latest batch results with the committed baseline
#
# Usage:
#   demo_gate.sh [--tolerance PCT] [--accept]
#
# Reads tmp/screenshots/scores.tsv (written by demo_batch.sh) and
# demos/parity.tsv (the committed baseline). A demo regresses when its pixel
# score rises by more than PCT (default 0.1) or its tree diff count rises.
# Prints improvements and regressions; exits 1 if anything regressed.
# --accept writes the latest scores into demos/parity.tsv after reporting
# (use it once the listed regressions are understood).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/_lib.sh"

TOL=0.1
ACCEPT=0
while [[ $# -gt 0 ]]; do
    case "$1" in
        --tolerance) shift; TOL="$1" ;;
        --accept)    ACCEPT=1 ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
    shift
done

SCORES="$SS_DIR/scores.tsv"
BASELINE="$PROJECT_DIR/demos/parity.tsv"
[[ -f "$SCORES" ]] || { echo "no $SCORES — run demo_batch.sh first" >&2; exit 1; }
[[ -f "$BASELINE" ]] || { echo "no $BASELINE" >&2; exit 1; }

rc=0
awk -F'\t' -v tol="$TOL" '
    FNR == NR {
        if ($0 ~ /^#/ || NF < 7) next
        bpix[$1] = $3; btree[$1] = $4
        next
    }
    FNR == 1 { next }
    {
        d = $1; pix = $3; tree = $4
        if (!(d in bpix)) { printf "NEW       %-14s pixel %s tree %s\n", d, pix, tree; next }
        dp = pix - bpix[d]
        dt = (tree == "-" || btree[d] == "-") ? 0 : tree - btree[d]
        if (dp > tol || dt > 0) {
            printf "REGRESSED %-14s pixel %6s -> %6s  tree %3s -> %3s\n", d, bpix[d], pix, btree[d], tree
            bad++
        } else if (dp < -tol || dt < 0) {
            printf "improved  %-14s pixel %6s -> %6s  tree %3s -> %3s\n", d, bpix[d], pix, btree[d], tree
            good++
        } else {
            same++
        }
        tp += pix; tb += bpix[d]
        if (tree != "-") tt += tree
        if (btree[d] != "-") tbt += btree[d]
    }
    END {
        printf "\n%d improved, %d regressed, %d unchanged; total pixel %.2f -> %.2f, total tree diffs %d -> %d\n",
            good, bad, same, tb, tp, tbt, tt
        exit (bad > 0)
    }' "$BASELINE" "$SCORES" || rc=$?

if [[ "$ACCEPT" == "1" ]]; then
    echo ""
    update_baseline "$SCORES"
    exit 0
fi
exit "$rc"
