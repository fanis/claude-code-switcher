// Copyright (c) 2025 Fanis Hatzidakis
// Licensed under PolyForm Internal Use License 1.0.0 - see LICENSE.md

//go:build windows

// Main window: application state, Run entry point, and the window procedure.

package gui

import (
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/fanis/claude-code-switcher/internal/config"
	"github.com/fanis/claude-code-switcher/internal/projects"
	"github.com/fanis/claude-code-switcher/internal/terminal"
	"github.com/fanis/claude-code-switcher/internal/update"
	"github.com/fanis/claude-code-switcher/internal/win32"
)

const (
	IDC_EDIT     = 101
	IDC_LISTBOX  = 102
	IDC_SORT     = 103
	IDC_SETTINGS = 104
)

var (
	mainHwnd         uintptr
	editHwnd         uintptr
	listHwnd         uintptr
	sortBtnHwnd      uintptr
	settingsBtnHwnd  uintptr
	hFont            uintptr
	hFontBold        uintptr
	hFontSmall       uintptr
	originalEditProc uintptr
	currentDPI       uint32  = 96 // Default DPI
	hWhiteBrush      uintptr      // also the normal list item background
	hSelectedBrush   uintptr
	hMissingBrush    uintptr

	allProjects      []projects.Project
	filteredProjects []projects.Project
	searchNames      []string // "Name Path" per allProjects entry, for fuzzy search
	sortByName       bool
	appVersion       string
	appConfig        *config.Config

	// showingDialog prevents close-on-focus-loss while a dialog is up.
	// Atomic because showUpdateNotification resets it from a goroutine.
	showingDialog atomic.Bool

	// configMu guards appConfig mutations and config.Save calls, which can
	// happen concurrently from the GUI thread and the update-check goroutine.
	configMu sync.Mutex
)

// saveConfig applies mutate to appConfig and persists it, serialized
// against the background update-check goroutine.
func saveConfig(mutate func()) {
	configMu.Lock()
	defer configMu.Unlock()
	mutate()
	config.Save(appConfig)
}

func Run(projectList []projects.Project, version string, cfg *config.Config) {
	allProjects = projectList
	filteredProjects = projectList
	appVersion = version
	appConfig = cfg
	rebuildSearchNames()

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
	if hwnd == 0 {
		win32.MessageBox(0, "Failed to create the main window.", "Claude Code Switcher",
			win32.MB_OK|win32.MB_ICONERROR)
		return
	}

	mainHwnd = hwnd
	terminal.SetParentHwnd(hwnd)

	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)
	procSetForegroundWindow.Call(hwnd)
	procSetFocus.Call(editHwnd)

	// One-time onboarding: ask about update notifications
	if !appConfig.AskedAboutUpdates {
		result := showMessageBox(hwnd,
			"Welcome! Thanks for installing Claude Code Switcher.\n\n"+
				"Would you like to be notified when a new version is available?\n\n"+
				"You can change this later in Settings.",
			"Claude Code Switcher",
			win32.MB_YESNO|win32.MB_ICONQUESTION)
		saveConfig(func() {
			appConfig.AskedAboutUpdates = true
			appConfig.UpdateCheckEnabled = result == win32.IDYES
		})
	}

	// Show pending update notification from a previous session
	if appConfig.PendingVersion != "" && update.IsNewer(appVersion, appConfig.PendingVersion) {
		procPostMessageW.Call(hwnd, WM_APP_UPDATE, 0, 0)
	}

	// Start background update check for next session
	startUpdateCheck()

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

	case WM_GETMINMAXINFO:
		mmi := (*MINMAXINFO)(unsafe.Pointer(lParam))
		mmi.MinTrackSize.X = dpiScale(400)
		mmi.MinTrackSize.Y = dpiScale(200)
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
		case IDC_SETTINGS:
			showSettingsDialog()
		}
		return 0

	case WM_APP_UPDATE:
		showUpdateNotification()
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
			mis.ItemHeight = uint32(dpiScale(40))
			return 1
		}
		return 0

	case WM_ACTIVATE:
		// Close window when it loses focus (launcher-style behavior)
		// But not if we're showing a dialog
		if wParam&0xFFFF == WA_INACTIVE && !showingDialog.Load() {
			// Post WM_CLOSE to close gracefully without re-entrancy issues
			procPostMessageW.Call(hwnd, WM_CLOSE, 0, 0)
		}
		return 0

	case WM_DESTROY:
		deleteObject(hFont)
		deleteObject(hFontBold)
		deleteObject(hFontSmall)
		deleteObject(hWhiteBrush)
		deleteObject(hSelectedBrush)
		deleteObject(hMissingBrush)
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}
