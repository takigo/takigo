//go:build darwin

package platform

// CommandMask is the state bit the Cocoa backend reports for the Command
// key (Tk's Mod2 on aqua). On X11 and Windows Mod2 is NumLock, so there it
// is zero.
const CommandMask = Mod2Mask
