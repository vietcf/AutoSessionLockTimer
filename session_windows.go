//go:build windows

package main

import (
	"errors"
	"syscall"
	"unsafe"
)

const (
	wmWtsSessionChange = 0x02B1
	wtsSessionLogon    = 0x5
	wtsSessionLock     = 0x7
	wtsSessionUnlock   = 0x8
)

var (
	modWtsapi32                          = syscall.NewLazyDLL("wtsapi32.dll")
	procWTSRegisterSessionNotification   = modWtsapi32.NewProc("WTSRegisterSessionNotification")
	procWTSUnRegisterSessionNotification = modWtsapi32.NewProc("WTSUnRegisterSessionNotification")

	modUser32            = syscall.NewLazyDLL("user32.dll")
	procRegisterClassExW = modUser32.NewProc("RegisterClassExW")
	procCreateWindowExW  = modUser32.NewProc("CreateWindowExW")
	procDefWindowProcW   = modUser32.NewProc("DefWindowProcW")
	procGetMessageW      = modUser32.NewProc("GetMessageW")
	procTranslateMessage = modUser32.NewProc("TranslateMessage")
	procDispatchMessageW = modUser32.NewProc("DispatchMessageW")
	procDestroyWindow    = modUser32.NewProc("DestroyWindow")
	procPostQuitMessage  = modUser32.NewProc("PostQuitMessage")
	procLockWorkStation  = modUser32.NewProc("LockWorkStation")

	modKernel32          = syscall.NewLazyDLL("kernel32.dll")
	procGetModuleHandleW = modKernel32.NewProc("GetModuleHandleW")
)

type (
	wndProcFunc func(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr
)

type WNDCLASSEX struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   syscall.Handle
	Icon       syscall.Handle
	Cursor     syscall.Handle
	Background syscall.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     syscall.Handle
}

type MSG struct {
	HWnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct {
		X int32
		Y int32
	}
}

func listenSessionEvents(state *TimerState) error {
	className, _ := syscall.UTF16PtrFromString("AutoLockSessionTimerHiddenWindow")
	instance := getModuleHandle()

	wndProc := syscall.NewCallback(func(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
		switch msg {
		case wmWtsSessionChange:
			switch wParam {
			case wtsSessionLogon:
				state.onUnlock()
			case wtsSessionUnlock:
				state.onUnlock()
			case wtsSessionLock:
				state.onLock()
			}
		}
		ret, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
		return ret
	})

	wcx := WNDCLASSEX{
		Size:      uint32(unsafe.Sizeof(WNDCLASSEX{})),
		WndProc:   wndProc,
		Instance:  instance,
		ClassName: className,
	}

	if r, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wcx))); r == 0 {
		if err != syscall.Errno(0) {
			return err
		}
		return errors.New("RegisterClassExW failed")
	}

	hwnd, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(className)),
		0,
		0, 0, 0, 0,
		0,
		0,
		uintptr(instance),
		0,
	)
	if hwnd == 0 {
		if err != syscall.Errno(0) {
			return err
		}
		return errors.New("CreateWindowExW failed")
	}

	// Register for session change notifications
	if r, _, err := procWTSRegisterSessionNotification.Call(hwnd, 0); r == 0 {
		procDestroyWindow.Call(hwnd)
		if err != syscall.Errno(0) {
			return err
		}
		return errors.New("WTSRegisterSessionNotification failed")
	}

	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) == -1 {
			break
		}
		if ret == 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	procWTSUnRegisterSessionNotification.Call(hwnd)
	procDestroyWindow.Call(hwnd)
	procPostQuitMessage.Call(0)
	return nil
}

func lockWorkstation() {
	procLockWorkStation.Call()
}

func getModuleHandle() syscall.Handle {
	handle, _, _ := procGetModuleHandleW.Call(0)
	return syscall.Handle(handle)
}
