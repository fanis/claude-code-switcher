// Copyright (c) 2025 Fanis Hatzidakis
// Licensed under PolyForm Internal Use License 1.0.0 - see LICENSE.md

//go:build windows

// Owner-drawn listbox item rendering.

package gui

import (
	"fmt"
	"syscall"
	"time"
)

func drawListItem(dis *DRAWITEMSTRUCT) {
	if dis.ItemID == 0xFFFFFFFF {
		return
	}

	idx := int(dis.ItemID)
	if idx >= len(filteredProjects) {
		return
	}

	proj := filteredProjects[idx]

	// Modern color scheme (colors in BGR format for Windows).
	// Background brushes are created once in createControls.
	var bgColor, textColor, secondaryColor uint32
	var bgBrush uintptr
	if dis.ItemState&ODS_SELECTED != 0 {
		bgColor = 0x00CC7A00        // Nice blue (#007ACC in RGB)
		textColor = 0x00FFFFFF      // White
		secondaryColor = 0x00E0E0E0 // Light gray
		bgBrush = hSelectedBrush
	} else if !proj.PathExists {
		bgColor = 0x00F0F0F0        // Light gray background
		textColor = 0x00808080      // Gray text
		secondaryColor = 0x00A0A0A0 // Lighter gray
		bgBrush = hMissingBrush
	} else {
		bgColor = 0x00FFFFFF        // White
		textColor = 0x00202020      // Near black
		secondaryColor = 0x00808080 // Gray
		bgBrush = hWhiteBrush
	}

	// Fill background
	setBkColor(dis.HDC, bgColor)
	setTextColor(dis.HDC, textColor)
	fillRect(dis.HDC, &dis.RcItem, syscall.Handle(bgBrush))

	// Draw project name (first line)
	nameRect := dis.RcItem
	nameRect.Left += dpiScale(8)
	nameRect.Top += dpiScale(4)
	nameRect.Bottom = nameRect.Top + dpiScale(18)

	nameText := proj.Name
	if !proj.PathExists {
		nameText = "[NOT FOUND] " + nameText
	}
	drawText(dis.HDC, nameText, &nameRect, DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)

	// Draw last used timestamp (first line, right-aligned)
	lastUsedStr := formatLastUsed(proj.LastUsed)
	timeRect := dis.RcItem
	timeRect.Right -= dpiScale(8)
	timeRect.Top += dpiScale(4)
	timeRect.Bottom = timeRect.Top + dpiScale(18)
	setTextColor(dis.HDC, secondaryColor)
	drawText(dis.HDC, lastUsedStr, &timeRect, DT_RIGHT|DT_SINGLELINE)

	// Draw path (second line)
	infoRect := dis.RcItem
	infoRect.Left += dpiScale(8)
	infoRect.Top += dpiScale(22)
	infoRect.Bottom = infoRect.Top + dpiScale(16)
	drawText(dis.HDC, proj.Path, &infoRect, DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
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
