// Package buffer implements the in-memory text storage for the editor.
//
// It has no dependency on any rendering or windowing library so it can be
// unit tested headlessly and reused by any future front-end (ebiten,
// raylib-go, go-sdl2, go-gl, ...).
package buffer

import "strings"

// Position identifies a caret location inside a Buffer.
// Col is a rune index into the line, in the range [0, LineLen(Line)].
type Position struct {
	Line int
	Col  int
}

// Buffer holds the text content as a slice of lines, each line a slice of
// runes. A line-oriented representation keeps the implementation simple
// while still giving O(line length) edits, which is sufficient until
// profiling shows a rope/piece-table is required (see CLAUDE.md).
type Buffer struct {
	lines [][]rune
}

// New creates a Buffer from the given text. Lines are split on "\n"; a
// trailing "\r" on each line (CRLF files) is stripped.
func New(content string) *Buffer {
	raw := strings.Split(content, "\n")
	lines := make([][]rune, len(raw))
	for i, l := range raw {
		l = strings.TrimSuffix(l, "\r")
		lines[i] = []rune(l)
	}
	return &Buffer{lines: lines}
}

// LineCount returns the number of lines in the buffer. An empty buffer
// still has one (empty) line, matching common editor conventions.
func (b *Buffer) LineCount() int {
	return len(b.lines)
}

// Line returns the content of line i as a string.
func (b *Buffer) Line(i int) string {
	return string(b.lines[i])
}

// LineLen returns the number of runes in line i.
func (b *Buffer) LineLen(i int) int {
	return len(b.lines[i])
}

// String returns the full buffer content joined with "\n".
func (b *Buffer) String() string {
	var sb strings.Builder
	for i, l := range b.lines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(string(l))
	}
	return sb.String()
}

// Clamp constrains pos so it always refers to a valid location in the
// buffer, clamping the line into range first and then the column.
func (b *Buffer) Clamp(pos Position) Position {
	if pos.Line < 0 {
		pos.Line = 0
	}
	if pos.Line >= b.LineCount() {
		pos.Line = b.LineCount() - 1
	}
	if pos.Col < 0 {
		pos.Col = 0
	}
	if max := b.LineLen(pos.Line); pos.Col > max {
		pos.Col = max
	}
	return pos
}

// InsertRune inserts r at pos and returns the caret position immediately
// after the inserted rune.
func (b *Buffer) InsertRune(pos Position, r rune) Position {
	line := b.lines[pos.Line]
	line = append(line, 0)
	copy(line[pos.Col+1:], line[pos.Col:])
	line[pos.Col] = r
	b.lines[pos.Line] = line
	return Position{Line: pos.Line, Col: pos.Col + 1}
}

// InsertNewline splits the line at pos into two lines and returns the
// caret position at the start of the new (second) line.
func (b *Buffer) InsertNewline(pos Position) Position {
	line := b.lines[pos.Line]
	before := append([]rune{}, line[:pos.Col]...)
	after := append([]rune{}, line[pos.Col:]...)

	b.lines = append(b.lines, nil)
	copy(b.lines[pos.Line+2:], b.lines[pos.Line+1:])
	b.lines[pos.Line] = before
	b.lines[pos.Line+1] = after

	return Position{Line: pos.Line + 1, Col: 0}
}

// DeleteBackward deletes the rune immediately before pos (Backspace
// behavior). If pos is at the start of a line, it joins the line with the
// previous one instead. It returns the resulting caret position and
// whether anything was deleted (false at the very start of the buffer).
func (b *Buffer) DeleteBackward(pos Position) (Position, bool) {
	if pos.Col > 0 {
		line := b.lines[pos.Line]
		copy(line[pos.Col-1:], line[pos.Col:])
		b.lines[pos.Line] = line[:len(line)-1]
		return Position{Line: pos.Line, Col: pos.Col - 1}, true
	}
	if pos.Line > 0 {
		prevLen := b.LineLen(pos.Line - 1)
		b.lines[pos.Line-1] = append(b.lines[pos.Line-1], b.lines[pos.Line]...)
		b.lines = append(b.lines[:pos.Line], b.lines[pos.Line+1:]...)
		return Position{Line: pos.Line - 1, Col: prevLen}, true
	}
	return pos, false
}

// DeleteForward deletes the rune at pos (Delete key behavior). If pos is
// at the end of a line, it joins the next line into the current one
// instead. It returns whether anything was deleted (false at the very end
// of the buffer).
func (b *Buffer) DeleteForward(pos Position) bool {
	if pos.Col < b.LineLen(pos.Line) {
		line := b.lines[pos.Line]
		copy(line[pos.Col:], line[pos.Col+1:])
		b.lines[pos.Line] = line[:len(line)-1]
		return true
	}
	if pos.Line < b.LineCount()-1 {
		b.lines[pos.Line] = append(b.lines[pos.Line], b.lines[pos.Line+1]...)
		b.lines = append(b.lines[:pos.Line+1], b.lines[pos.Line+2:]...)
		return true
	}
	return false
}
