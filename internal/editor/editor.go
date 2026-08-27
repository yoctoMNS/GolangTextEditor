// Package editor implements editing commands and file I/O on top of a buffer.Buffer.
// It knows nothing about rendering or input devices: any front-end (ebiten ,raylib-go, go-sdl2, go-gl, a terminal UI, tests, ...)
// drives it purely through the methods below.
package editor

import (
	"os"

	"github.com/yoctoMNS/GolangTextEditor/internal/buffer"
)

// Editor holds a buffer, the caret position and file metadata.
type Editor struct {
	Buf      *buffer.Buffer
	Cursor   buffer.Position
	Path     string
	Modified bool
}

// New creates an empty Editor with a single empty line.
func New() *Editor {
	return &Editor{
		Buf: buffer.New(""),
	}
}

// Load reads path into a fresh buffer and resets the caret to the start.
func (e *Editor) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	e.Buf = buffer.New(string(data))
	e.Cursor = buffer.Position{}
	e.Path = path
	e.Modified = false
	return nil
}

// Save writes the buffer to Path.
// It returns an error if Path is empty; use SaveAs for the first save of a new file.
func (e *Editor) Save() error {
	if e.Path == "" {
		return os.ErrInvalid
	}
	return e.SaveAs(e.Path)
}

// SaveAs writes the buffer to path and remembers it as the current Path.
func (e *Editor) SaveAs(path string) error {
	if err := os.WriteFile(path, []byte(e.Buf.String()), 0o644); err != nil {
		return err
	}
	e.Path = path
	e.Modified = false
	return nil
}

// InsertRune inserts r at the caret and advances the caret past it.
func (e *Editor) InsertRune(r rune) {
	e.Cursor = e.Buf.InsertRune(e.Cursor, r)
	e.Modified = true
}

// InsertNewline splits the current line at the caret.
func (e *Editor) InsertNewline() {
	e.Cursor = e.Buf.InsertNewline(e.Cursor)
	e.Modified = true
}

// Backspace deletes the rune before the caret, joining lines if needed.
func (e *Editor) Backspace() {
	if pos, ok := e.Buf.DeleteBackward(e.Cursor); ok {
		e.Cursor = pos
		e.Modified = true
	}
}

// Delete deletes the rune at the caret, joining lines if needed.
func (e *Editor) Delete() {
	if e.Buf.DeleteForward(e.Cursor) {
		e.Modified = true
	}
}

// MoveLeft moves the caret one rune to the left,
// wrapping to the end of the previous line at the start of a line.
func (e *Editor) MoveLeft() {
	if e.Cursor.Col > 0 {
		e.Cursor.Col--
		return
	}
	if e.Cursor.Line > 0 {
		e.Cursor.Line--
		e.Cursor.Col = e.Buf.LineLen(e.Cursor.Line)
	}
}

// MoveRight moves the caret one rune to the right,
// wrapping to the start of the next line at the end of a line.
func (e *Editor) MoveRight() {
	if e.Cursor.Col < e.Buf.LineLen(e.Cursor.Line) {
		e.Cursor.Col++
		return
	}
	if e.Cursor.Line < e.Buf.LineCount()-1 {
		e.Cursor.Line++
		e.Cursor.Col = 0
	}
}

// MoveUp moves the caret one line up,
// clamping the column to the target line's length.
func (e *Editor) MoveUp() {
	e.Cursor = e.Buf.Clamp(buffer.Position{
		Line: e.Cursor.Line - 1,
		Col:  e.Cursor.Col,
	})
}

// MoveDown moves the caret one line down,
// clamping the column to the target line's length.
func (e *Editor) MoveDown() {
	e.Cursor = e.Buf.Clamp(buffer.Position{
		Line: e.Cursor.Line + 1,
		Col:  e.Cursor.Col,
	})
}

// Home moves the caret to the start of the current line.
func (e *Editor) Home() {
	e.Cursor.Col = 0
}

// End moves the caret to the end of the current line.
func (e *Editor) End() {
	e.Cursor.Col = e.Buf.LineLen(e.Cursor.Line)
}
