// Copyright (c) 2025 Fanis Hatzidakis
// Licensed under PolyForm Internal Use License 1.0.0 - see LICENSE.md

//go:build windows

// Main window controls: creation, layout, keyboard handling, search
// filtering, sorting, and project launching.

package gui

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/fanis/claude-code-switcher/internal/fuzzy"
	"github.com/fanis/claude-code-switcher/internal/projects"
	"github.com/fanis/claude-code-switcher/internal/terminal"
	"github.com/fanis/claude-code-switcher/internal/win32"
)

func createControls(hwnd uintptr) {
	hInstance, _, _ := procGetModuleHandleW.Call(0)

	// Get DPI for proper font scaling
	currentDPI = getDPI(hwnd)

	// Create font with proper quality settings
	const (
		FW_NORMAL           = 400
		FW_BOLD             = 700
		DEFAULT_CHARSET     = 1
		OUT_DEFAULT_PRECIS  = 0
		CLIP_DEFAULT_PRECIS = 0
		CLEARTYPE_QUALITY   = 5
		DEFAULT_PITCH       = 0
		FF_DONTCARE         = 0
	)
	makeFont := func(size int32, weight uintptr) uintptr {
		f, _, _ := procCreateFontW.Call(
			negInt(int(-dpiScale(size))),
			0, 0, 0, weight, 0, 0, 0,
			DEFAULT_CHARSET, OUT_DEFAULT_PRECIS, CLIP_DEFAULT_PRECIS,
			CLEARTYPE_QUALITY, DEFAULT_PITCH|FF_DONTCARE,
			uintptr(unsafe.Pointer(utf16PtrFromString("Segoe UI"))),
		)
		return f
	}

	hFont = makeFont(14, FW_NORMAL)
	hFontBold = makeFont(14, FW_BOLD)
	hFontSmall = makeFont(12, FW_NORMAL)

	// Brushes: white doubles as dialog/static background and the normal
	// list item background; the others are the list item variants.
	// Created once here instead of per list item draw.
	hWhiteBrush = uintptr(createSolidBrush(0x00FFFFFF))
	hSelectedBrush = uintptr(createSolidBrush(0x00CC7A00))
	hMissingBrush = uintptr(createSolidBrush(0x00F0F0F0))

	// Search edit box (positions/sizes are set for real in resizeControls)
	editHwnd, _, _ = procCreateWindowExW.Call(
		WS_EX_CLIENTEDGE,
		uintptr(unsafe.Pointer(utf16PtrFromString("EDIT"))),
		0,
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|ES_AUTOHSCROLL,
		0, 0, 0, 0,
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
		0, 0, 0, 0,
		hwnd, IDC_SORT, hInstance, 0,
	)
	procSendMessageW.Call(sortBtnHwnd, WM_SETFONT, hFont, 1)

	// Settings button (gear icon)
	settingsBtnHwnd, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("BUTTON"))),
		uintptr(unsafe.Pointer(utf16PtrFromString("⚙"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP,
		0, 0, 0, 0,
		hwnd, IDC_SETTINGS, hInstance, 0,
	)
	procSendMessageW.Call(settingsBtnHwnd, WM_SETFONT, hFont, 1)

	// Project listbox (owner-draw for custom rendering)
	listHwnd, _, _ = procCreateWindowExW.Call(
		WS_EX_CLIENTEDGE,
		uintptr(unsafe.Pointer(utf16PtrFromString("LISTBOX"))),
		0,
		WS_CHILD|WS_VISIBLE|WS_VSCROLL|WS_TABSTOP|LBS_NOTIFY|LBS_NOINTEGRALHEIGHT|LBS_OWNERDRAWFIXED|LBS_HASSTRINGS,
		0, 0, 0, 0,
		hwnd, IDC_LISTBOX, hInstance, 0,
	)
	procSendMessageW.Call(listHwnd, WM_SETFONT, hFont, 1)

	// Set item height for owner-draw (scale based on DPI)
	procSendMessageW.Call(listHwnd, LB_SETITEMHEIGHT, 0, uintptr(dpiScale(40)))

	populateList()

	// The window was created at 96-DPI dimensions before the DPI was
	// known; resize and re-center it to match the actual display DPI.
	if currentDPI != 96 {
		w := dpiScale(600)
		h := dpiScale(450)
		screenWidth, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
		screenHeight, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)
		x := (int32(screenWidth) - w) / 2
		y := (int32(screenHeight) - h) / 2
		procMoveWindow.Call(hwnd, uintptr(x), uintptr(y), uintptr(w), uintptr(h), 1)
	}
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
			// Move selection down in listbox. LB_GETCURSEL returns LB_ERR
			// (-1) when nothing is selected, so the +1 selects the first item.
			count, _, _ := procSendMessageW.Call(listHwnd, LB_GETCOUNT, 0, 0)
			cur, _, _ := procSendMessageW.Call(listHwnd, LB_GETCURSEL, 0, 0)
			if next := int(cur) + 1; next < int(count) {
				procSendMessageW.Call(listHwnd, LB_SETCURSEL, uintptr(next), 0)
			}
			return 0
		case VK_UP:
			// Move selection up in listbox
			cur, _, _ := procSendMessageW.Call(listHwnd, LB_GETCURSEL, 0, 0)
			if int(cur) > 0 {
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
			showSettingsDialog()
			return 0
		}
	}

	ret, _, _ := procCallWindowProcW.Call(originalEditProc, hwnd, uintptr(msg), wParam, lParam)
	return ret
}

// deleteWordBackward deletes the word before the cursor.
// EM_GETSEL/EM_SETSEL positions are UTF-16 code-unit indices, so the text
// is walked as UTF-16 rather than as a Go (UTF-8) string.
func deleteWordBackward(hwnd uintptr) {
	length, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if length == 0 {
		return
	}

	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), length+1)

	// Get cursor position
	var start, end uint32
	procSendMessageW.Call(hwnd, EM_GETSEL, uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&end)))

	if start == 0 {
		return
	}

	isSpace := func(c uint16) bool { return c == ' ' || c == '\t' }

	// Find word boundary (skip spaces, then skip non-spaces)
	pos := int(start)
	if pos > int(length) {
		pos = int(length)
	}
	for pos > 0 && isSpace(buf[pos-1]) {
		pos--
	}
	for pos > 0 && !isSpace(buf[pos-1]) {
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

	margin := dpiScale(10)
	gap := dpiScale(6)
	sortBtnWidth := dpiScale(90)
	settingsBtnWidth := dpiScale(30)
	ctrlHeight := dpiScale(30)
	listTop := margin + ctrlHeight + margin

	totalBtnWidth := sortBtnWidth + gap + settingsBtnWidth
	editWidth := width - totalBtnWidth - margin*2 - gap

	procMoveWindow.Call(editHwnd, uintptr(margin), uintptr(margin), uintptr(editWidth), uintptr(ctrlHeight), 1)
	procMoveWindow.Call(sortBtnHwnd, uintptr(margin+editWidth+gap), uintptr(margin), uintptr(sortBtnWidth), uintptr(ctrlHeight), 1)
	procMoveWindow.Call(settingsBtnHwnd, uintptr(width-settingsBtnWidth-margin), uintptr(margin), uintptr(settingsBtnWidth), uintptr(ctrlHeight), 1)
	procMoveWindow.Call(listHwnd, uintptr(margin), uintptr(listTop), uintptr(width-margin*2), uintptr(height-listTop-margin), 1)
	procInvalidateRect.Call(listHwnd, 0, 1)
}

func populateList() {
	procSendMessageW.Call(listHwnd, LB_RESETCONTENT, 0, 0)

	for _, proj := range filteredProjects {
		// Add the project name as the string (for accessibility)
		text := utf16PtrFromString(proj.Name)
		procSendMessageW.Call(listHwnd, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(text)))
	}

	if len(filteredProjects) > 0 {
		procSendMessageW.Call(listHwnd, LB_SETCURSEL, 0, 0)
	}
}

// rebuildSearchNames refreshes the per-project search strings. Must be
// called whenever allProjects is reassigned or reordered.
func rebuildSearchNames() {
	searchNames = searchNames[:0]
	for _, p := range allProjects {
		searchNames = append(searchNames, p.Name+" "+p.Path)
	}
}

func onSearchChanged() {
	searchText := getWindowText(editHwnd)
	if searchText == "" {
		filteredProjects = allProjects
		populateList()
		return
	}

	scored := fuzzy.FilterAndScore(searchText, searchNames)

	filteredProjects = nil
	for _, item := range scored {
		filteredProjects = append(filteredProjects, allProjects[item.Index])
	}

	populateList()
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
	rebuildSearchNames()

	populateList()
}

func onProjectSelected() {
	sel, _, _ := procSendMessageW.Call(listHwnd, LB_GETCURSEL, 0, 0)
	// LB_ERR (-1) comes back as an all-ones uintptr; convert to int so the
	// no-selection case (e.g. Enter on an empty filter result) is caught.
	idx := int(sel)
	if idx < 0 || idx >= len(filteredProjects) {
		return
	}

	proj := &filteredProjects[idx]

	// Check if project path exists
	if !proj.PathExists {
		showMessageBox(mainHwnd,
			"The project directory no longer exists:\n\n"+proj.Path+"\n\n"+
				"It may have been moved or deleted.",
			"Project Not Found", win32.MB_ICONERROR)
		return
	}

	// Show opening indication
	procSetWindowTextW.Call(mainHwnd, uintptr(unsafe.Pointer(utf16PtrFromString(fmt.Sprintf("Opening %s...", proj.Name)))))
	enableMainControls(false)

	// Open the configured terminal
	// Set flag to prevent close on focus loss during terminal dialogs
	showingDialog.Store(true)
	err := terminal.OpenProject(proj.Path, appConfig.Terminal)
	showingDialog.Store(false)

	if err != nil {
		// Restore UI on failure
		procSetWindowTextW.Call(mainHwnd, uintptr(unsafe.Pointer(utf16PtrFromString("Claude Code Switcher"))))
		enableMainControls(true)
		showMessageBox(mainHwnd, "Failed to open terminal: "+err.Error(), "Error", win32.MB_ICONERROR)
		return
	}

	procDestroyWindow.Call(mainHwnd)
}

func enableMainControls(enabled bool) {
	var v uintptr
	if enabled {
		v = 1
	}
	procEnableWindow.Call(editHwnd, v)
	procEnableWindow.Call(listHwnd, v)
	procEnableWindow.Call(sortBtnHwnd, v)
	procEnableWindow.Call(settingsBtnHwnd, v)
}
