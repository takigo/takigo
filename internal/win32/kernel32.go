//go:build windows

package win32

import (
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")
	procGetLastError        = kernel32.NewProc("GetLastError")
	procGlobalAlloc         = kernel32.NewProc("GlobalAlloc")
	procGlobalFree          = kernel32.NewProc("GlobalFree")
	procGlobalLock          = kernel32.NewProc("GlobalLock")
	procGlobalUnlock        = kernel32.NewProc("GlobalUnlock")
	procGlobalSize          = kernel32.NewProc("GlobalSize")
	procMultiByteToWideChar = kernel32.NewProc("MultiByteToWideChar")
	procWideCharToMultiByte = kernel32.NewProc("WideCharToMultiByte")
)

func GetModuleHandle(moduleName *uint16) HINSTANCE {
	r, _, _ := procGetModuleHandleW.Call(uintptr(unsafe.Pointer(moduleName)))
	return HINSTANCE(r)
}

func GetLastError() uint32 {
	r, _, _ := procGetLastError.Call()
	return uint32(r)
}

func GlobalAlloc(flags uint32, bytes uintptr) HGLOBAL {
	r, _, _ := procGlobalAlloc.Call(uintptr(flags), bytes)
	return HGLOBAL(r)
}

func GlobalFree(hMem HGLOBAL) HGLOBAL {
	r, _, _ := procGlobalFree.Call(uintptr(hMem))
	return HGLOBAL(r)
}

// GlobalLockPtr locks a global memory object and returns the raw pointer as uintptr.
func GlobalLockPtr(hMem HGLOBAL) uintptr {
	r, _, _ := procGlobalLock.Call(uintptr(hMem))
	return r
}

func GlobalUnlock(hMem HGLOBAL) bool {
	r, _, _ := procGlobalUnlock.Call(uintptr(hMem))
	return r != 0
}

func GlobalSize(hMem HGLOBAL) uintptr {
	r, _, _ := procGlobalSize.Call(uintptr(hMem))
	return r
}
