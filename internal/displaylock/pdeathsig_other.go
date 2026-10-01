//go:build unix && !linux

package displaylock

import "os/exec"

// dieWithParent does nothing here, so a test run may leave its Xvfb behind;
// Xvfb is rarely installed on these systems anyway.
func dieWithParent(*exec.Cmd) {}
