#!/usr/bin/env bash
# PostToolUse hook: gofmt a .go file right after Claude edits or writes it,
# so formatting never reaches `make check`. Reads the hook JSON on stdin.
# Without jq or gofmt it does nothing.
command -v jq >/dev/null && command -v gofmt >/dev/null || exit 0
f=$(jq -r '.tool_input.file_path // empty')
[[ $f == *.go && -f $f ]] || exit 0
case $f in */tk/* | */tcl/*) exit 0 ;; esac
out=$(gofmt -l -w "$f" 2>&1) || { echo "gofmt: $out" >&2; exit 2; }
exit 0
