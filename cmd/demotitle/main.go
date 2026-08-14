// Command demotitle extracts the first takigo.Title("...") string from a Go
// source file. It is invoked by scripts/demo_screenshot.sh to discover the
// window title a demo expects, replacing a fragile grep -oP pattern that
// required GNU PCRE and broke on escaped strings or commented-out lines.
//
// Usage:
//
//	demotitle <path-to-main.go>            prints the Title string
//	demotitle -geometry <path-to-main.go>  prints the Geometry string (or empty)
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: demotitle [-geometry] <path-to-go-file>")
		os.Exit(2)
	}

	geometry := false
	args := os.Args[1:]
	if args[0] == "-geometry" {
		geometry = true
		args = args[1:]
	}
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: demotitle [-geometry] <path-to-go-file>")
		os.Exit(2)
	}
	path := args[0]

	src, err := os.ReadFile(path)
	if err != nil {
		os.Exit(1)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.SkipObjectResolution)
	if err != nil {
		os.Exit(1)
	}

	method := "Title"
	if geometry {
		method = "Geometry"
	}
	if v := findStringArg(file, method); v != "" {
		fmt.Print(v)
	} else {
		os.Exit(1)
	}
}

// findStringArg walks the AST looking for the first call expression whose
// function name is `method` (e.g. "Title", "Geometry"), accepts either the
// fully-qualified `takigo.Method(...)` form or a top-level call (in case
// the demo imports takigo as an alias), and returns the first string-literal
// argument unquoted.
func findStringArg(file *ast.File, method string) string {
	var found string
	ast.Inspect(file, func(n ast.Node) bool {
		if found != "" {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fn := call.Fun.(type) {
		case *ast.SelectorExpr:
			// takigo.Title(...) — selector on an Ident or selector chain.
			if fn.Sel.Name != method {
				return true
			}
			if id, ok := fn.X.(*ast.Ident); ok && id.Name != "takigo" {
				// E.g. `lbl.Title(...)` — not the takigo one. Bail.
				return true
			}
		case *ast.Ident:
			if fn.Name != method {
				return true
			}
		default:
			return true
		}
		if len(call.Args) < 1 {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		found = unquote(lit.Value)
		return false
	})
	return found
}

// unquote removes the surrounding quotes of a Go string literal. Handles
// both double-quoted strings (with possible escape sequences) and raw
// backtick strings.
func unquote(s string) string {
	if len(s) < 2 {
		return ""
	}
	if s[0] == '`' && s[len(s)-1] == '`' {
		return s[1 : len(s)-1]
	}
	if s[0] == '"' && s[len(s)-1] == '"' {
		if u, err := strconv.Unquote(s); err == nil {
			return u
		}
	}
	// Fallback: strip outer quotes only.
	return strings.TrimSuffix(strings.TrimPrefix(s, "\""), "\"")
}
