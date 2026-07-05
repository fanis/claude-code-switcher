// Copyright (c) 2025 Fanis Hatzidakis
// Licensed under PolyForm Internal Use License 1.0.0 - see LICENSE.md

//go:build windows

// Win32 declarations (DLL procs, constants, structs) and thin helper
// wrappers shared by the rest of the gui package.

package gui

import (
	"syscall"
	"unsafe"

	"github.com/fanis/claude-code-switcher/internal/win32"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	comctl32 = syscall.NewLazyDLL("comctl32.dll")

	procRegisterClassExW     = user32.NewProc("RegisterClassExW")
	procCreateWindowExW      = user32.NewProc("CreateWindowExW")
	procDefWindowProcW       = user32.NewProc("DefWindowProcW")
	procGetMessageW          = user32.NewProc("GetMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessageW     = user32.NewProc("DispatchMessageW")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procDestroyWindow        = user32.NewProc("DestroyWindow")
	procShowWindow           = user32.NewProc("ShowWindow")
	procUpdateWindow         = user32.NewProc("UpdateWindow")
	procSetFocus             = user32.NewProc("SetFocus")
	procSendMessageW         = user32.NewProc("SendMessageW")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procSetWindowTextW       = user32.NewProc("SetWindowTextW")
	procGetDlgItem           = user32.NewProc("GetDlgItem")
	procSetWindowLongPtrW    = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProcW      = user32.NewProc("CallWindowProcW")
	procGetClientRect        = user32.NewProc("GetClientRect")
	procMoveWindow           = user32.NewProc("MoveWindow")
	procGetSystemMetrics     = user32.NewProc("GetSystemMetrics")
	procSetForegroundWindow  = user32.NewProc("SetForegroundWindow")
	procInvalidateRect       = user32.NewProc("InvalidateRect")
	procGetDpiForWindow      = user32.NewProc("GetDpiForWindow")
	procPostMessageW         = user32.NewProc("PostMessageW")
	procEnableWindow         = user32.NewProc("EnableWindow")
	procIsDialogMessageW     = user32.NewProc("IsDialogMessageW")
	procClientToScreen       = user32.NewProc("ClientToScreen")
	procFillRect             = user32.NewProc("FillRect")
	procDrawTextW            = user32.NewProc("DrawTextW")
	procGetModuleHandleW     = kernel32.NewProc("GetModuleHandleW")
	procCreateFontW          = gdi32.NewProc("CreateFontW")
	procDeleteObject         = gdi32.NewProc("DeleteObject")
	procSetBkColor           = gdi32.NewProc("SetBkColor")
	procSetTextColor         = gdi32.NewProc("SetTextColor")
	procCreateSolidBrush     = gdi32.NewProc("CreateSolidBrush")
	procInitCommonControlsEx = comctl32.NewProc("InitCommonControlsEx")
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_VSCROLL          = 0x00200000
	WS_BORDER           = 0x00800000
	WS_TABSTOP          = 0x00010000
	WS_POPUP            = 0x80000000
	WS_CAPTION          = 0x00C00000
	WS_SYSMENU          = 0x00080000
	WS_EX_CLIENTEDGE    = 0x00000200

	ES_AUTOHSCROLL = 0x0080

	LBS_NOTIFY           = 0x0001
	LBS_NOINTEGRALHEIGHT = 0x0100
	LBS_OWNERDRAWFIXED   = 0x0010
	LBS_HASSTRINGS       = 0x0040

	LB_ADDSTRING     = 0x0180
	LB_RESETCONTENT  = 0x0184
	LB_GETCURSEL     = 0x0188
	LB_SETCURSEL     = 0x0186
	LB_GETCOUNT      = 0x018B
	LB_SETITEMHEIGHT = 0x01A0

	WM_CREATE         = 0x0001
	WM_DESTROY        = 0x0002
	WM_SIZE           = 0x0005
	WM_ACTIVATE       = 0x0006
	WM_CLOSE          = 0x0010
	WM_GETMINMAXINFO  = 0x0024
	WM_DRAWITEM       = 0x002B
	WM_MEASUREITEM    = 0x002C
	WM_SETFONT        = 0x0030
	WM_KEYDOWN        = 0x0100
	WM_CHAR           = 0x0102
	WM_COMMAND        = 0x0111
	WM_CTLCOLORSTATIC = 0x0138
	WM_APP            = 0x8000
	WM_APP_UPDATE     = WM_APP + 1

	WA_INACTIVE = 0

	EN_CHANGE  = 0x0300
	LBN_DBLCLK = 2

	VK_TAB    = 0x09
	VK_RETURN = 0x0D
	VK_ESCAPE = 0x1B
	VK_UP     = 0x26
	VK_DOWN   = 0x28
	VK_F1     = 0x70

	EM_GETSEL     = 0x00B0
	EM_SETSEL     = 0x00B1
	EM_REPLACESEL = 0x00C2

	SW_HIDE = 0
	SW_SHOW = 5

	SM_CXSCREEN = 0
	SM_CYSCREEN = 1

	// -4 in two's complement, valid on both 32- and 64-bit builds
	GWLP_WNDPROC = ^uintptr(3)

	BM_GETCHECK     = 0x00F0
	BM_SETCHECK     = 0x00F1
	BST_CHECKED     = 1
	BS_AUTOCHECKBOX = 0x0003
	SS_ETCHEDHORZ   = 0x0010
	SS_CENTER       = 0x0001

	CBS_DROPDOWNLIST = 0x0003
	CBS_HASSTRINGS   = 0x0200
	CB_ADDSTRING     = 0x0143
	CB_SETCURSEL     = 0x014E
	CB_GETCURSEL     = 0x0147
	CBN_SELCHANGE    = 1

	COLOR_WINDOW = 5

	DT_LEFT         = 0x0000
	DT_RIGHT        = 0x0002
	DT_SINGLELINE   = 0x0020
	DT_END_ELLIPSIS = 0x8000

	ODS_SELECTED = 0x0001

	ICC_LISTVIEW_CLASSES = 0x00000001
)

type WNDCLASSEXW struct {
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
	Hwnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type POINT struct {
	X, Y int32
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

type DRAWITEMSTRUCT struct {
	CtlType    uint32
	CtlID      uint32
	ItemID     uint32
	ItemAction uint32
	ItemState  uint32
	HwndItem   syscall.Handle
	HDC        syscall.Handle
	RcItem     RECT
	ItemData   uintptr
}

type MEASUREITEMSTRUCT struct {
	CtlType    uint32
	CtlID      uint32
	ItemID     uint32
	ItemWidth  uint32
	ItemHeight uint32
	ItemData   uintptr
}

type MINMAXINFO struct {
	Reserved     POINT
	MaxSize      POINT
	MaxPosition  POINT
	MinTrackSize POINT
	MaxTrackSize POINT
}

type INITCOMMONCONTROLSEX struct {
	Size uint32
	ICC  uint32
}

func utf16PtrFromString(s string) *uint16 {
	return win32.UTF16Ptr(s)
}

// negInt converts a negative int to uintptr for Win32 API calls
func negInt(n int) uintptr {
	return uintptr(int32(n))
}

// dpiScale scales a base value defined at 96 DPI to the current DPI.
func dpiScale(base int32) int32 {
	return (base * int32(currentDPI)) / 96
}

func getDlgItem(hwnd uintptr, id uintptr) uintptr {
	ret, _, _ := procGetDlgItem.Call(hwnd, id)
	return ret
}

func getWindowText(hwnd uintptr) string {
	length, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), length+1)
	return syscall.UTF16ToString(buf)
}

func setBkColor(hdc syscall.Handle, color uint32) {
	procSetBkColor.Call(uintptr(hdc), uintptr(color))
}

func setTextColor(hdc syscall.Handle, color uint32) {
	procSetTextColor.Call(uintptr(hdc), uintptr(color))
}

func createSolidBrush(color uint32) syscall.Handle {
	ret, _, _ := procCreateSolidBrush.Call(uintptr(color))
	return syscall.Handle(ret)
}

func fillRect(hdc syscall.Handle, rect *RECT, brush syscall.Handle) {
	procFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(rect)), uintptr(brush))
}

func deleteObject(obj uintptr) {
	if obj != 0 {
		procDeleteObject.Call(obj)
	}
}

func drawText(hdc syscall.Handle, text string, rect *RECT, format uint32) {
	textPtr := utf16PtrFromString(text)
	procDrawTextW.Call(uintptr(hdc), uintptr(unsafe.Pointer(textPtr)), negInt(-1), uintptr(unsafe.Pointer(rect)), uintptr(format))
}

func showMessageBox(hwnd uintptr, text, caption string, flags uint32) int {
	showingDialog.Store(true)
	defer func() {
		showingDialog.Store(false)
		procSetFocus.Call(editHwnd)
	}()

	return win32.MessageBox(hwnd, text, caption, flags)
}

func openURL(url string) {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shellExecute := shell32.NewProc("ShellExecuteW")

	shellExecute.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("open"))),
		uintptr(unsafe.Pointer(utf16PtrFromString(url))),
		0,
		0,
		1,
	)
}
