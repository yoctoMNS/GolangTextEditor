package editor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yoctoMNS/GolangTextEditor/internal/buffer"
)

// TestTypingAndNewline exercises a realistic sequence of edits (as opposed
// to the table-driven tests below, which check one independent action at a
// time) to make sure Editor's methods compose correctly across calls.
func TestTypingAndNewline(t *testing.T) {
	e := New()
	for _, r := range "hi" {
		e.InsertRune(r)
	}
	e.InsertNewline()
	for _, r := range "there" {
		e.InsertRune(r)
	}
	if got, want := e.Buf.String(), "hi\nthere"; got != want {
		t.Fatalf("Buf.String() = %q, want %q", got, want)
	}
	if want := (buffer.Position{Line: 1, Col: 5}); e.Cursor != want {
		t.Fatalf("Cursor = %+v, want %+v", e.Cursor, want)
	}
	if !e.Modified {
		t.Fatal("expected Modified to be true after edits")
	}
}

func TestBackspace(t *testing.T) {
	tests := []struct {
		name         string
		initial      string
		cursor       buffer.Position
		wantStr      string
		wantCursor   buffer.Position
		wantModified bool
	}{
		{
			name:         "within a line",
			initial:      "hello",
			cursor:       buffer.Position{Line: 0, Col: 5},
			wantStr:      "hell",
			wantCursor:   buffer.Position{Line: 0, Col: 4},
			wantModified: true,
		},
		{
			name:         "at start of line joins with previous line",
			initial:      "hello\nworld",
			cursor:       buffer.Position{Line: 1, Col: 0},
			wantStr:      "helloworld",
			wantCursor:   buffer.Position{Line: 0, Col: 5},
			wantModified: true,
		},
		{
			name:         "at start of buffer is a no-op and leaves Modified false",
			initial:      "hello",
			cursor:       buffer.Position{Line: 0, Col: 0},
			wantStr:      "hello",
			wantCursor:   buffer.Position{Line: 0, Col: 0},
			wantModified: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()
			e.Buf = buffer.New(tt.initial)
			e.Cursor = tt.cursor

			e.Backspace()

			if got := e.Buf.String(); got != tt.wantStr {
				t.Errorf("Buf.String() = %q, want %q", got, tt.wantStr)
			}
			if e.Cursor != tt.wantCursor {
				t.Errorf("Cursor = %+v, want %+v", e.Cursor, tt.wantCursor)
			}
			if e.Modified != tt.wantModified {
				t.Errorf("Modified = %v, want %v", e.Modified, tt.wantModified)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name         string
		initial      string
		cursor       buffer.Position
		wantStr      string
		wantModified bool
	}{
		{
			name:         "within a line",
			initial:      "hello",
			cursor:       buffer.Position{Line: 0, Col: 0},
			wantStr:      "ello",
			wantModified: true,
		},
		{
			name:         "at end of line joins with next line",
			initial:      "hello\nworld",
			cursor:       buffer.Position{Line: 0, Col: 5},
			wantStr:      "helloworld",
			wantModified: true,
		},
		{
			name:         "at end of buffer is a no-op and leaves Modified false",
			initial:      "hello",
			cursor:       buffer.Position{Line: 0, Col: 5},
			wantStr:      "hello",
			wantModified: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()
			e.Buf = buffer.New(tt.initial)
			e.Cursor = tt.cursor

			e.Delete()

			if got := e.Buf.String(); got != tt.wantStr {
				t.Errorf("Buf.String() = %q, want %q", got, tt.wantStr)
			}
			if e.Modified != tt.wantModified {
				t.Errorf("Modified = %v, want %v", e.Modified, tt.wantModified)
			}
		})
	}
}

func TestMovement(t *testing.T) {
	tests := []struct {
		name  string
		buf   string
		start buffer.Position
		move  func(*Editor)
		want  buffer.Position
	}{
		{
			name:  "MoveLeft at start of buffer is a no-op",
			buf:   "ab\nc",
			start: buffer.Position{Line: 0, Col: 0},
			move:  (*Editor).MoveLeft,
			want:  buffer.Position{Line: 0, Col: 0},
		},
		{
			name:  "MoveLeft at start of line wraps to end of previous line",
			buf:   "ab\nc",
			start: buffer.Position{Line: 1, Col: 0},
			move:  (*Editor).MoveLeft,
			want:  buffer.Position{Line: 0, Col: 2},
		},
		{
			name:  "MoveRight at end of line wraps to start of next line",
			buf:   "ab\nc",
			start: buffer.Position{Line: 0, Col: 2},
			move:  (*Editor).MoveRight,
			want:  buffer.Position{Line: 1, Col: 0},
		},
		{
			name:  "MoveRight at end of buffer is a no-op",
			buf:   "ab\nc",
			start: buffer.Position{Line: 1, Col: 1},
			move:  (*Editor).MoveRight,
			want:  buffer.Position{Line: 1, Col: 1},
		},
		{
			name:  "MoveUp at first line is a no-op",
			buf:   "ab\nc",
			start: buffer.Position{Line: 0, Col: 1},
			move:  (*Editor).MoveUp,
			want:  buffer.Position{Line: 0, Col: 1},
		},
		{
			name:  "MoveDown at last line clamps the column to the shorter line",
			buf:   "ab\nc",
			start: buffer.Position{Line: 1, Col: 0},
			move:  (*Editor).MoveDown,
			want:  buffer.Position{Line: 1, Col: 0},
		},
		{
			name:  "MoveDown clamps column when the next line is shorter",
			buf:   "ab\nc",
			start: buffer.Position{Line: 0, Col: 2},
			move:  (*Editor).MoveDown,
			want:  buffer.Position{Line: 1, Col: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()
			e.Buf = buffer.New(tt.buf)
			e.Cursor = tt.start

			tt.move(e)

			if e.Cursor != tt.want {
				t.Errorf("Cursor = %+v, want %+v", e.Cursor, tt.want)
			}
		})
	}
}

func TestHomeAndEnd(t *testing.T) {
	tests := []struct {
		name    string
		start   buffer.Position
		move    func(*Editor)
		wantCol int
	}{
		{
			name:    "End moves to the last column of the line",
			start:   buffer.Position{Line: 0, Col: 5},
			move:    (*Editor).End,
			wantCol: 11,
		},
		{
			name:    "Home moves to the first column of the line",
			start:   buffer.Position{Line: 0, Col: 5},
			move:    (*Editor).Home,
			wantCol: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()
			e.Buf = buffer.New("hello world")
			e.Cursor = tt.start

			tt.move(e)

			if e.Cursor.Col != tt.wantCol {
				t.Errorf("Cursor.Col = %d, want %d", e.Cursor.Col, tt.wantCol)
			}
		})
	}
}

func TestLoadSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	content := "line one\nline two\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	e := New()
	if err := e.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got, want := e.Buf.Line(0), "line one"; got != want {
		t.Fatalf("Line(0) = %q, want %q", got, want)
	}
	if e.Modified {
		t.Fatal("expected Modified to be false right after Load")
	}

	e.InsertRune('!')
	if !e.Modified {
		t.Fatal("expected Modified to be true after edit")
	}

	if err := e.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if e.Modified {
		t.Fatal("expected Modified to be false right after Save")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if got, want := string(data), "!line one\nline two\n"; got != want {
		t.Fatalf("saved content = %q, want %q", got, want)
	}
}

func TestSaveWithoutPathFails(t *testing.T) {
	e := New()
	if err := e.Save(); err == nil {
		t.Fatal("expected Save() to fail when Path is empty")
	}
}
