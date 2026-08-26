package buffer

import "testing"

func TestNewAndString(t *testing.T) {
	b := New("hello\nworld")
	if got, want := b.LineCount(), 2; got != want {
		t.Fatalf("LineCount() = %d, want %d", got, want)
	}
	if got, want := b.String(), "hello\nworld"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestNewStripsCR(t *testing.T) {
	b := New("hello\r\nworld\r\n")
	if got, want := b.Line(0), "hello"; got != want {
		t.Fatalf("Line(0) = %q, want %q", got, want)
	}
	if got, want := b.LineCount(), 3; got != want {
		t.Fatalf("LineCount() = %d, want %d", got, want)
	}
}

func TestInsertRune(t *testing.T) {
	b := New("helloworld")
	pos := b.InsertRune(Position{Line: 0, Col: 5}, ' ')
	if got, want := b.Line(0), "hello world"; got != want {
		t.Fatalf("Line(0) = %q, want %q", got, want)
	}
	if want := (Position{Line: 0, Col: 6}); pos != want {
		t.Fatalf("returned pos = %+v, want %+v", pos, want)
	}
}

func TestInsertNewline(t *testing.T) {
	b := New("helloworld")
	pos := b.InsertNewline(Position{Line: 0, Col: 5})
	if got, want := b.LineCount(), 2; got != want {
		t.Fatalf("LineCount() = %d, want %d", got, want)
	}
	if got, want := b.Line(0), "hello"; got != want {
		t.Fatalf("Line(0) = %q, want %q", got, want)
	}
	if got, want := b.Line(1), "world"; got != want {
		t.Fatalf("Line(1) = %q, want %q", got, want)
	}
	if want := (Position{Line: 1, Col: 0}); pos != want {
		t.Fatalf("returned pos = %+v, want %+v", pos, want)
	}
}

func TestDeleteBackwardWithinLine(t *testing.T) {
	b := New("hello")
	pos, ok := b.DeleteBackward(Position{Line: 0, Col: 5})
	if !ok {
		t.Fatal("expected deletion to succeed")
	}
	if got, want := b.Line(0), "hell"; got != want {
		t.Fatalf("Line(0) = %q, want %q", got, want)
	}
	if want := (Position{Line: 0, Col: 4}); pos != want {
		t.Fatalf("returned pos = %+v, want %+v", pos, want)
	}
}

func TestDeleteBackwardJoinsLines(t *testing.T) {
	b := New("hello\nworld")
	pos, ok := b.DeleteBackward(Position{Line: 1, Col: 0})
	if !ok {
		t.Fatal("expected deletion to succeed")
	}
	if got, want := b.LineCount(), 1; got != want {
		t.Fatalf("LineCount() = %d, want %d", got, want)
	}
	if got, want := b.Line(0), "helloworld"; got != want {
		t.Fatalf("Line(0) = %q, want %q", got, want)
	}
	if want := (Position{Line: 0, Col: 5}); pos != want {
		t.Fatalf("returned pos = %+v, want %+v", pos, want)
	}
}

func TestDeleteBackwardAtStartOfBuffer(t *testing.T) {
	b := New("hello")
	pos, ok := b.DeleteBackward(Position{Line: 0, Col: 0})
	if ok {
		t.Fatal("expected deletion to fail at start of buffer")
	}
	if want := (Position{Line: 0, Col: 0}); pos != want {
		t.Fatalf("returned pos = %+v, want %+v", pos, want)
	}
}

func TestDeleteForwardWithinLine(t *testing.T) {
	b := New("hello")
	if ok := b.DeleteForward(Position{Line: 0, Col: 0}); !ok {
		t.Fatal("expected deletion to succeed")
	}
	if got, want := b.Line(0), "ello"; got != want {
		t.Fatalf("Line(0) = %q, want %q", got, want)
	}
}

func TestDeleteForwardJoinsLines(t *testing.T) {
	b := New("hello\nworld")
	if ok := b.DeleteForward(Position{Line: 0, Col: 5}); !ok {
		t.Fatal("expected deletion to succeed")
	}
	if got, want := b.LineCount(), 1; got != want {
		t.Fatalf("LineCount() = %d, want %d", got, want)
	}
	if got, want := b.Line(0), "helloworld"; got != want {
		t.Fatalf("Line(0) = %q, want %q", got, want)
	}
}

func TestDeleteForwardAtEndOfBuffer(t *testing.T) {
	b := New("hello")
	if ok := b.DeleteForward(Position{Line: 0, Col: 5}); ok {
		t.Fatal("expected deletion to fail at end of buffer")
	}
}

func TestClamp(t *testing.T) {
	b := New("hi\nworld")
	if got, want := b.Clamp(Position{Line: -1, Col: -1}), (Position{Line: 0, Col: 0}); got != want {
		t.Fatalf("Clamp = %+v, want %+v", got, want)
	}
	if got, want := b.Clamp(Position{Line: 0, Col: 99}), (Position{Line: 0, Col: 2}); got != want {
		t.Fatalf("Clamp = %+v, want %+v", got, want)
	}
	if got, want := b.Clamp(Position{Line: 99, Col: 0}), (Position{Line: 1, Col: 0}); got != want {
		t.Fatalf("Clamp = %+v, want %+v", got, want)
	}
}

func TestUnicodeRunes(t *testing.T) {
	b := New("こんにちは")
	if got, want := b.LineLen(0), 5; got != want {
		t.Fatalf("LineLen(0) = %d, want %d", got, want)
	}
	pos := b.InsertRune(Position{Line: 0, Col: 2}, '!')
	if got, want := b.Line(0), "こん!にちは"; got != want {
		t.Fatalf("Line(0) = %q, want %q", got, want)
	}
	if want := (Position{Line: 0, Col: 3}); pos != want {
		t.Fatalf("returned pos = %+v, want %+v", pos, want)
	}
}
