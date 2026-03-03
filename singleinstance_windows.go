//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	modKernel32SI    = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutexW = modKernel32SI.NewProc("CreateMutexW")
	procCloseHandle  = modKernel32SI.NewProc("CloseHandle")
	instanceMutex    syscall.Handle
)

const (
	errorAlreadyExists = 183
)

func ensureSingleInstance() bool {
	name, _ := syscall.UTF16PtrFromString("Local\\AutoLockSessionTimer")
	handle, _, err := procCreateMutexW.Call(0, 1, uintptr(unsafe.Pointer(name)))
	if handle == 0 {
		// Fail closed: if we cannot create/open mutex reliably, do not spawn another instance.
		return false
	}
	instanceMutex = syscall.Handle(handle)
	if errno, ok := err.(syscall.Errno); ok && errno == errorAlreadyExists {
		return false
	}
	return true
}

func releaseSingleInstance() {
	if instanceMutex == 0 {
		return
	}
	_, _, _ = procCloseHandle.Call(uintptr(instanceMutex))
	instanceMutex = 0
}
