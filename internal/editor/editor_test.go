package editor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yoctoMNS/GolangTextEditor/internal/buffer"
)

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

func TestBackspaceJoinsLines(t *testing.T) {
	e := New()
	e.Buf = buffer.New("hello\nworld")
	e.Cursor = buffer.Position{Line: 1, Col: 0}
	e.Backspace()
	if got, want := e.Buf.String(), "helloworld"; got != want {
		t.Fatalf("Buf.String() = %q, want %q", got, want)
	}
	if want := (buffer.Position{Line: 0, Col: 5}); e.Cursor != want {
		t.Fatalf("Cursor = %+v, want %+v", e.Cursor, want)
	}
}

func TestMovementClampsAtBufferEdges(t *testing.T) {
	e := New()
	e.Buf = buffer.New("ab\nc")
	e.Cursor = buffer.Position{Line: 0, Col: 0}

	e.MoveLeft() // already at start, no-op
	if want := (buffer.Position{Line: 0, Col: 0}); e.Cursor != want {
		t.Fatalf("Cursor = %+v, want %+v", e.Cursor, want)
	}

	e.MoveUp() // already at first line, no-op
	if want := (buffer.Position{Line: 0, Col: 0}); e.Cursor != want {
		t.Fatalf("Cursor = %+v, want %+v", e.Cursor, want)
	}

	e.Cursor = buffer.Position{Line: 0, Col: 2}
	e.MoveRight() // wraps to next line
	if want := (buffer.Position{Line: 1, Col: 0}); e.Cursor != want {
		t.Fatalf("Cursor = %+v, want %+v", e.Cursor, want)
	}

	e.MoveDown() // already at last line, no-op
	if want := (buffer.Position{Line: 1, Col: 0}); e.Cursor != want {
		t.Fatalf("Cursor = %+v, want %+v", e.Cursor, want)
	}
}

func TestHomeAndEnd(t *testing.T) {
	e := New()
	e.Buf = buffer.New("hello world")
	e.Cursor = buffer.Position{Line: 0, Col: 5}
	e.End()
	if want := 11; e.Cursor.Col != want {
		t.Fatalf("Cursor.Col after End() = %d, want %d", e.Cursor.Col, want)
	}
	e.Home()
	if want := 0; e.Cursor.Col != want {
		t.Fatalf("Cursor.Col after Home() = %d, want %d", e.Cursor.Col, want)
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
