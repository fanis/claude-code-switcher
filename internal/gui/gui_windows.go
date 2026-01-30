//go:build windows

package gui

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/fanis/claude-code-switcher/internal/fuzzy"
	"github.com/fanis/claude-code-switcher/internal/projects"
	"github.com/fanis/claude-code-switcher/internal/terminal"
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
	procGetDpiForWindow      = user32.NewProc("GetDpiForWindow")
	procPostMessageW         = user32.NewProc("PostMessageW")
	procEnableWindow         = user32.NewProc("EnableWindow")
)

const (
	WM_CLOSE = 0x0010
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
	WM_ACTIVATE       = 0x0006

	WA_INACTIVE = 0

	EN_CHANGE = 0x0300
	LBN_DBLCLK = 2

	VK_RETURN    = 0x0D
	VK_ESCAPE    = 0x1B
	VK_UP        = 0x26
	VK_DOWN      = 0x28
	VK_TAB       = 0x09
	VK_BACK      = 0x08
	VK_CONTROL   = 0x11

	EM_SETSEL    = 0x00B1
	EM_GETSEL    = 0x00B0
	EM_REPLACESEL = 0x00C2

	SW_SHOW = 5

	SM_CXSCREEN = 0
	SM_CYSCREEN = 1

	GWLP_WNDPROC  uintptr = 0xFFFFFFFFFFFFFFFC // -4 as uintptr
	GWLP_USERDATA uintptr = 0xFFFFFFFFFFFFFFEB // -21 as uintptr

	MB_YESNOCANCEL  = 0x00000003
	MB_ICONERROR    = 0x00000010
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
	IDC_ABOUT   = 104
)

const (
	VK_F1 = 0x70
)

var (
	mainHwnd         uintptr
	editHwnd         uintptr
	listHwnd         uintptr
	sortBtnHwnd      uintptr
	aboutBtnHwnd     uintptr
	hFont            uintptr
	originalEditProc uintptr
	currentDPI       uint32 = 96 // Default DPI

	allProjects      []projects.Project
	filteredProjects []projects.Project
	sortByName       bool
	selectedAction   string
	showingDialog    bool // Prevent close on focus loss while showing dialog
	appVersion       string
)

func utf16PtrFromString(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

// negInt converts a negative int to uintptr for Win32 API calls
func negInt(n int) uintptr {
	return uintptr(int32(n))
}

func Run(projectList []projects.Project, version string) {
	allProjects = projectList
	filteredProjects = projectList
	appVersion = version

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
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16PtrFromString("Claude Code Switcher"))),
		WS_OVERLAPPEDWINDOW,
		uintptr(x), uintptr(y),
		uintptr(windowWidth), uintptr(windowHeight),
		0, 0, hInstance, 0,
	)

	mainHwnd = hwnd
	terminal.SetParentHwnd(hwnd)

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
		case IDC_ABOUT:
			showingDialog = true
			showAboutDialog()
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
			// Scale item height based on DPI (base 40 at 96 DPI)
			baseHeight := uint32(40)
			mis.ItemHeight = (baseHeight * currentDPI) / 96
			return 1
		}
		return 0

	case WM_ACTIVATE:
		// Close window when it loses focus (launcher-style behavior)
		// But not if we're showing a dialog
		if wParam&0xFFFF == WA_INACTIVE && !showingDialog {
			// Post WM_CLOSE to close gracefully without re-entrancy issues
			procPostMessageW.Call(hwnd, WM_CLOSE, 0, 0)
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

	// Get DPI for proper font scaling
	currentDPI = getDPI(hwnd)
	dpi := currentDPI

	// Base font size at 96 DPI, scale for current DPI
	baseFontSize := 14
	scaledFontSize := (baseFontSize * int(dpi)) / 96

	// Create font with proper quality settings
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
		negInt(-scaledFontSize),     // Height (negative for character height)
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
		10, 10, 490, 30,
		hwnd, IDC_EDIT, hInstance, 0,
	)
	procSendMessageW.Call(editHwnd, WM_SETFONT, hFont, 1)

	// Subclass the edit control to handle special keys
	originalEditProc, _, _ = procSetWindowLongPtrW.Call(editHwnd, GWLP_WNDPROC, syscall.NewCallback(editSubclassProc))

	// Sort button
	sortBtnHwnd, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("BUTTON"))),
		uintptr(unsafe.Pointer(utf16PtrFromString("By: Recent"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP,
		500, 10, 90, 30,
		hwnd, IDC_SORT, hInstance, 0,
	)
	procSendMessageW.Call(sortBtnHwnd, WM_SETFONT, hFont, 1)

	// About button (small "?" button)
	aboutBtnHwnd, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("BUTTON"))),
		uintptr(unsafe.Pointer(utf16PtrFromString("?"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP,
		592, 10, 26, 30,
		hwnd, IDC_ABOUT, hInstance, 0,
	)
	procSendMessageW.Call(aboutBtnHwnd, WM_SETFONT, hFont, 1)

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

	// Set item height for owner-draw (scale based on DPI)
	itemHeight := (40 * dpi) / 96
	procSendMessageW.Call(listHwnd, LB_SETITEMHEIGHT, 0, uintptr(itemHeight))

	populateList()
}

func editSubclassProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_CHAR:
		// Handle Ctrl+Backspace (comes as 0x7F character)
		if wParam == 0x7F {
			deleteWordBackward(hwnd)
			return 0
		}
	case WM_KEYDOWN:
		switch wParam {
		case VK_TAB:
			// Tab toggles sort mode
			toggleSort()
			return 0
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
		case VK_F1:
			showAboutDialog()
			return 0
		}
	}

	ret, _, _ := procCallWindowProcW.Call(originalEditProc, hwnd, uintptr(msg), wParam, lParam)
	return ret
}

// deleteWordBackward deletes the word before the cursor
func deleteWordBackward(hwnd uintptr) {
	// Get current text
	length, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if length == 0 {
		return
	}

	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), length+1)
	text := syscall.UTF16ToString(buf)

	// Get cursor position
	var start, end uint32
	procSendMessageW.Call(hwnd, EM_GETSEL, uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&end)))

	if start == 0 {
		return
	}

	// Find word boundary (skip spaces, then skip non-spaces)
	pos := int(start)
	for pos > 0 && (text[pos-1] == ' ' || text[pos-1] == '\t') {
		pos--
	}
	for pos > 0 && text[pos-1] != ' ' && text[pos-1] != '\t' {
		pos--
	}

	// Select from word start to cursor and delete
	procSendMessageW.Call(hwnd, EM_SETSEL, uintptr(pos), uintptr(start))
	procSendMessageW.Call(hwnd, EM_REPLACESEL, 0, uintptr(unsafe.Pointer(utf16PtrFromString(""))))
}

// getDPI returns the DPI for the window, with fallback for older Windows
func getDPI(hwnd uintptr) uint32 {
	// Try GetDpiForWindow (Windows 10 1607+)
	if procGetDpiForWindow.Find() == nil {
		dpi, _, _ := procGetDpiForWindow.Call(hwnd)
		if dpi > 0 {
			return uint32(dpi)
		}
	}
	// Fallback to 96 (standard DPI)
	return 96
}

func resizeControls(hwnd uintptr) {
	var rect RECT
	procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))

	width := rect.Right - rect.Left
	height := rect.Bottom - rect.Top

	sortBtnWidth := int32(90)
	aboutBtnWidth := int32(26)
	margin := int32(10)
	gap := int32(6)

	totalBtnWidth := sortBtnWidth + gap + aboutBtnWidth
	editWidth := width - totalBtnWidth - margin*2 - gap

	procMoveWindow.Call(editHwnd, uintptr(margin), uintptr(margin), uintptr(editWidth), 30, 1)
	procMoveWindow.Call(sortBtnHwnd, uintptr(margin+editWidth+gap), uintptr(margin), uintptr(sortBtnWidth), 30, 1)
	procMoveWindow.Call(aboutBtnHwnd, uintptr(width-aboutBtnWidth-margin), uintptr(margin), uintptr(aboutBtnWidth), 30, 1)
	procMoveWindow.Call(listHwnd, uintptr(margin), 50, uintptr(width-margin*2), uintptr(height-60), 1)
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

	// DPI-scaled values
	scale := func(base int32) int32 {
		return (base * int32(currentDPI)) / 96
	}

	// Modern color scheme (colors in BGR format for Windows)
	var bgColor, textColor, secondaryColor uint32
	if dis.ItemState&ODS_SELECTED != 0 {
		bgColor = 0x00CC7A00       // Nice blue (#007ACC in RGB)
		textColor = 0x00FFFFFF     // White
		secondaryColor = 0x00E0E0E0 // Light gray
	} else if !proj.PathExists {
		bgColor = 0x00F0F0F0       // Light gray background
		textColor = 0x00808080     // Gray text
		secondaryColor = 0x00A0A0A0 // Lighter gray
	} else {
		bgColor = 0x00FFFFFF       // White
		textColor = 0x00202020     // Near black
		secondaryColor = 0x00808080 // Gray
	}

	// Fill background
	setBkColor(dis.HDC, bgColor)
	setTextColor(dis.HDC, textColor)

	brush := createSolidBrush(bgColor)
	fillRect(dis.HDC, &dis.RcItem, brush)
	deleteObject(brush)

	// Draw project name (first line)
	nameRect := dis.RcItem
	nameRect.Left += scale(8)
	nameRect.Top += scale(4)
	nameRect.Bottom = nameRect.Top + scale(18)

	nameText := proj.Name
	if !proj.PathExists {
		nameText = "[NOT FOUND] " + nameText
	} else if proj.InUse {
		nameText = "[ACTIVE] " + nameText
	}
	drawText(dis.HDC, nameText, &nameRect, DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)

	// Draw path and last used (second line)
	setTextColor(dis.HDC, secondaryColor)

	infoRect := dis.RcItem
	infoRect.Left += scale(8)
	infoRect.Top += scale(22)
	infoRect.Bottom = infoRect.Top + scale(16)

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

var (
	aboutDlgHwnd uintptr
	aboutLinkHwnd uintptr
)

const (
	WM_NOTIFY     = 0x004E
	NM_CLICK      = 0xFFFFFFFE // -2
	NM_RETURN     = 0xFFFFFFFD // -3
	WS_POPUP      = 0x80000000
	WS_CAPTION    = 0x00C00000
	WS_SYSMENU    = 0x00080000
	DS_MODALFRAME = 0x80
	DS_CENTER     = 0x0800
	SS_CENTER     = 0x0001

	IDC_ABOUT_LINK = 201
	IDC_ABOUT_OK   = 202
)

func showAboutDialog() {
	showingDialog = true
	defer func() {
		showingDialog = false
		procSetFocus.Call(editHwnd)
	}()

	hInstance, _, _ := procGetModuleHandleW.Call(0)

	// Register dialog class
	className := utf16PtrFromString("ClaudeAboutDialog")

	wc := WNDCLASSEXW{
		Size:       uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		WndProc:    syscall.NewCallback(aboutDlgProc),
		Instance:   syscall.Handle(hInstance),
		ClassName:  className,
		Background: syscall.Handle(COLOR_WINDOW + 1),
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	// Get main window position for centering
	var mainRect RECT
	procGetClientRect.Call(mainHwnd, uintptr(unsafe.Pointer(&mainRect)))

	// Get screen position of main window
	var pt POINT
	pt.X = mainRect.Left
	pt.Y = mainRect.Top
	clientToScreen := user32.NewProc("ClientToScreen")
	clientToScreen.Call(mainHwnd, uintptr(unsafe.Pointer(&pt)))

	dlgWidth := 300
	dlgHeight := 260
	dlgX := int(pt.X) + (int(mainRect.Right-mainRect.Left)-dlgWidth)/2
	dlgY := int(pt.Y) + (int(mainRect.Bottom-mainRect.Top)-dlgHeight)/2

	// Disable main window for modal behavior
	procEnableWindow.Call(mainHwnd, 0)

	// Create dialog window
	aboutDlgHwnd, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16PtrFromString("About"))),
		WS_POPUP|WS_CAPTION|WS_SYSMENU,
		uintptr(dlgX), uintptr(dlgY),
		uintptr(dlgWidth), uintptr(dlgHeight),
		mainHwnd, 0, hInstance, 0,
	)

	// Create controls
	createAboutControls(aboutDlgHwnd, hInstance)

	procShowWindow.Call(aboutDlgHwnd, SW_SHOW)
	procUpdateWindow.Call(aboutDlgHwnd)

	// Modal message loop
	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 || aboutDlgHwnd == 0 {
			break
		}
		// Handle ESC key to close dialog
		if msg.Message == WM_KEYDOWN && msg.WParam == VK_ESCAPE {
			procDestroyWindow.Call(aboutDlgHwnd)
			aboutDlgHwnd = 0
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	// Re-enable main window
	procEnableWindow.Call(mainHwnd, 1)
}

func createAboutControls(hwnd uintptr, hInstance uintptr) {
	// Title
	titleText := "Claude Code Switcher"
	if appVersion != "" {
		titleText += " v" + appVersion
	}
	titleHwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("STATIC"))),
		uintptr(unsafe.Pointer(utf16PtrFromString(titleText))),
		WS_CHILD|WS_VISIBLE|SS_CENTER,
		10, 10, 280, 20,
		hwnd, 0, hInstance, 0,
	)
	procSendMessageW.Call(titleHwnd, WM_SETFONT, hFont, 1)

	// Author
	authorHwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("STATIC"))),
		uintptr(unsafe.Pointer(utf16PtrFromString("by Fanis Hatzidakis"))),
		WS_CHILD|WS_VISIBLE|SS_CENTER,
		10, 32, 280, 18,
		hwnd, 0, hInstance, 0,
	)
	procSendMessageW.Call(authorHwnd, WM_SETFONT, hFont, 1)

	// Shortcuts
	shortcutsText := "Keyboard shortcuts:\n" +
		"  Tab - Toggle sort\n" +
		"  Arrows - Navigate\n" +
		"  Enter - Open project\n" +
		"  Escape - Close\n" +
		"  F1 - About"
	shortcutsHwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("STATIC"))),
		uintptr(unsafe.Pointer(utf16PtrFromString(shortcutsText))),
		WS_CHILD|WS_VISIBLE,
		15, 58, 270, 100,
		hwnd, 0, hInstance, 0,
	)
	procSendMessageW.Call(shortcutsHwnd, WM_SETFONT, hFont, 1)

	// GitHub button (more reliable than SysLink)
	githubHwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("BUTTON"))),
		uintptr(unsafe.Pointer(utf16PtrFromString("Open GitHub"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP,
		15, 185, 100, 26,
		hwnd, IDC_ABOUT_LINK, hInstance, 0,
	)
	procSendMessageW.Call(githubHwnd, WM_SETFONT, hFont, 1)

	// OK button
	okHwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("BUTTON"))),
		uintptr(unsafe.Pointer(utf16PtrFromString("OK"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP,
		185, 185, 80, 26,
		hwnd, IDC_ABOUT_OK, hInstance, 0,
	)
	procSendMessageW.Call(okHwnd, WM_SETFONT, hFont, 1)
}

func aboutDlgProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_KEYDOWN:
		if wParam == VK_ESCAPE {
			procDestroyWindow.Call(hwnd)
			aboutDlgHwnd = 0
			return 0
		}
	case WM_COMMAND:
		wmId := wParam & 0xFFFF
		if wmId == IDC_ABOUT_OK {
			procDestroyWindow.Call(hwnd)
			aboutDlgHwnd = 0
			return 0
		}
		if wmId == IDC_ABOUT_LINK {
			openURL("https://github.com/fanis/claude-code-switcher")
			return 0
		}
	case WM_CLOSE:
		procDestroyWindow.Call(hwnd)
		aboutDlgHwnd = 0
		return 0
	case WM_DESTROY:
		aboutDlgHwnd = 0
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

type NMHDR struct {
	HwndFrom uintptr
	IdFrom   uintptr
	Code     uint32
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

func toggleSort() {
	sortByName = !sortByName

	if sortByName {
		procSetWindowTextW.Call(sortBtnHwnd, uintptr(unsafe.Pointer(utf16PtrFromString("By: Name"))))
		projects.SortByName(allProjects)
		projects.SortByName(filteredProjects)
	} else {
		procSetWindowTextW.Call(sortBtnHwnd, uintptr(unsafe.Pointer(utf16PtrFromString("By: Recent"))))
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

	proj := &filteredProjects[sel]

	// Check if project path exists
	if !proj.PathExists {
		showMessageBox(mainHwnd,
			"The project directory no longer exists:\n\n"+proj.Path+"\n\n"+
				"It may have been moved or deleted.",
			"Project Not Found", MB_ICONERROR)
		return
	}

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

	// Show opening indication
	procSetWindowTextW.Call(mainHwnd, uintptr(unsafe.Pointer(utf16PtrFromString(fmt.Sprintf("Opening %s...", proj.Name)))))
	procEnableWindow.Call(editHwnd, 0)
	procEnableWindow.Call(listHwnd, 0)
	procEnableWindow.Call(sortBtnHwnd, 0)

	// Open in Windows Terminal
	// Set flag to prevent close on focus loss during terminal dialogs
	showingDialog = true
	err := terminal.OpenInWindowsTerminal(proj.Path)
	showingDialog = false

	if err != nil {
		// Restore UI on failure
		procSetWindowTextW.Call(mainHwnd, uintptr(unsafe.Pointer(utf16PtrFromString("Claude Code Switcher"))))
		procEnableWindow.Call(editHwnd, 1)
		procEnableWindow.Call(listHwnd, 1)
		procEnableWindow.Call(sortBtnHwnd, 1)
		showMessageBox(mainHwnd, "Failed to open terminal: "+err.Error(), "Error", 0)
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
	showingDialog = true
	defer func() {
		showingDialog = false
		procSetFocus.Call(editHwnd)
	}()

	ret, _, _ := procMessageBoxW.Call(
		hwnd,
		uintptr(unsafe.Pointer(utf16PtrFromString(text))),
		uintptr(unsafe.Pointer(utf16PtrFromString(caption))),
		uintptr(flags),
	)
	return int(ret)
}
