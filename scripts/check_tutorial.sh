#!/usr/bin/env bash
# Compile every Go code block in docs/tutorial.md.
#
# The complete programs ("package main") are compiled as they are. The other
# blocks are fragments: each one is wrapped in a generated function together
# with a preamble that declares the names the fragment uses, and all of them
# are compiled in one throwaway module that replaces takigo with this
# checkout. The module lives in a temporary directory, so no file is left
# behind and the repository's go.mod is never touched.
#
# How a block is classified:
#
#   program  it says "package main": compiled as it is
#   file     it declares a function or a type at the top level: the
#            declarations are compiled as they are
#   split    it starts with an import block: the imports are compiled as they
#            are, the rest of the block becomes statements
#   stmt     the rest: the block becomes statements of a generated function
#   expr     a block that is a single expression: it becomes "_ = <expr>"
#
# The tutorial writes "..." where more code would go. Such a placeholder is
# dropped, so that a fragment around it still compiles.
set -euo pipefail
cd "$(dirname "$0")/.."
root=$PWD

work=$(mktemp -d "${TMPDIR:-/tmp}/takigo-tutorial.XXXXXX")
trap 'rm -rf "$work"' EXIT

# --------------------------------------------------------------------- #
# 1. extract the Go blocks
# --------------------------------------------------------------------- #
mkdir -p "$work/blocks"
awk -v out="$work/blocks" '
/^```go[ \t]*$/ { inblock = 1; buf = ""; n++; start = NR + 1; next }
/^```[ \t]*$/ && inblock {
	inblock = 0
	f = sprintf("%s/b%03d.go", out, n)
	printf "%s", buf > f
	close(f)
	printf "%d\t%d\n", n, start
	next
}
inblock { buf = buf $0 "\n" }
' docs/tutorial.md > "$work/blocks/index.txt"

# --------------------------------------------------------------------- #
# 2. map package names to import paths
# --------------------------------------------------------------------- #
{
	while IFS= read -r dir; do
		rel=${dir#./}
		[ "$rel" = "." ] && rel=""
		f=""
		for g in "$dir"/*.go; do
			case $g in *_test.go) continue ;; esac
			f=$g
			break
		done
		[ -n "$f" ] || continue
		name=$(awk '/^package /{print $2; exit}' "$f")
		[ -n "$name" ] || continue
		case $name in main) continue ;; esac
		if [ -z "$rel" ]; then
			printf '%s\t%s\n' "$name" "github.com/takigo/takigo"
		else
			printf '%s\t%s\n' "$name" "github.com/takigo/takigo/$rel"
		fi
	done < <(find . -name '*.go' \
		-not -path '*/.git/*' -not -path '*/tmp/*' -not -path '*/tk/*' \
		-not -path '*/tcl/*' -not -path '*/internal/*' -not -path '*/demos/*' \
		-not -path '*/tests/*' -not -path '*/docs/*' -not -path '*/scripts/*' \
		| xargs -n1 dirname | sort -u)
	for p in fmt log os strconv strings time errors context io sort testing \
		sync math bytes reflect unicode; do
		printf '%s\t%s\n' "$p" "$p"
	done
	printf 'path/filepath\tpath/filepath\n'
	printf 'encoding/json\tencoding/json\n'
} > "$work/pkgmap.txt"

# Aliases the tutorial uses for standard library packages.
cat > "$work/alias.txt" <<'ALIAS'
goimage	image
gocolor	image/color
ALIAS

# --------------------------------------------------------------------- #
# 3. preamble table: the names a fragment uses but does not declare
#
# The blocks are numbered from 1 in the order they appear in the tutorial.
# A preamble declares the free names of that block; a name the block
# declares itself must not appear here.
# --------------------------------------------------------------------- #
cat > "$work/preambles.txt" <<'PRE'
### 2
var app *takigo.App
### 3
var parent widget.Caregiver

var onOK = func() {}
### 4
var b *button.Button

var onCancel = func() {}
### 5
var ok *button.Button
### 7
var app *takigo.App
### 8
var w *frame.Frame
### 9
var toolbar, status, sidebar, content *frame.Frame
### 10
var ok, cancel, help *frame.Frame
### 11
var sidebar *frame.Frame
### 12
var searchBar, content *frame.Frame
### 13
var nameLabel, nameEntry *frame.Frame
var parent widget.Caregiver
### 14
var f *frame.Frame
### 15
var entry *entry.Entry
### 16
var a, b, c, d *frame.Frame
### 18
var app *takigo.App
var badge, bar, corner *frame.Frame
### 19
var app *takigo.App
var lbl *label.Label
### 20
var app *takigo.App
### 21
var app *takigo.App
### 22
var app *takigo.App
### 23
var app *takigo.App

var save = func() {}
### 24
var app *takigo.App
### 25
var f *frame.Frame
### 26
var app *takigo.App

var setVolume = func(v float64) {}
### 27
var app *takigo.App
### 28
var app *takigo.App
### 29
var app *takigo.App
### 32
var app *takigo.App
var entry *entry.Entry

var search = func(q string) {}
### 33
var eng *bind.Engine
### 34
var eng *bind.Engine
var status *label.Label
### 35
var eng *bind.Engine
var editor, list *frame.Frame

var save = func(*bind.EventData) bool { return false }
var open = func(*bind.EventData) bool { return false }
var copySelection = func(*bind.EventData) bool { return false }
### 36
var eng *bind.Engine
var someWidget *frame.Frame

var save = func() {}
### 37
var app *takigo.App
var c *frame.Frame
### 38
var app *takigo.App
var entry *entry.Entry
### 39
var body *frame.Frame
### 41
var app *takigo.App

var open = func() {}
var save = func() {}
### 42
var app *takigo.App
var target *frame.Frame

var doCopy = func() {}
var doPaste = func() {}
### 43
var app *takigo.App

var setSize = func(s string) {}
### 45
var app *takigo.App
### 46
var app *takigo.App
### 47
var app *takigo.App
### 48
var app *takigo.App
### 49
var app *takigo.App
var oldName string
var swatch *label.Label
var tw *text.TextWidget

var rename = func(name string) {}
### 51
var app *takigo.App
### 52
var tw *text.TextWidget
### 53
var tw *text.TextWidget
### 54
var tw *text.TextWidget
### 55
var tw *text.TextWidget

var openURL = func() {}
### 56
var tw *text.TextWidget
### 57
var tw *text.TextWidget
### 58
var tw *text.TextWidget
var img *image.Photo

var showMore = func() {}
### 59
var app *takigo.App
var tw *text.TextWidget
### 60
var app *takigo.App
var img *image.Photo
### 61
var c *canvas.Canvas
var rect canvas.ItemID
### 62
var c *canvas.Canvas
var x, y float64
### 63
var c *canvas.Canvas
### 65
var f *frame.Frame
var hsb, vsb *scrollbar.Scrollbar
### 66
var c *canvas.Canvas
### 67
var app *takigo.App
### 68
var app *takigo.App
### 69
var app *takigo.App
### 70
var app *takigo.App
var list *frame.Frame
var selected []string

var addFile = func(path string) {}
### 71
var app *takigo.App
var status *label.Label
### 72
var app *takigo.App
var clock *ttk.Label
### 73
var app *takigo.App
var url string
var status *label.Label

var fetch = func(u string) (any, error) { return nil, nil }
var show = func(data any) {}
### 74
var app *takigo.App
var lb *listbox.Listbox

var use = func(n int) {}
### 77
var root *ttk.Frame
### 78
var page *ttk.Frame
var fileIcon widget.WidgetImage

var sortBySize = func() {}
var loadChildrenLazily = func(id string) {}
var openItem = func(id string) {}
### 79
var app *takigo.App
var f *ttk.Frame
PRE

# Fragments that are listings rather than Go code and cannot be compiled:
# "<block number><TAB>reason". Every other fragment must compile.
cat > "$work/skip.txt" <<'SKIP'
SKIP

# Fragments that need another mode than the automatic one:
# "<block number><TAB>mode", with mode one of stmt, split, file, expr.
cat > "$work/modes.txt" <<'MODES'
MODES

# --------------------------------------------------------------------- #
# 4. generate one file per fragment
# --------------------------------------------------------------------- #
mkdir -p "$work/pre"
awk -v dir="$work/pre" '
/^### [0-9]+$/ { id = $2; next }
{ if (id != "") print $0 > (dir "/" id ".txt") }
' "$work/preambles.txt"

cat > "$work/gen.awk" <<'GEN'
# gen.awk: turn one tutorial block into a compilable Go file.
#
#   mode=stmt|split|file|expr
#                       stmt: the block is a list of statements
#                       split: the block starts with an import block
#                       file: the block holds top level declarations
#                       expr: the block is a single expression
#   pkgmap, alias: "name<TAB>path" tables
#   prefile: preamble file (may be empty)
#   out: generated file, fn: name of the generated function
function iskw(w) {
	return (w == "var" || w == "for" || w == "if" || w == "case" ||
		w == "switch" || w == "range" || w == "return" || w == "else" ||
		w == "func" || w == "go" || w == "defer" || w == "select" ||
		w == "type" || w == "struct" || w == "interface" || w == "const" ||
		w == "package" || w == "import" || w == "chan" || w == "map" ||
		w == "make" || w == "new" || w == "len" || w == "cap" ||
		w == "append" || w == "copy" || w == "delete" || w == "panic" ||
		w == "recover" || w == "print" || w == "println" || w == "_")
}

function tokenize(s,   n) {
	delete TOK
	n = 0
	while (match(s, /[A-Za-z_][A-Za-z0-9_]*/)) {
		TOK[++n] = substr(s, RSTART, RLENGTH)
		s = substr(s, RSTART + RLENGTH)
	}
	return n
}

BEGIN {
	while ((getline l < pkgmap) > 0) { split(l, a, "\t"); PKG[a[1]] = a[2] }
	while ((getline l < alias) > 0) { split(l, a, "\t"); ALIAS[a[1]] = a[2] }
	nPre = 0
	while ((getline l < prefile) > 0) PRE[++nPre] = l
}
{ L[NR] = $0; N = NR }
END {
	# --- the block's own import block -------------------------------
	impStart = 0; impEnd = 0
	for (i = 1; i <= N; i++) {
		if (L[i] ~ /^import[ \t]*\(/) {
			impStart = i
			for (j = i + 1; j <= N; j++)
				if (L[j] ~ /^[ \t]*\)/) { impEnd = j; break }
			break
		}
		if (L[i] ~ /^import[ \t]+/) { impStart = i; impEnd = i; break }
	}
	nOwn = 0
	for (i = impStart; i <= impEnd && impStart > 0; i++) {
		l = L[i]
		sub(/\/\/.*/, "", l)
		if (match(l, /"[^"]+"/) == 0) continue
		path = substr(l, RSTART + 1, RLENGTH - 2)
		pre = substr(l, 1, RSTART - 1)
		sub(/^import[ \t]+/, "", pre)
		gsub(/^[ \t]+|[ \t]+$/, "", pre)
		if (pre == "" || pre == "import" || pre == "(") {
			n = split(path, parts, "/")
			name = parts[n]
			alias = ""
		} else {
			name = pre
			alias = pre
		}
		BOUND[name] = 1
		nOwn++
		OWNPATH[nOwn] = path
		OWNALIAS[nOwn] = alias
	}

	# --- names the block declares -----------------------------------
	for (i = 1; i <= N; i++) {
		l = L[i]
		gsub(/"[^"]*"/, "\"\"", l)
		sub(/\/\/.*/, "", l)
		if (match(l, /:=/)) {
			head = substr(l, 1, RSTART - 1)
			n = tokenize(head)
			for (k = 1; k <= n; k++)
				if (!iskw(TOK[k])) DECL[TOK[k]] = 1
		}
		if (match(l, /(^|[^A-Za-z0-9_])var[ \t]+[A-Za-z_][A-Za-z0-9_]*/)) {
			head = substr(l, RSTART, RLENGTH)
			sub(/^[^A-Za-z0-9_]*var[ \t]+/, "", head)
			DECL[head] = 1
		}
	}

	# --- package qualifiers the block and the preamble use ------------
	for (i = 1; i <= N + nPre; i++) {
		if (i <= N) {
			if (impStart > 0 && i >= impStart && i <= impEnd) continue
			l = L[i]
		} else {
			l = PRE[i - N]
		}
		gsub(/"[^"]*"/, "\"\"", l)
		sub(/\/\/.*/, "", l)
		while (match(l, /[A-Za-z_][A-Za-z0-9_]*\./)) {
			USED[substr(l, RSTART, RLENGTH - 1)] = 1
			l = substr(l, RSTART + RLENGTH)
		}
	}

	# --- names declared at the top level of the block ----------------
	# (a fragment may leave them unused, which Go rejects)
	nTop = 0
	for (i = 1; i <= N; i++) {
		if (impStart > 0 && i >= impStart && i <= impEnd) continue
		l = L[i]
		if (l ~ /^[ \t]/) continue
		gsub(/"[^"]*"/, "\"\"", l)
		if (match(l, /:=/)) {
			head = substr(l, 1, RSTART - 1)
			# a name declared by a for, if, case or switch lives in a
			# nested scope and cannot be used at the top level
			if (head !~ /(^|[^A-Za-z0-9_])(for|if|case|switch)[ \t]/) {
				n = tokenize(head)
				for (k = 1; k <= n; k++) {
					if (iskw(TOK[k])) continue
					if (!(TOK[k] in TOP)) TOP[nTop++] = TOK[k]
				}
			}
		} else if (match(l, /^var[ \t]+[A-Za-z_][A-Za-z0-9_]*/)) {
			head = substr(l, RSTART, RLENGTH)
			sub(/^var[ \t]+/, "", head)
			if (!(head in TOP)) TOP[nTop++] = head
		}
	}

	# --- emit ------------------------------------------------------
	printf "package fragcheck\n\n" > out
	printf "import (\n" > out
	for (i = 1; i <= nOwn; i++) {
		if (OWNALIAS[i] == "")
			printf "\t\"%s\"\n", OWNPATH[i] > out
		else
			printf "\t%s \"%s\"\n", OWNALIAS[i], OWNPATH[i] > out
	}
	for (q in USED) {
		if (q in BOUND) continue
		if (q in DECL) continue
		if (q in PKG) {
			n = split(PKG[q], parts, "/")
			if (parts[n] != q) continue
			printf "\t\"%s\"\n", PKG[q] > out
		} else if (q in ALIAS) {
			printf "\t%s \"%s\"\n", q, ALIAS[q] > out
		}
	}
	printf ")\n\n" > out

	if (mode == "file") {
		for (i = 1; i <= nPre; i++) print PRE[i] > out
		for (i = 1; i <= N; i++) {
			if (impStart > 0 && i >= impStart && i <= impEnd) continue
			print L[i] > out
		}
	} else {
		printf "func %s() {\n", fn > out
		for (i = 1; i <= nPre; i++) print "\t" PRE[i] > out
		if (mode == "expr") {
			# the block is a single expression: keep it as one
			expr = ""
			for (i = 1; i <= N; i++) {
				if (impStart > 0 && i >= impStart && i <= impEnd) continue
				expr = expr L[i] " "
			}
			gsub(/[ \t]+/, " ", expr)
			gsub(/^ /, "", expr)
			gsub(/ $/, "", expr)
			printf "\t_ = %s\n", expr > out
		} else {
			for (i = 1; i <= N; i++) {
				if (impStart > 0 && i >= impStart && i <= impEnd) continue
				print "\t" L[i] > out
			}
		}
		for (i = 0; i < nTop; i++) print "\t_ = " TOP[i] > out
		printf "}\n" > out
	}
	close(out)
}
GEN

mkdir -p "$work/frag"
nfrag=0
nprog=0
nskip=0
while IFS=$'\t' read -r n _; do
	block="$work/blocks/b$(printf '%03d' "$n").go"
	# The tutorial writes "..." where more code would go; such a
	# placeholder must not stop the fragment from compiling.
	sed -e 's/,[[:space:]]*\.\.\.[[:space:]]*)/)/g' \
		-e 's/{[[:space:]]*\.\.\.[[:space:]]*}/{}/g' \
		-e '/^[[:space:]]*\.\.\.[[:space:]]*$/d' "$block" > "$work/blocks/f$n.go"
	if grep -q '^package main' "$work/blocks/f$n.go"; then
		nprog=$((nprog + 1))
		d="$work/prog/p$(printf '%03d' "$n")"
		mkdir -p "$d"
		cp "$work/blocks/f$n.go" "$d/main.go"
		continue
	fi
	if grep -q "^$n	" "$work/skip.txt"; then
		nskip=$((nskip + 1))
		continue
	fi
	nfrag=$((nfrag + 1))
	# classify the fragment
	mode=$(awk -v i="$n" '$1 == i {print $2}' "$work/modes.txt")
	case $mode in
	stmt | split | file | expr) ;;
	*)
		mode=stmt
		if grep -q '^func[ \t]' "$work/blocks/f$n.go"; then
			mode=file
		elif grep -q '^import' "$work/blocks/f$n.go"; then
			mode=split
		fi
		;;
	esac
	pre="$work/pre/$n.txt"
	[ -f "$pre" ] || : > "$pre"
	awk -v mode="$mode" -v pkgmap="$work/pkgmap.txt" -v alias="$work/alias.txt" \
		-v prefile="$pre" -v out="$work/frag/frag$n.go" -v fn="frag$n" \
		-f "$work/gen.awk" "$work/blocks/f$n.go"
done < "$work/blocks/index.txt"

# --------------------------------------------------------------------- #
# 5. compile
# --------------------------------------------------------------------- #
req=$(go list -m -f '{{.GoVersion}}' 2>/dev/null || echo 1.27.0)
localv=$(go env GOVERSION | sed 's/^go//')
if [ "$(printf '%s\n%s\n' "$req" "$localv" | sort -V | tail -1)" = "$localv" ]; then
	# The local toolchain is new enough: the check must not need a network.
	export GOPROXY=off
fi
export GOFLAGS=-mod=mod

cat > "$work/go.mod" <<MOD
module tutorialcheck

go $req

require github.com/takigo/takigo v0.0.0

replace github.com/takigo/takigo => $root
MOD

# Name the tutorial blocks a compiler error points at.
report() {
	awk '
	{
		if (match($0, /(frag|prog\/p)[0-9]+/)) {
			id = substr($0, RSTART, RLENGTH)
			gsub(/[^0-9]/, "", id)
			if (!(id in seen)) { seen[id] = 1; print id }
		}
	}' "$2" | while read -r id; do
		line=$(awk -v i="$id" '$1 == i {print $2}' "$work/blocks/index.txt")
		echo "FAIL $1 $id (docs/tutorial.md line $line)"
	done
}

fail=0
for d in "$work"/prog/*/; do
	[ -d "$d" ] || continue
	rel=${d#"$work"/}
	rel=${rel%/}
	if ! (cd "$work" && go vet "./$rel" 2>"$d/err.txt"); then
		fail=1
		report program "$d/err.txt"
		cat "$d/err.txt"
	fi
done
if [ -d "$work/frag" ] && [ -n "$(ls -A "$work/frag")" ]; then
	# -e makes the compiler list every error, not just the first one.
	if ! (cd "$work" && go build -gcflags=-e ./frag 2>"$work/frag-err.txt"); then
		fail=1
		report fragment "$work/frag-err.txt"
		cat "$work/frag-err.txt"
	elif ! (cd "$work" && go vet ./frag 2>"$work/frag-err.txt"); then
		fail=1
		report fragment "$work/frag-err.txt"
		cat "$work/frag-err.txt"
	fi
fi

if [ "$nskip" -gt 0 ]; then
	echo "skipped $nskip fragment(s):"
	cat "$work/skip.txt"
fi
if [ "$fail" = 0 ]; then
	echo "tutorial programs: all $nprog compile, $nfrag fragments compile"
fi
exit "$fail"
