//go:build windows

package gui

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/fhatzidakis/claude-code-switcher/internal/fuzzy"
	"github.com/fhatzidakis/claude-code-switcher/internal/projects"
	"github.com/fhatzidakis/claude-code-switcher/internal/terminal"
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
	procGetWindowLongPtrW    = user32.NewProc("GetWindowLongPtrW")
	procCallWindowProcW      = user32.NewProc("CallWindowProcW")
	procGetClientRect        = user32.NewProc("GetClientRect")
	procMoveWindow           = user32.NewProc("MoveWindow")
	procGetSystemMetrics     = user32.NewProc("GetSystemMetrics")
	procSetForegroundWindow  = user32.NewProc("SetForegroundWindow")
	procGetModuleHandleW     = kernel32.NewProc("GetModuleHandleW")
	procCreateFontW          = gdi32.NewProc("CreateFontW")
	procDeleteObject         = gdi32.NewProc("DeleteObject")
	procMessageBoxW          = user32.NewProc("MessageBoxW")
	procInitCommonControlsEx = comctl32.NewProc("InitCommonControlsEx")
	procInvalidateRect       = user32.NewProc("InvalidateRect")
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_VSCROLL          = 0x00200000
	WS_BORDER           = 0x00800000
	WS_TABSTOP          = 0x00010000
	WS_EX_CLIENTEDGE    = 0x00000200
	WS_EX_TOPMOST       = 0x00000008

	ES_AUTOHSCROLL = 0x0080

	LBS_NOTIFY        = 0x0001
	LBS_NOINTEGRALHEIGHT = 0x0100
	LBS_OWNERDRAWFIXED = 0x0010
	LBS_HASSTRINGS    = 0x0040

	LB_ADDSTRING      = 0x0180
	LB_RESETCONTENT   = 0x0184
	LB_GETCURSEL      = 0x0188
	LB_SETCURSEL      = 0x0186
	LB_GETCOUNT       = 0x018B
	LB_GETITEMDATA    = 0x0199
	LB_SETITEMDATA    = 0x019A
	LB_SETITEMHEIGHT  = 0x01A0

	WM_CREATE         = 0x0001
	WM_DESTROY        = 0x0002
	WM_SIZE           = 0x0005
	WM_SETFONT        = 0x0030
	WM_COMMAND        = 0x0111
	WM_KEYDOWN        = 0x0100
	WM_CHAR           = 0x0102
	WM_DRAWITEM       = 0x002B
	WM_MEASUREITEM    = 0x002C

	EN_CHANGE = 0x0300
	LBN_DBLCLK = 2

	VK_RETURN = 0x0D
	VK_ESCAPE = 0x1B
	VK_UP     = 0x26
	VK_DOWN   = 0x28
	VK_TAB    = 0x09

	SW_SHOW = 5

	SM_CXSCREEN = 0
	SM_CYSCREEN = 1

	GWLP_WNDPROC  uintptr = 0xFFFFFFFFFFFFFFFC // -4 as uintptr
	GWLP_USERDATA uintptr = 0xFFFFFFFFFFFFFFEB // -21 as uintptr

	MB_YESNOCANCEL = 0x00000003
	MB_ICONQUESTION = 0x00000020
	IDYES    = 6
	IDNO     = 7
	IDCANCEL = 2

	COLOR_WINDOW     = 5
	COLOR_HIGHLIGHT  = 13
	COLOR_HIGHLIGHTTEXT = 14

	DT_LEFT       = 0x0000
	DT_SINGLELINE = 0x0020
	DT_VCENTER    = 0x0004
	DT_END_ELLIPSIS = 0x8000

	ODT_LISTBOX   = 2
	ODA_DRAWENTIRE = 0x0001
	ODA_SELECT    = 0x0002
	ODS_SELECTED  = 0x0001

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

type INITCOMMONCONTROLSEX struct {
	Size uint32
	ICC  uint32
}

const (
	IDC_EDIT    = 101
	IDC_LISTBOX = 102
	IDC_SORT    = 103
)

var (
	mainHwnd         uintptr
	editHwnd         uintptr
	listHwnd         uintptr
	sortBtnHwnd      uintptr
	hFont            uintptr
	originalEditProc uintptr

	allProjects      []projects.Project
	filteredProjects []projects.Project
	sortByName       bool
	selectedAction   string
)

func utf16PtrFromString(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

// negInt converts a negative int to uintptr for Win32 API calls
func negInt(n int) uintptr {
	return uintptr(int32(n))
}

func Run(projectList []projects.Project) {
	allProjects = projectList
	filteredProjects = projectList

	// Initialize common controls
	var icc INITCOMMONCONTROLSEX
	icc.Size = uint32(unsafe.Sizeof(icc))
	icc.ICC = ICC_LISTVIEW_CLASSES
	procInitCommonControlsEx.Call(uintptr(unsafe.Pointer(&icc)))

	hInstance, _, _ := procGetModuleHandleW.Call(0)

	className := utf16PtrFromString("ClaudeCodeSwitcher")

	wc := WNDCLASSEXW{
		Size:       uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		WndProc:    syscall.NewCallback(wndProc),
		Instance:   syscall.Handle(hInstance),
		ClassName:  className,
		Background: syscall.Handle(COLOR_WINDOW + 1),
	}

	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	// Get screen dimensions for centering
	screenWidth, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	screenHeight, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)

	windowWidth := 600
	windowHeight := 450
	x := (int(screenWidth) - windowWidth) / 2
	y := (int(screenHeight) - windowHeight) / 2

	hwnd, _, _ := procCreateWindowExW.Call(
		WS_EX_TOPMOST,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16PtrFromString("Claude Code Switcher"))),
		WS_OVERLAPPEDWINDOW,
		uintptr(x), uintptr(y),
		uintptr(windowWidth), uintptr(windowHeight),
		0, 0, hInstance, 0,
	)

	mainHwnd = hwnd

	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)
	procSetForegroundWindow.Call(hwnd)
	procSetFocus.Call(editHwnd)

	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_CREATE:
		createControls(hwnd)
		return 0

	case WM_SIZE:
		resizeControls(hwnd)
		return 0

	case WM_COMMAND:
		wmId := wParam & 0xFFFF
		wmEvent := (wParam >> 16) & 0xFFFF

		switch wmId {
		case IDC_EDIT:
			if wmEvent == EN_CHANGE {
				onSearchChanged()
			}
		case IDC_LISTBOX:
			if wmEvent == LBN_DBLCLK {
				onProjectSelected()
			}
		case IDC_SORT:
			toggleSort()
		}
		return 0

	case WM_DRAWITEM:
		dis := (*DRAWITEMSTRUCT)(unsafe.Pointer(lParam))
		if dis.CtlID == IDC_LISTBOX {
			drawListItem(dis)
			return 1
		}
		return 0

	case WM_MEASUREITEM:
		mis := (*MEASUREITEMSTRUCT)(unsafe.Pointer(lParam))
		if mis.CtlID == IDC_LISTBOX {
			mis.ItemHeight = 50 // Height for two lines
			return 1
		}
		return 0

	case WM_DESTROY:
		if hFont != 0 {
			procDeleteObject.Call(hFont)
		}
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func createControls(hwnd uintptr) {
	hInstance, _, _ := procGetModuleHandleW.Call(0)

	// Create font with proper quality settings
	// CreateFontW parameters:
	// Height, Width, Escapement, Orientation, Weight,
	// Italic, Underline, StrikeOut, CharSet,
	// OutputPrecision, ClipPrecision, Quality, PitchAndFamily, FaceName
	const (
		FW_NORMAL           = 400
		DEFAULT_CHARSET     = 1
		OUT_DEFAULT_PRECIS  = 0
		CLIP_DEFAULT_PRECIS = 0
		CLEARTYPE_QUALITY   = 5
		DEFAULT_PITCH       = 0
		FF_DONTCARE         = 0
	)
	hFont, _, _ = procCreateFontW.Call(
		negInt(-18),                 // Height (negative for character height)
		0,                           // Width (0 = default aspect ratio)
		0,                           // Escapement
		0,                           // Orientation
		FW_NORMAL,                   // Weight
		0,                           // Italic
		0,                           // Underline
		0,                           // StrikeOut
		DEFAULT_CHARSET,             // CharSet
		OUT_DEFAULT_PRECIS,          // OutputPrecision
		CLIP_DEFAULT_PRECIS,         // ClipPrecision
		CLEARTYPE_QUALITY,           // Quality - ClearType for smooth fonts
		DEFAULT_PITCH|FF_DONTCARE,   // PitchAndFamily
		uintptr(unsafe.Pointer(utf16PtrFromString("Segoe UI"))),
	)

	// Search edit box
	editHwnd, _, _ = procCreateWindowExW.Call(
		WS_EX_CLIENTEDGE,
		uintptr(unsafe.Pointer(utf16PtrFromString("EDIT"))),
		0,
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|ES_AUTOHSCROLL,
		10, 10, 480, 30,
		hwnd, IDC_EDIT, hInstance, 0,
	)
	procSendMessageW.Call(editHwnd, WM_SETFONT, hFont, 1)

	// Subclass the edit control to handle special keys
	originalEditProc, _, _ = procSetWindowLongPtrW.Call(editHwnd, GWLP_WNDPROC, syscall.NewCallback(editSubclassProc))

	// Sort button
	sortBtnHwnd, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("BUTTON"))),
		uintptr(unsafe.Pointer(utf16PtrFromString("Sort: Recent"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP,
		500, 10, 80, 30,
		hwnd, IDC_SORT, hInstance, 0,
	)
	procSendMessageW.Call(sortBtnHwnd, WM_SETFONT, hFont, 1)

	// Project listbox (owner-draw for custom rendering)
	listHwnd, _, _ = procCreateWindowExW.Call(
		WS_EX_CLIENTEDGE,
		uintptr(unsafe.Pointer(utf16PtrFromString("LISTBOX"))),
		0,
		WS_CHILD|WS_VISIBLE|WS_VSCROLL|WS_TABSTOP|LBS_NOTIFY|LBS_NOINTEGRALHEIGHT|LBS_OWNERDRAWFIXED|LBS_HASSTRINGS,
		10, 50, 565, 350,
		hwnd, IDC_LISTBOX, hInstance, 0,
	)
	procSendMessageW.Call(listHwnd, WM_SETFONT, hFont, 1)

	// Set item height for owner-draw
	procSendMessageW.Call(listHwnd, LB_SETITEMHEIGHT, 0, 50)

	populateList()
}

func editSubclassProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_KEYDOWN:
		switch wParam {
		case VK_DOWN:
			// Move selection down in listbox
			count, _, _ := procSendMessageW.Call(listHwnd, LB_GETCOUNT, 0, 0)
			cur, _, _ := procSendMessageW.Call(listHwnd, LB_GETCURSEL, 0, 0)
			if cur < count-1 {
				procSendMessageW.Call(listHwnd, LB_SETCURSEL, cur+1, 0)
			}
			return 0
		case VK_UP:
			// Move selection up in listbox
			cur, _, _ := procSendMessageW.Call(listHwnd, LB_GETCURSEL, 0, 0)
			if cur > 0 {
				procSendMessageW.Call(listHwnd, LB_SETCURSEL, cur-1, 0)
			}
			return 0
		case VK_RETURN:
			onProjectSelected()
			return 0
		case VK_ESCAPE:
			procDestroyWindow.Call(mainHwnd)
			return 0
		}
	}

	ret, _, _ := procCallWindowProcW.Call(originalEditProc, hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func resizeControls(hwnd uintptr) {
	var rect RECT
	procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))

	width := rect.Right - rect.Left
	height := rect.Bottom - rect.Top

	procMoveWindow.Call(editHwnd, 10, 10, uintptr(width-110), 30, 1)
	procMoveWindow.Call(sortBtnHwnd, uintptr(width-90), 10, 80, 30, 1)
	procMoveWindow.Call(listHwnd, 10, 50, uintptr(width-20), uintptr(height-60), 1)
}

func populateList() {
	procSendMessageW.Call(listHwnd, LB_RESETCONTENT, 0, 0)

	for i, proj := range filteredProjects {
		// Add the project name as the string (for accessibility)
		text := utf16PtrFromString(proj.Name)
		procSendMessageW.Call(listHwnd, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(text)))
		// Store the index in the original slice as item data
		procSendMessageW.Call(listHwnd, LB_SETITEMDATA, uintptr(i), uintptr(i))
	}

	if len(filteredProjects) > 0 {
		procSendMessageW.Call(listHwnd, LB_SETCURSEL, 0, 0)
	}
}

func drawListItem(dis *DRAWITEMSTRUCT) {
	if dis.ItemID == 0xFFFFFFFF {
		return
	}

	idx := int(dis.ItemID)
	if idx >= len(filteredProjects) {
		return
	}

	proj := filteredProjects[idx]

	// Set colors based on selection state
	var bgColor, textColor uint32
	if dis.ItemState&ODS_SELECTED != 0 {
		bgColor = getSysColor(COLOR_HIGHLIGHT)
		textColor = getSysColor(COLOR_HIGHLIGHTTEXT)
	} else {
		bgColor = getSysColor(COLOR_WINDOW)
		textColor = 0x00000000 // Black
	}

	// Fill background
	setBkColor(dis.HDC, bgColor)
	setTextColor(dis.HDC, textColor)

	brush := createSolidBrush(bgColor)
	fillRect(dis.HDC, &dis.RcItem, brush)
	deleteObject(brush)

	// Draw project name (first line, bold-ish)
	nameRect := dis.RcItem
	nameRect.Left += 8
	nameRect.Top += 4
	nameRect.Bottom = nameRect.Top + 20

	nameText := proj.Name
	if proj.InUse {
		nameText = "[ACTIVE] " + nameText
	}
	drawText(dis.HDC, nameText, &nameRect, DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)

	// Draw path and last used (second line, smaller/dimmer)
	if dis.ItemState&ODS_SELECTED == 0 {
		setTextColor(dis.HDC, 0x00666666) // Gray
	}

	infoRect := dis.RcItem
	infoRect.Left += 8
	infoRect.Top += 26
	infoRect.Bottom = infoRect.Top + 18

	lastUsedStr := formatLastUsed(proj.LastUsed)
	infoText := fmt.Sprintf("%s  -  %s", proj.Path, lastUsedStr)
	drawText(dis.HDC, infoText, &infoRect, DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
}

func formatLastUsed(t time.Time) string {
	if t.IsZero() {
		return "Never"
	}

	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "Just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "Yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}

func onSearchChanged() {
	// Get search text
	length, _, _ := procGetWindowTextLengthW.Call(editHwnd)
	if length == 0 {
		filteredProjects = allProjects
		populateList()
		return
	}

	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(editHwnd, uintptr(unsafe.Pointer(&buf[0])), length+1)
	searchText := syscall.UTF16ToString(buf)

	// Fuzzy filter
	var names []string
	for _, p := range allProjects {
		names = append(names, p.Name+" "+p.Path)
	}

	scored := fuzzy.FilterAndScore(searchText, names)

	filteredProjects = nil
	for _, item := range scored {
		filteredProjects = append(filteredProjects, allProjects[item.Index])
	}

	populateList()
}

func toggleSort() {
	sortByName = !sortByName

	if sortByName {
		procSetWindowTextW.Call(sortBtnHwnd, uintptr(unsafe.Pointer(utf16PtrFromString("Sort: Name"))))
		projects.SortByName(allProjects)
		projects.SortByName(filteredProjects)
	} else {
		procSetWindowTextW.Call(sortBtnHwnd, uintptr(unsafe.Pointer(utf16PtrFromString("Sort: Recent"))))
		projects.SortByLastUsed(allProjects)
		projects.SortByLastUsed(filteredProjects)
	}

	populateList()
}

func onProjectSelected() {
	sel, _, _ := procSendMessageW.Call(listHwnd, LB_GETCURSEL, 0, 0)
	if sel == 0xFFFFFFFF || int(sel) >= len(filteredProjects) {
		return
	}

	proj := filteredProjects[sel]

	// Check if project is in use
	if proj.InUse {
		result := showMessageBox(
			mainHwnd,
			fmt.Sprintf("Claude is already running in '%s'.\n\nWhat would you like to do?", proj.Name),
			"Project In Use",
			MB_YESNOCANCEL|MB_ICONQUESTION,
		)

		switch result {
		case IDYES:
			// Focus existing - TODO: implement window focusing
			procDestroyWindow.Call(mainHwnd)
			return
		case IDNO:
			// Open new - continue below
		case IDCANCEL:
			return
		}
	}

	// Open in Windows Terminal
	err := terminal.OpenInWindowsTerminal(proj.Path)
	if err != nil {
		showMessageBox(mainHwnd, "Failed to open Windows Terminal: "+err.Error(), "Error", 0)
		return
	}

	procDestroyWindow.Call(mainHwnd)
}

// Win32 helper functions
func getSysColor(index int) uint32 {
	procGetSysColor := user32.NewProc("GetSysColor")
	ret, _, _ := procGetSysColor.Call(uintptr(index))
	return uint32(ret)
}

func setBkColor(hdc syscall.Handle, color uint32) {
	procSetBkColor := gdi32.NewProc("SetBkColor")
	procSetBkColor.Call(uintptr(hdc), uintptr(color))
}

func setTextColor(hdc syscall.Handle, color uint32) {
	procSetTextColor := gdi32.NewProc("SetTextColor")
	procSetTextColor.Call(uintptr(hdc), uintptr(color))
}

func createSolidBrush(color uint32) syscall.Handle {
	procCreateSolidBrush := gdi32.NewProc("CreateSolidBrush")
	ret, _, _ := procCreateSolidBrush.Call(uintptr(color))
	return syscall.Handle(ret)
}

func fillRect(hdc syscall.Handle, rect *RECT, brush syscall.Handle) {
	procFillRect := user32.NewProc("FillRect")
	procFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(rect)), uintptr(brush))
}

func deleteObject(obj syscall.Handle) {
	procDeleteObject.Call(uintptr(obj))
}

func drawText(hdc syscall.Handle, text string, rect *RECT, format uint32) {
	procDrawTextW := user32.NewProc("DrawTextW")
	textPtr := utf16PtrFromString(text)
	procDrawTextW.Call(uintptr(hdc), uintptr(unsafe.Pointer(textPtr)), uintptr(len(text)), uintptr(unsafe.Pointer(rect)), uintptr(format))
}

func showMessageBox(hwnd uintptr, text, caption string, flags uint32) int {
	ret, _, _ := procMessageBoxW.Call(
		hwnd,
		uintptr(unsafe.Pointer(utf16PtrFromString(text))),
		uintptr(unsafe.Pointer(utf16PtrFromString(caption))),
		uintptr(flags),
	)
	return int(ret)
}
