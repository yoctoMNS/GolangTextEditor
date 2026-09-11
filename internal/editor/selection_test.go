package editor

import (
	"testing"

	"github.com/yoctoMNS/GolangTextEditor/internal/buffer"
)

func TestStartSelectionIsIdempotent(t *testing.T) {
	e := New()
	e.Buf = buffer.New("hello world")
	e.Cursor = buffer.Position{Line: 0, Col: 2}

	e.StartSelection()
	e.Cursor = buffer.Position{Line: 0, Col: 5}
	e.StartSelection() // must not move the anchor to the new cursor position

	start, end, ok := e.SelectionRange()
	if !ok {
		t.Fatal("expected a selection to be active")
	}
	if want := (buffer.Position{Line: 0, Col: 2}); start != want {
		t.Errorf("start = %+v, want %+v", start, want)
	}
	if want := (buffer.Position{Line: 0, Col: 5}); end != want {
		t.Errorf("end = %+v, want %+v", end, want)
	}
}

func TestClearSelection(t *testing.T) {
	e := New()
	e.StartSelection()
	e.ClearSelection()

	if e.HasSelection() {
		t.Error("expected HasSelection() to be false after ClearSelection()")
	}
	if _, _, ok := e.SelectionRange(); ok {
		t.Error("expected SelectionRange() ok to be false after ClearSelection()")
	}
}

func TestHasSelection(t *testing.T) {
	tests := []struct {
		name        string
		anchorFirst bool // start selection before moving the cursor
		moveCursor  bool // move the cursor away from the anchor
		want        bool
	}{
		{
			name:        "no anchor set",
			anchorFirst: false,
			moveCursor:  false,
			want:        false,
		},
		{
			name:        "anchor equals cursor (empty selection)",
			anchorFirst: true,
			moveCursor:  false,
			want:        false,
		},
		{
			name:        "anchor differs from cursor",
			anchorFirst: true,
			moveCursor:  true,
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()
			e.Buf = buffer.New("hello world")
			if tt.anchorFirst {
				e.StartSelection()
			}
			if tt.moveCursor {
				e.Cursor.Col = 5
			}
			if got := e.HasSelection(); got != tt.want {
				t.Errorf("HasSelection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectionRangeNormalizesDirection(t *testing.T) {
	e := New()
	e.Buf = buffer.New("hello world")

	// Select backward: anchor after cursor.
	e.Cursor = buffer.Position{Line: 0, Col: 8}
	e.StartSelection()
	e.Cursor = buffer.Position{Line: 0, Col: 3}

	start, end, ok := e.SelectionRange()
	if !ok {
		t.Fatal("expected a selection to be active")
	}
	if want := (buffer.Position{Line: 0, Col: 3}); start != want {
		t.Errorf("start = %+v, want %+v", start, want)
	}
	if want := (buffer.Position{Line: 0, Col: 8}); end != want {
		t.Errorf("end = %+v, want %+v", end, want)
	}
}

func TestSelectionOnLine(t *testing.T) {
	// Selection spans "lo\nworld\nhe" across three lines of "hello\nworld\nhey".
	anchor := buffer.Position{Line: 0, Col: 3}
	cursor := buffer.Position{Line: 2, Col: 2}

	tests := []struct {
		name         string
		line         int
		wantColStart int
		wantColEnd   int
		wantOK       bool
	}{
		{
			name:   "line before selection",
			line:   -1, // not a real line, but out of range either way
			wantOK: false,
		},
		{
			name:         "start line: from anchor column to end of line",
			line:         0,
			wantColStart: 3,
			wantColEnd:   5, // len("hello")
			wantOK:       true,
		},
		{
			name:         "middle line: fully selected",
			line:         1,
			wantColStart: 0,
			wantColEnd:   5, // len("world")
			wantOK:       true,
		},
		{
			name:         "end line: from start of line to cursor column",
			line:         2,
			wantColStart: 0,
			wantColEnd:   2,
			wantOK:       true,
		},
		{
			name:   "line after selection",
			line:   3,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()
			e.Buf = buffer.New("hello\nworld\nhey")
			a := anchor
			e.Anchor = &a
			e.Cursor = cursor

			colStart, colEnd, ok := e.SelectionOnLine(tt.line)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if colStart != tt.wantColStart {
				t.Errorf("colStart = %d, want %d", colStart, tt.wantColStart)
			}
			if colEnd != tt.wantColEnd {
				t.Errorf("colEnd = %d, want %d", colEnd, tt.wantColEnd)
			}
		})
	}
}

func TestSelectionOnLineNoSelection(t *testing.T) {
	e := New()
	e.Buf = buffer.New("hello")
	if _, _, ok := e.SelectionOnLine(0); ok {
		t.Error("expected ok to be false when there is no active selection")
	}
}

func TestInsertRuneReplacesSelection(t *testing.T) {
	e := New()
	e.Buf = buffer.New("hello world")
	e.Cursor = buffer.Position{Line: 0, Col: 0}
	e.StartSelection()
	e.Cursor = buffer.Position{Line: 0, Col: 5} // selects "hello"

	e.InsertRune('X')

	if got, want := e.Buf.String(), "X world"; got != want {
		t.Fatalf("Buf.String() = %q, want %q", got, want)
	}
	if want := (buffer.Position{Line: 0, Col: 1}); e.Cursor != want {
		t.Fatalf("Cursor = %+v, want %+v", e.Cursor, want)
	}
	if e.HasSelection() {
		t.Error("expected selection to be cleared")
	}
	if !e.Modified {
		t.Error("expected Modified to be true")
	}
}

func TestInsertNewlineReplacesSelection(t *testing.T) {
	e := New()
	e.Buf = buffer.New("hello world")
	e.Cursor = buffer.Position{Line: 0, Col: 0}
	e.StartSelection()
	e.Cursor = buffer.Position{Line: 0, Col: 5} // selects "hello"

	e.InsertNewline()

	if got, want := e.Buf.String(), "\n world"; got != want {
		t.Fatalf("Buf.String() = %q, want %q", got, want)
	}
	if want := (buffer.Position{Line: 1, Col: 0}); e.Cursor != want {
		t.Fatalf("Cursor = %+v, want %+v", e.Cursor, want)
	}
	if e.HasSelection() {
		t.Error("expected selection to be cleared")
	}
}

func TestBackspaceDeletesSelectionOnly(t *testing.T) {
	e := New()
	e.Buf = buffer.New("hello world")
	e.Cursor = buffer.Position{Line: 0, Col: 6}
	e.StartSelection()
	e.Cursor = buffer.Position{Line: 0, Col: 11} // selects "world"

	e.Backspace()

	// If Backspace additionally deleted the rune before the selection,
	// this would be "hello" instead of "hello ".
	if got, want := e.Buf.String(), "hello "; got != want {
		t.Fatalf("Buf.String() = %q, want %q", got, want)
	}
	if want := (buffer.Position{Line: 0, Col: 6}); e.Cursor != want {
		t.Fatalf("Cursor = %+v, want %+v", e.Cursor, want)
	}
	if e.HasSelection() {
		t.Error("expected selection to be cleared")
	}
}

func TestDeleteDeletesSelectionOnly(t *testing.T) {
	e := New()
	e.Buf = buffer.New("hello world")
	e.Cursor = buffer.Position{Line: 0, Col: 0}
	e.StartSelection()
	e.Cursor = buffer.Position{Line: 0, Col: 5} // selects "hello"

	e.Delete()

	// If Delete additionally deleted the rune after the selection, this
	// would be "world" instead of " world".
	if got, want := e.Buf.String(), " world"; got != want {
		t.Fatalf("Buf.String() = %q, want %q", got, want)
	}
	if want := (buffer.Position{Line: 0, Col: 0}); e.Cursor != want {
		t.Fatalf("Cursor = %+v, want %+v", e.Cursor, want)
	}
	if e.HasSelection() {
		t.Error("expected selection to be cleared")
	}
}

func TestSelectedText(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		anchor    buffer.Position
		cursor    buffer.Position
		hasAnchor bool
		want      string
	}{
		{
			name:      "no selection returns empty string",
			content:   "hello world",
			hasAnchor: false,
			want:      "",
		},
		{
			name:      "forward selection within a line",
			content:   "hello world",
			anchor:    buffer.Position{Line: 0, Col: 0},
			cursor:    buffer.Position{Line: 0, Col: 5},
			hasAnchor: true,
			want:      "hello",
		},
		{
			name:      "selection spanning multiple lines",
			content:   "hello\nworld",
			anchor:    buffer.Position{Line: 0, Col: 3},
			cursor:    buffer.Position{Line: 1, Col: 3},
			hasAnchor: true,
			want:      "lo\nwor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()
			e.Buf = buffer.New(tt.content)
			if tt.hasAnchor {
				anchor := tt.anchor
				e.Anchor = &anchor
			}
			e.Cursor = tt.cursor

			if got := e.SelectedText(); got != tt.want {
				t.Errorf("SelectedText() = %q, want %q", got, tt.want)
			}
		})
	}
}
