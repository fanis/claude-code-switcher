// Copyright (c) 2025 Fanis Hatzidakis
// Licensed under PolyForm Internal Use License 1.0.0 - see LICENSE.md

//go:build windows

// Settings dialog: update check toggle, terminal selector, about info.

package gui

import (
	"sync"
	"syscall"
	"unsafe"
)

const (
	IDC_SETTINGS_CHECK    = 201
	IDC_SETTINGS_GITHUB   = 202
	IDC_SETTINGS_OK       = 203
	IDC_SETTINGS_TERMINAL = 204
	IDC_SETTINGS_CUSTOM   = 205

	settingsClassName = "ClaudeSettingsDialog"
)

var (
	settingsDlgHwnd         uintptr
	settingsCustomEditHwnd  uintptr
	settingsCustomLabelHwnd uintptr

	// The window class (and its permanent syscall.NewCallback) is
	// registered once; re-registering on every open would leak a callback
	// each time and fail anyway since the class already exists.
	settingsClassOnce sync.Once
)

func showSettingsDialog() {
	showingDialog.Store(true)
	defer func() {
		showingDialog.Store(false)
		procSetFocus.Call(editHwnd)
	}()

	hInstance, _, _ := procGetModuleHandleW.Call(0)

	className := utf16PtrFromString(settingsClassName)
	settingsClassOnce.Do(func() {
		wc := WNDCLASSEXW{
			Size:       uint32(unsafe.Sizeof(WNDCLASSEXW{})),
			WndProc:    syscall.NewCallback(settingsDlgProc),
			Instance:   syscall.Handle(hInstance),
			ClassName:  className,
			Background: syscall.Handle(COLOR_WINDOW + 1),
		}
		procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	})

	// Center on main window
	var mainRect RECT
	procGetClientRect.Call(mainHwnd, uintptr(unsafe.Pointer(&mainRect)))
	var pt POINT
	pt.X = mainRect.Left
	pt.Y = mainRect.Top
	procClientToScreen.Call(mainHwnd, uintptr(unsafe.Pointer(&pt)))

	dlgWidth := dpiScale(310)
	dlgHeight := dpiScale(400)
	dlgX := pt.X + (mainRect.Right-mainRect.Left-dlgWidth)/2
	dlgY := pt.Y + (mainRect.Bottom-mainRect.Top-dlgHeight)/2

	procEnableWindow.Call(mainHwnd, 0)

	settingsDlgHwnd, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16PtrFromString("Settings"))),
		WS_POPUP|WS_CAPTION|WS_SYSMENU,
		uintptr(dlgX), uintptr(dlgY),
		uintptr(dlgWidth), uintptr(dlgHeight),
		mainHwnd, 0, hInstance, 0,
	)

	createSettingsControls(settingsDlgHwnd, hInstance)

	procShowWindow.Call(settingsDlgHwnd, SW_SHOW)
	procUpdateWindow.Call(settingsDlgHwnd)

	// Modal message loop. The loop condition also catches the dialog being
	// destroyed inside DispatchMessage (OK button, close box), so it exits
	// immediately instead of blocking in GetMessageW until the next message.
	var msg MSG
	for settingsDlgHwnd != 0 {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			// WM_QUIT: re-post it so the main message loop terminates too
			procPostQuitMessage.Call(msg.WParam)
			break
		}
		if msg.Message == WM_KEYDOWN && msg.WParam == VK_ESCAPE {
			procDestroyWindow.Call(settingsDlgHwnd)
			break
		}
		// IsDialogMessageW handles Tab/Shift+Tab focus cycling
		handled, _, _ := procIsDialogMessageW.Call(settingsDlgHwnd, uintptr(unsafe.Pointer(&msg)))
		if handled != 0 {
			continue
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	procEnableWindow.Call(mainHwnd, 1)
}

func createSettingsControls(hwnd uintptr, hInstance uintptr) {
	su := func(base int32) uintptr {
		return uintptr(dpiScale(base))
	}

	// Vertical layout cursor: each control is placed at y and advances it
	// by its height plus a gap, so inserting a row only needs local edits.
	y := int32(12)
	row := func(height, gapAfter int32) (top uintptr) {
		top = su(y)
		y += height + gapAfter
		return top
	}

	static := func(text string, style uintptr, x, w uintptr, top uintptr, h int32, font uintptr) uintptr {
		var textArg uintptr
		if text != "" {
			textArg = uintptr(unsafe.Pointer(utf16PtrFromString(text)))
		}
		ctl, _, _ := procCreateWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(utf16PtrFromString("STATIC"))),
			textArg,
			style,
			x, top, w, su(h),
			hwnd, 0, hInstance, 0,
		)
		if font != 0 {
			procSendMessageW.Call(ctl, WM_SETFONT, font, 1)
		}
		return ctl
	}

	// --- Updates section ---
	static("Updates", WS_CHILD|WS_VISIBLE, su(15), su(270), row(18, 4), 18, hFontBold)

	checkHwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("BUTTON"))),
		uintptr(unsafe.Pointer(utf16PtrFromString("Check for new versions on startup"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX,
		su(15), row(20, 2), su(270), su(20),
		hwnd, IDC_SETTINGS_CHECK, hInstance, 0,
	)
	procSendMessageW.Call(checkHwnd, WM_SETFONT, hFont, 1)
	if appConfig.UpdateCheckEnabled {
		procSendMessageW.Call(checkHwnd, BM_SETCHECK, BST_CHECKED, 0)
	}

	static("When a new version is found, you will be\nnotified once on the next launch.",
		WS_CHILD|WS_VISIBLE, su(32), su(260), row(32, 8), 32, hFontSmall)

	// --- Separator ---
	static("", WS_CHILD|WS_VISIBLE|SS_ETCHEDHORZ, su(15), su(270), row(2, 8), 2, 0)

	// --- Terminal section ---
	static("Terminal", WS_CHILD|WS_VISIBLE, su(15), su(270), row(18, 4), 18, hFontBold)

	comboHwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("COMBOBOX"))),
		0,
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|CBS_DROPDOWNLIST|CBS_HASSTRINGS,
		su(15), row(26, 0), su(270), su(150), // height includes the open dropdown
		hwnd, IDC_SETTINGS_TERMINAL, hInstance, 0,
	)
	procSendMessageW.Call(comboHwnd, WM_SETFONT, hFont, 1)

	termOptions := []string{"Auto-detect", "Windows Terminal", "WezTerm", "cmd.exe", "Custom..."}
	for _, opt := range termOptions {
		procSendMessageW.Call(comboHwnd, CB_ADDSTRING, 0,
			uintptr(unsafe.Pointer(utf16PtrFromString(opt))))
	}

	// Map config value to combo index
	termIndex := 0
	customValue := ""
	switch appConfig.Terminal {
	case "":
		termIndex = 0
	case "wt":
		termIndex = 1
	case "wezterm":
		termIndex = 2
	case "cmd":
		termIndex = 3
	default:
		termIndex = 4
		customValue = appConfig.Terminal
	}
	procSendMessageW.Call(comboHwnd, CB_SETCURSEL, uintptr(termIndex), 0)

	// Custom command label (visible only when Custom is selected)
	customLabelStyle := uintptr(WS_CHILD)
	if termIndex == 4 {
		customLabelStyle |= WS_VISIBLE
	}
	settingsCustomLabelHwnd = static("Use {dir} and {claude} as placeholders:",
		customLabelStyle, su(15), su(270), row(16, 2), 16, hFontSmall)

	// Custom command edit
	customStyle := uintptr(WS_CHILD | WS_BORDER | WS_TABSTOP | ES_AUTOHSCROLL)
	if termIndex == 4 {
		customStyle |= WS_VISIBLE
	}
	settingsCustomEditHwnd, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("EDIT"))),
		uintptr(unsafe.Pointer(utf16PtrFromString(customValue))),
		customStyle,
		su(15), row(24, 10), su(270), su(24),
		hwnd, IDC_SETTINGS_CUSTOM, hInstance, 0,
	)
	procSendMessageW.Call(settingsCustomEditHwnd, WM_SETFONT, hFont, 1)

	// --- Separator ---
	static("", WS_CHILD|WS_VISIBLE|SS_ETCHEDHORZ, su(15), su(270), row(2, 8), 2, 0)

	// --- About section ---
	static("About", WS_CHILD|WS_VISIBLE, su(15), su(270), row(18, 4), 18, hFontBold)

	titleText := "Claude Code Switcher"
	if appVersion != "" {
		titleText += " v" + appVersion
	}
	static(titleText, WS_CHILD|WS_VISIBLE|SS_CENTER, su(15), su(270), row(18, 0), 18, hFont)
	static("by Fanis Hatzidakis", WS_CHILD|WS_VISIBLE|SS_CENTER, su(15), su(270), row(16, 38), 16, hFontSmall)

	// --- Bottom buttons ---
	btnTop := row(26, 0)
	githubHwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("BUTTON"))),
		uintptr(unsafe.Pointer(utf16PtrFromString("Open GitHub"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP,
		su(15), btnTop, su(100), su(26),
		hwnd, IDC_SETTINGS_GITHUB, hInstance, 0,
	)
	procSendMessageW.Call(githubHwnd, WM_SETFONT, hFont, 1)

	okHwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16PtrFromString("BUTTON"))),
		uintptr(unsafe.Pointer(utf16PtrFromString("OK"))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP,
		su(200), btnTop, su(80), su(26),
		hwnd, IDC_SETTINGS_OK, hInstance, 0,
	)
	procSendMessageW.Call(okHwnd, WM_SETFONT, hFont, 1)
}

func settingsDlgProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_KEYDOWN:
		if wParam == VK_ESCAPE {
			procDestroyWindow.Call(hwnd)
			return 0
		}
	case WM_COMMAND:
		wmId := wParam & 0xFFFF
		wmEvent := (wParam >> 16) & 0xFFFF
		switch wmId {
		case IDC_SETTINGS_OK:
			procDestroyWindow.Call(hwnd)
			return 0
		case IDC_SETTINGS_GITHUB:
			openURL("https://github.com/fanis/claude-code-switcher")
			return 0
		case IDC_SETTINGS_CHECK:
			if wmEvent == 0 { // BN_CLICKED
				checked, _, _ := procSendMessageW.Call(
					getDlgItem(hwnd, IDC_SETTINGS_CHECK), BM_GETCHECK, 0, 0)
				saveConfig(func() {
					appConfig.UpdateCheckEnabled = checked == BST_CHECKED
				})
			}
			return 0
		case IDC_SETTINGS_TERMINAL:
			if wmEvent == CBN_SELCHANGE {
				sel, _, _ := procSendMessageW.Call(
					getDlgItem(hwnd, IDC_SETTINGS_TERMINAL), CB_GETCURSEL, 0, 0)
				termValues := []string{"", "wt", "wezterm", "cmd"}
				if int(sel) >= 0 && int(sel) < len(termValues) {
					saveConfig(func() {
						appConfig.Terminal = termValues[sel]
					})
					procShowWindow.Call(settingsCustomLabelHwnd, SW_HIDE)
					procShowWindow.Call(settingsCustomEditHwnd, SW_HIDE)
				} else if int(sel) == len(termValues) {
					// Custom selected
					saveConfig(func() {
						appConfig.Terminal = getWindowText(settingsCustomEditHwnd)
					})
					procShowWindow.Call(settingsCustomLabelHwnd, SW_SHOW)
					procShowWindow.Call(settingsCustomEditHwnd, SW_SHOW)
					procSetFocus.Call(settingsCustomEditHwnd)
				}
			}
			return 0
		case IDC_SETTINGS_CUSTOM:
			if wmEvent == EN_CHANGE {
				// Only save if Custom is selected in the dropdown
				sel, _, _ := procSendMessageW.Call(
					getDlgItem(hwnd, IDC_SETTINGS_TERMINAL), CB_GETCURSEL, 0, 0)
				if sel == 4 {
					saveConfig(func() {
						appConfig.Terminal = getWindowText(settingsCustomEditHwnd)
					})
				}
			}
			return 0
		}
	case WM_CTLCOLORSTATIC:
		return hWhiteBrush
	case WM_CLOSE:
		procDestroyWindow.Call(hwnd)
		return 0
	case WM_DESTROY:
		settingsDlgHwnd = 0
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}
