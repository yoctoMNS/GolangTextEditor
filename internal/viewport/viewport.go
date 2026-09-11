// Package viewport implements the pure scrolling arithmetic behind
// internal/app's line-based text viewport (which lines are visible, and
// how far to scroll to keep a given line in view). It has no dependency
// on any rendering library so it can be unit tested headlessly, unlike
// internal/app itself (see CLAUDE.md and internal/app/app.go).
package viewport

// LineCount returns how many text lines of lineHeight fit within
// availableHeight, always at least 1.
func LineCount(availableHeight, lineHeight int) int {
	lines := availableHeight / lineHeight
	if lines < 1 {
		lines = 1
	}
	return lines
}

// ClampScroll returns the topmost visible line so that cursorLine stays
// within the viewport of visibleLines rows, scrolling by the minimum
// amount needed rather than always re-centering.
func ClampScroll(scroll, cursorLine, visibleLines, lineCount int) int {
	if cursorLine < scroll {
		scroll = cursorLine
	}
	if cursorLine >= scroll+visibleLines {
		scroll = cursorLine - visibleLines + 1
	}
	if maxScroll := lineCount - visibleLines; scroll > maxScroll {
		scroll = maxScroll
	}
	if scroll < 0 {
		scroll = 0
	}
	return scroll
}
