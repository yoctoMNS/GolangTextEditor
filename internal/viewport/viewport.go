// Package viewport implements the pure scrolling arithmetic behind
// internal/app's line-based text viewport (which lines are visible, and
// how far to scroll to keep a given line in view). It has no dependency
// on any rendering library so it can be unit tested headlessly, unlike
// internal/app itself (see CLAUDE.md and internal/app/app.go).
package viewport

import "strconv"

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

// PositionAt returns the (line, col) a click at pixel (x, y) targets,
// given the pixel origin of the text area (originX, originY, i.e. after
// margins and the line-number gutter), the topmost visible buffer line
// (scrollLine), and the pixel size of one character cell (charWidth,
// lineHeight). The result is not clamped to the buffer's actual bounds
// (a click can target a line/column past the end of a short file or
// line); callers should clamp it, e.g. via buffer.Buffer.Clamp, before
// using it as a caret position.
func PositionAt(x, y, originX, originY, scrollLine, charWidth, lineHeight int) (line, col int) {
	line = scrollLine + floorDiv(y-originY, lineHeight)
	col = floorDiv(x-originX, charWidth)
	return line, col
}

// floorDiv divides a by b rounding toward negative infinity, unlike Go's
// built-in / which rounds toward zero. This matters for PositionAt because
// a click above or left of the text origin (a negative numerator) should
// map to an earlier line/column, not clamp toward 0 the way truncating
// division would.
func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// GutterWidth returns the pixel width of a line-number gutter wide enough
// to fit lineCount (at least 2 digits, so short files don't get an oddly
// narrow gutter), given the width of one character and the horizontal
// padding to leave on each side of the digits.
func GutterWidth(lineCount, charWidth, paddingX int) int {
	digits := len(strconv.Itoa(lineCount))
	if digits < 2 {
		digits = 2
	}
	return digits*charWidth + paddingX*2
}
