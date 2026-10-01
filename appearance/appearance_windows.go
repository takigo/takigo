//go:build windows

package appearance

import (
	"syscall"
	"unsafe"
)

// system reads AppsUseLightTheme, which is 0 when apps are asked to be dark.
func system() Mode {
	path, err := syscall.UTF16PtrFromString(`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`)
	if err != nil {
		return Light
	}
	var key syscall.Handle
	if syscall.RegOpenKeyEx(syscall.HKEY_CURRENT_USER, path, 0, syscall.KEY_READ, &key) != nil {
		return Light
	}
	defer syscall.RegCloseKey(key)

	name, err := syscall.UTF16PtrFromString("AppsUseLightTheme")
	if err != nil {
		return Light
	}
	var typ, value uint32
	size := uint32(unsafe.Sizeof(value))
	if syscall.RegQueryValueEx(key, name, nil, &typ, (*byte)(unsafe.Pointer(&value)), &size) != nil {
		return Light
	}
	if typ == syscall.REG_DWORD && value == 0 {
		return Dark
	}
	return Light
}
