// Package stubs holds the fakes that display-free tests share: a geometry
// manager that counts the requests it gets and a display server whose
// window operations do nothing, enough for windows to be laid out without
// a display.
package stubs

import (
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

// Manager is a window.GeomManager that counts RequestProc calls.
type Manager struct{ Requests int }

func (*Manager) Name() string                   { return "stub" }
func (m *Manager) RequestProc(*window.Window)   { m.Requests++ }
func (*Manager) LostContentProc(*window.Window) {}

// Server is a platform.DisplayServer whose window operations do nothing;
// every other method panics, so a test that reaches one knows.
type Server struct{ platform.DisplayServer }

func (Server) MoveResizeWindow(platform.WindowID, int, int, uint, uint) {}
func (Server) MoveWindow(platform.WindowID, int, int)                   {}
func (Server) ResizeWindow(platform.WindowID, uint, uint)               {}
func (Server) MapWindow(platform.WindowID)                              {}
func (Server) UnmapWindow(platform.WindowID)                            {}
func (Server) DestroyWindow(platform.WindowID)                          {}
func (Server) Flush()                                                   {}

// Display returns a window.Display on a Server with an empty window table.
func Display() *window.Display {
	return &window.Display{Server: Server{}, Windows: map[platform.WindowID]*window.Window{}}
}
