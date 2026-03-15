//go:build linux || freebsd || openbsd || netbsd

package xlib

import "errors"

var (
	ErrNoDisplay = errors.New("xlib: cannot open display")
)
