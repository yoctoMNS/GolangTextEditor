package editor

import "github.com/yoctoMNS/GolangTextEditor/internal/buffer"

// StartSelection anchors a selection at the current caret position. It is
// a no-op if a selection is already active, so callers can invoke it on
// every selecting movement (e.g. each Shift+Arrow press) without resetting
// the anchor each time.
func (e *Editor) StartSelection() {
	if e.Anchor == nil {
		a := e.Cursor
		e.Anchor = &a
	}
}

// ClearSelection deactivates the current selection, if any.
func (e *Editor) ClearSelection() {
	e.Anchor = nil
}

// HasSelection reports whether a non-empty selection is active. An anchor
// equal to the current caret position (e.g. after Shift+Arrow moved back
// onto its starting point) counts as no selection.
func (e *Editor) HasSelection() bool {
	return e.Anchor != nil && *e.Anchor != e.Cursor
}

// SelectionRange returns the current selection's start and end in document
// order (start <= end), regardless of whether the caret is before or after
// the anchor. ok is false when there is no active selection.
func (e *Editor) SelectionRange() (start, end buffer.Position, ok bool) {
	if !e.HasSelection() {
		return buffer.Position{}, buffer.Position{}, false
	}
	anchor, cursor := *e.Anchor, e.Cursor
	if positionLess(cursor, anchor) {
		return cursor, anchor, true
	}
	return anchor, cursor, true
}

// SelectedText returns the text covered by the current selection, or ""
// if there is no active selection.
func (e *Editor) SelectedText() string {
	start, end, ok := e.SelectionRange()
	if !ok {
		return ""
	}
	return e.Buf.TextRange(start, end)
}

func positionLess(a, b buffer.Position) bool {
	if a.Line != b.Line {
		return a.Line < b.Line
	}
	return a.Col < b.Col
}
