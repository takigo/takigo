//go:build windows

package appearance

import (
	"syscall"
	"unsafe"
)

// system reads AppsUseLightTheme, which is 0 when apps are asked to be dark.
func system() Mode {
	m, _ := cachedLookup(registryMode)
	return m
}

func registryMode() (Mode, bool) {
	path, err := syscall.UTF16PtrFromString(`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`)
	if err != nil {
		return Light, false
	}
	var key syscall.Handle
	if syscall.RegOpenKeyEx(syscall.HKEY_CURRENT_USER, path, 0, syscall.KEY_READ, &key) != nil {
		return Light, false
	}
	defer syscall.RegCloseKey(key)

	name, err := syscall.UTF16PtrFromString("AppsUseLightTheme")
	if err != nil {
		return Light, false
	}
	var typ, value uint32
	size := uint32(unsafe.Sizeof(value))
	if syscall.RegQueryValueEx(key, name, nil, &typ, (*byte)(unsafe.Pointer(&value)), &size) != nil {
		return Light, false
	}
	if typ == syscall.REG_DWORD && value == 0 {
		return Dark, true
	}
	return Light, true
}
