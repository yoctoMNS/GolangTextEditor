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
