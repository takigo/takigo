#!/usr/bin/env bash
# Compile every complete program ("package main" code block) in docs/tutorial.md.
# The output directory starts with "_" so ./... does not pick it up.
set -euo pipefail
cd "$(dirname "$0")/.."
out=tmp/_tutorial_check
rm -rf "$out"; mkdir -p "$out"
awk -v out="$out" '
/^```go$/ { inblock=1; buf=""; next }
/^```$/ && inblock { inblock=0; if (buf ~ /(^|\n)package main\n/) { n++; d=sprintf("%s/p%02d", out, n); system("mkdir -p " d); printf "%s", buf > (d "/main.go"); close(d "/main.go") } next }
inblock { buf = buf $0 "\n" }
' docs/tutorial.md
fail=0
for d in "$out"/p*/; do
    if ! go vet "./$d" 2>"$d/err.txt"; then
        fail=1
        echo "FAIL $d ($(grep -n -m1 -F "$(sed -n 2,4p "$d/main.go" | head -1)" docs/tutorial.md | cut -d: -f1 || true))"
        cat "$d/err.txt"
    fi
done
[ "$fail" = 0 ] && echo "tutorial programs: all $(ls -d "$out"/p*/ | wc -l) compile"
exit "$fail"
