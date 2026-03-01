package xlib

import "errors"

var (
	ErrNoDisplay = errors.New("xlib: cannot open display")
)
