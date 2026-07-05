// Copyright (c) 2025 Fanis Hatzidakis
// Licensed under PolyForm Internal Use License 1.0.0 - see LICENSE.md

//go:build windows

// Two-phase update check: a background goroutine records a pending version
// in the config; the next launch notifies via WM_APP_UPDATE.

package gui

import (
	"fmt"
	"time"

	"github.com/fanis/claude-code-switcher/internal/config"
	"github.com/fanis/claude-code-switcher/internal/update"
	"github.com/fanis/claude-code-switcher/internal/win32"
)

const checkDateLayout = "2006-01-02"

// startUpdateCheck launches the background release check (at most once per
// day). The result is stored for the NEXT launch to show.
func startUpdateCheck() {
	if !appConfig.UpdateCheckEnabled || appConfig.LastCheckDate == time.Now().Format(checkDateLayout) {
		return
	}
	go func() {
		latest, url, err := update.CheckLatest()
		if err != nil || !update.IsNewer(appVersion, latest) {
			return
		}
		// Mutate appConfig only under configMu: the GUI thread may be
		// saving settings changes concurrently.
		configMu.Lock()
		defer configMu.Unlock()
		if latest == appConfig.DismissedVersion {
			return
		}
		appConfig.PendingVersion = latest
		appConfig.PendingURL = url
		appConfig.LastCheckDate = time.Now().Format(checkDateLayout)
		config.Save(appConfig)
	}()
}

func showUpdateNotification() {
	var version, url string
	// Clear pending so we don't show again
	saveConfig(func() {
		version = appConfig.PendingVersion
		url = appConfig.PendingURL
		appConfig.PendingVersion = ""
		appConfig.PendingURL = ""
	})

	result := showMessageBox(mainHwnd,
		fmt.Sprintf("Version %s is available.\n\nOpen the download page?", version),
		"Update Available",
		win32.MB_YESNO|win32.MB_ICONQUESTION)

	if result == win32.IDYES {
		// Keep showingDialog true while opening browser to prevent
		// the main window from closing on focus loss. Reset after a delay
		// since ShellExecute returns before the browser steals focus.
		showingDialog.Store(true)
		openURL(url)
		go func() {
			time.Sleep(2 * time.Second)
			showingDialog.Store(false)
		}()
	} else {
		// User dismissed - don't notify again for this version
		saveConfig(func() {
			appConfig.DismissedVersion = version
		})
	}
}
