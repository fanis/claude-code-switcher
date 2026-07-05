// Copyright (c) 2025 Fanis Hatzidakis
// Licensed under PolyForm Internal Use License 1.0.0 - see LICENSE.md

//go:build windows

// Package win32 holds small Win32 helpers shared by the other packages,
// so string conversion and message boxes have a single implementation.
package win32

import (
	"syscall"
	"unsafe"
)

var (
	user32          = syscall.NewLazyDLL("user32.dll")
	procMessageBoxW = user32.NewProc("MessageBoxW")
)

// MessageBox flags and return values.
const (
	MB_OK              = 0x00000000
	MB_YESNO           = 0x00000004
	MB_ICONERROR       = 0x00000010
	MB_ICONQUESTION    = 0x00000020
	MB_ICONINFORMATION = 0x00000040
	IDYES              = 6
)

// UTF16Ptr converts a Go string to a null-terminated UTF-16 pointer for
// Win32 calls. Unlike the deprecated syscall.StringToUTF16Ptr, the
// underlying conversion errors on embedded nulls instead of silently
// truncating; on error a pointer to an empty string is returned.
func UTF16Ptr(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		return &[]uint16{0}[0]
	}
	return p
}

// MessageBox shows a message box owned by hwnd (0 for none) and returns
// the pressed button id (e.g. IDYES).
func MessageBox(hwnd uintptr, text, caption string, flags uint32) int {
	ret, _, _ := procMessageBoxW.Call(
		hwnd,
		uintptr(unsafe.Pointer(UTF16Ptr(text))),
		uintptr(unsafe.Pointer(UTF16Ptr(caption))),
		uintptr(flags),
	)
	return int(ret)
}
