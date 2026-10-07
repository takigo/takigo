#!/usr/bin/env bash
# Compile every complete program ("package main" code block) in docs/tutorial.md.
# tmp/ is a module of its own (tmp/go.mod, written here and by scripts/_lib.sh)
# so that scratch programs under it never break `go build ./...`; the programs
# are checked from a nested module that replaces takigo with this checkout.
set -euo pipefail
cd "$(dirname "$0")/.."
root=$PWD
[[ -f tmp/go.mod ]] || { mkdir -p tmp; printf 'module takigo-tmp\n\ngo 1.27.0\n' > tmp/go.mod; }
out=tmp/_tutorial_check
rm -rf "$out"; mkdir -p "$out"
awk -v out="$out" '
/^```go$/ { inblock=1; buf=""; next }
/^```$/ && inblock { inblock=0; if (buf ~ /(^|\n)package main\n/) { n++; d=sprintf("%s/p%02d", out, n); system("mkdir -p " d); printf "%s", buf > (d "/main.go"); close(d "/main.go") } next }
inblock { buf = buf $0 "\n" }
' docs/tutorial.md
cat > "$out/go.mod" <<MOD
module tutorialcheck

go $(go list -m -f '{{.GoVersion}}')

require github.com/takigo/takigo v0.0.0

replace github.com/takigo/takigo => $root
MOD
fail=0
for d in "$out"/p*/; do
    p=$(basename "$d")
    if ! (cd "$out" && go vet "./$p" 2>"$p/err.txt"); then
        fail=1
        echo "FAIL $d ($(grep -n -m1 -F "$(sed -n 2,4p "$d/main.go" | head -1)" docs/tutorial.md | cut -d: -f1 || true))"
        cat "$d/err.txt"
    fi
done
[ "$fail" = 0 ] && echo "tutorial programs: all $(ls -d "$out"/p*/ | wc -l) compile"
exit "$fail"
