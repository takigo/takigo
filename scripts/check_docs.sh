#!/usr/bin/env bash
# Check that the agent-facing documents still describe this tree: every repo
# path they name in backticks exists, `file:NN` references point inside the
# file, every `make <goal>` they mention is a Makefile goal and the demo
# count they quote is the number of demos. Run from anywhere; CI runs it.
set -euo pipefail
cd "$(dirname "$0")/.."

files=(AGENTS.md CONTRIBUTING.md THREADING.md scripts/README.md)
while IFS= read -r f; do files+=("$f"); done < <(find . -mindepth 2 -maxdepth 3 -name AGENTS.md -not -path './tmp/*' -not -path './tk/*' -not -path './tcl/*' | sed 's|^\./||' | sort)
while IFS= read -r f; do files+=("$f"); done < <(find .agents/skills -name '*.md' | sort)

fail=0
checked=0
problem() { echo "$1: $2"; fail=1; }

# Paths written outside the repo: the Tk sources (gitignored), scratch
# output, module paths and Tk-relative paths such as library/button.tcl.
skip='^(tk|tcl|tmp|library|generic|unix|github\.com|golang\.org|proxy\.golang\.org|tcltk)(/|$)'

for f in "${files[@]}"; do
    while IFS= read -r tok; do
        tok=${tok#\`}; tok=${tok%\`}
        tok=${tok#./}
        [[ $tok =~ ^[A-Za-z0-9_.][A-Za-z0-9_./-]*(:[0-9]+(-[0-9]+)?)?$ ]] || continue
        [[ $tok == */* || $tok == *.md ]] || continue
        [[ $tok =~ $skip ]] && continue
        [[ $tok == *..* ]] && continue
        # Go names such as grid.Row/ColumnWeight or Width/Height.
        [[ ${tok%/*} == *[A-Z]* ]] && continue
        path=${tok%%:*}
        checked=$((checked + 1))
        if [[ ! -e $path ]]; then
            if [[ -e $(dirname "$f")/$path ]]; then
                path=$(dirname "$f")/$path
            else
                problem "$f" "missing path $tok"
                continue
            fi
        fi
        if [[ $tok == *:* ]]; then
            line=${tok##*:}; line=${line%%-*}
            n=$(wc -l < "$path")
            (( line <= n )) || problem "$f" "$tok: file has only $n lines"
        fi
    done < <(grep -oE '`[^` ]+`' "$f" | sort -u)

    while IFS= read -r goal; do
        checked=$((checked + 1))
        grep -qE "^$goal:" Makefile || problem "$f" "no Makefile goal '$goal'"
    done < <(grep -oE '`make [a-z][a-z-]*(`| [A-Z[])' "$f" | sed -E 's/`make ([a-z-]+).*/\1/' | sort -u)
done

demos=$(ls -d demos/*/main.go | wc -l)
for f in AGENTS.md demos/AGENTS.md README.md; do
    [[ -f $f ]] || continue
    while IFS= read -r n; do
        checked=$((checked + 1))
        [[ $n == "$demos" ]] || problem "$f" "says $n demos, there are $demos"
    done < <(grep -oE '\*{0,2}[0-9]+\*{0,2} demos\b' "$f" | tr -d '*' | awk '{print $1}')
done

if (( fail )); then exit 1; fi
echo "docs: all $checked references resolve"
