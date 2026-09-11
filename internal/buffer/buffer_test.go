package buffer

import "testing"

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		wantLineCount int
		wantString    string
	}{
		{
			name:          "single line",
			content:       "hello",
			wantLineCount: 1,
			wantString:    "hello",
		},
		{
			name:          "multiple lines",
			content:       "hello\nworld",
			wantLineCount: 2,
			wantString:    "hello\nworld",
		},
		{
			name:          "CRLF line endings are normalized to LF",
			content:       "hello\r\nworld\r\n",
			wantLineCount: 3,
			wantString:    "hello\nworld\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.content)
			if got := b.LineCount(); got != tt.wantLineCount {
				t.Errorf("LineCount() = %d, want %d", got, tt.wantLineCount)
			}
			if got := b.String(); got != tt.wantString {
				t.Errorf("String() = %q, want %q", got, tt.wantString)
			}
		})
	}
}

func TestInsertRune(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		pos     Position
		r       rune
		wantStr string
		wantPos Position
	}{
		{
			name:    "middle of line",
			initial: "helloworld",
			pos:     Position{Line: 0, Col: 5},
			r:       ' ',
			wantStr: "hello world",
			wantPos: Position{Line: 0, Col: 6},
		},
		{
			name:    "start of line",
			initial: "world",
			pos:     Position{Line: 0, Col: 0},
			r:       'h',
			wantStr: "hworld",
			wantPos: Position{Line: 0, Col: 1},
		},
		{
			name:    "end of line",
			initial: "hello",
			pos:     Position{Line: 0, Col: 5},
			r:       '!',
			wantStr: "hello!",
			wantPos: Position{Line: 0, Col: 6},
		},
		{
			name:    "multi-byte rune keeps rune-based column indexing",
			initial: "こんにちは",
			pos:     Position{Line: 0, Col: 2},
			r:       '!',
			wantStr: "こん!にちは",
			wantPos: Position{Line: 0, Col: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.initial)
			gotPos := b.InsertRune(tt.pos, tt.r)
			if got := b.String(); got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}
			if gotPos != tt.wantPos {
				t.Errorf("returned pos = %+v, want %+v", gotPos, tt.wantPos)
			}
		})
	}
}

func TestInsertNewline(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		pos     Position
		wantStr string
		wantPos Position
	}{
		{
			name:    "splits a line in the middle",
			initial: "helloworld",
			pos:     Position{Line: 0, Col: 5},
			wantStr: "hello\nworld",
			wantPos: Position{Line: 1, Col: 0},
		},
		{
			name:    "at start of line leaves an empty line before",
			initial: "hello",
			pos:     Position{Line: 0, Col: 0},
			wantStr: "\nhello",
			wantPos: Position{Line: 1, Col: 0},
		},
		{
			name:    "at end of line leaves an empty line after",
			initial: "hello",
			pos:     Position{Line: 0, Col: 5},
			wantStr: "hello\n",
			wantPos: Position{Line: 1, Col: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.initial)
			gotPos := b.InsertNewline(tt.pos)
			if got := b.String(); got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}
			if gotPos != tt.wantPos {
				t.Errorf("returned pos = %+v, want %+v", gotPos, tt.wantPos)
			}
		})
	}
}

func TestDeleteBackward(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		pos     Position
		wantOK  bool
		wantStr string
		wantPos Position
	}{
		{
			name:    "within a line",
			initial: "hello",
			pos:     Position{Line: 0, Col: 5},
			wantOK:  true,
			wantStr: "hell",
			wantPos: Position{Line: 0, Col: 4},
		},
		{
			name:    "at start of line joins with previous line",
			initial: "hello\nworld",
			pos:     Position{Line: 1, Col: 0},
			wantOK:  true,
			wantStr: "helloworld",
			wantPos: Position{Line: 0, Col: 5},
		},
		{
			name:    "at start of buffer is a no-op",
			initial: "hello",
			pos:     Position{Line: 0, Col: 0},
			wantOK:  false,
			wantStr: "hello",
			wantPos: Position{Line: 0, Col: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.initial)
			gotPos, gotOK := b.DeleteBackward(tt.pos)
			if gotOK != tt.wantOK {
				t.Errorf("ok = %v, want %v", gotOK, tt.wantOK)
			}
			if got := b.String(); got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}
			if gotPos != tt.wantPos {
				t.Errorf("returned pos = %+v, want %+v", gotPos, tt.wantPos)
			}
		})
	}
}

func TestDeleteForward(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		pos     Position
		wantOK  bool
		wantStr string
	}{
		{
			name:    "within a line",
			initial: "hello",
			pos:     Position{Line: 0, Col: 0},
			wantOK:  true,
			wantStr: "ello",
		},
		{
			name:    "at end of line joins with next line",
			initial: "hello\nworld",
			pos:     Position{Line: 0, Col: 5},
			wantOK:  true,
			wantStr: "helloworld",
		},
		{
			name:    "at end of buffer is a no-op",
			initial: "hello",
			pos:     Position{Line: 0, Col: 5},
			wantOK:  false,
			wantStr: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.initial)
			gotOK := b.DeleteForward(tt.pos)
			if gotOK != tt.wantOK {
				t.Errorf("ok = %v, want %v", gotOK, tt.wantOK)
			}
			if got := b.String(); got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

func TestTextRange(t *testing.T) {
	tests := []struct {
		name    string
		content string
		start   Position
		end     Position
		want    string
	}{
		{
			name:    "within a single line",
			content: "hello world",
			start:   Position{Line: 0, Col: 6},
			end:     Position{Line: 0, Col: 11},
			want:    "world",
		},
		{
			name:    "empty range returns empty string",
			content: "hello",
			start:   Position{Line: 0, Col: 2},
			end:     Position{Line: 0, Col: 2},
			want:    "",
		},
		{
			name:    "spans two adjacent lines",
			content: "hello\nworld",
			start:   Position{Line: 0, Col: 3},
			end:     Position{Line: 1, Col: 3},
			want:    "lo\nwor",
		},
		{
			name:    "spans three lines including a full middle line",
			content: "one\ntwo\nthree",
			start:   Position{Line: 0, Col: 1},
			end:     Position{Line: 2, Col: 2},
			want:    "ne\ntwo\nth",
		},
		{
			name:    "whole buffer matches String()",
			content: "a\nbb\nccc",
			start:   Position{Line: 0, Col: 0},
			end:     Position{Line: 2, Col: 3},
			want:    "a\nbb\nccc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.content)
			if got := b.TextRange(tt.start, tt.end); got != tt.want {
				t.Errorf("TextRange(%+v, %+v) = %q, want %q", tt.start, tt.end, got, tt.want)
			}
		})
	}
}

func TestDeleteRange(t *testing.T) {
	tests := []struct {
		name    string
		content string
		start   Position
		end     Position
		want    string
	}{
		{
			name:    "within a single line",
			content: "hello world",
			start:   Position{Line: 0, Col: 5},
			end:     Position{Line: 0, Col: 11},
			want:    "hello",
		},
		{
			name:    "empty range is a no-op",
			content: "hello",
			start:   Position{Line: 0, Col: 2},
			end:     Position{Line: 0, Col: 2},
			want:    "hello",
		},
		{
			name:    "spans two adjacent lines",
			content: "hello\nworld",
			start:   Position{Line: 0, Col: 3},
			end:     Position{Line: 1, Col: 3},
			want:    "helld",
		},
		{
			name:    "spans three lines, removing a full middle line",
			content: "one\ntwo\nthree",
			start:   Position{Line: 0, Col: 1},
			end:     Position{Line: 2, Col: 2},
			want:    "oree",
		},
		{
			name:    "deleting the whole buffer leaves one empty line",
			content: "a\nbb\nccc",
			start:   Position{Line: 0, Col: 0},
			end:     Position{Line: 2, Col: 3},
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.content)
			b.DeleteRange(tt.start, tt.end)
			if got := b.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	b := New("hi\nworld")

	tests := []struct {
		name string
		pos  Position
		want Position
	}{
		{
			name: "negative line and column clamp to buffer start",
			pos:  Position{Line: -1, Col: -1},
			want: Position{Line: 0, Col: 0},
		},
		{
			name: "column past end of line clamps to line length",
			pos:  Position{Line: 0, Col: 99},
			want: Position{Line: 0, Col: 2},
		},
		{
			name: "line past end of buffer clamps to last line",
			pos:  Position{Line: 99, Col: 0},
			want: Position{Line: 1, Col: 0},
		},
		{
			name: "already valid position is unchanged",
			pos:  Position{Line: 1, Col: 3},
			want: Position{Line: 1, Col: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b.Clamp(tt.pos); got != tt.want {
				t.Errorf("Clamp(%+v) = %+v, want %+v", tt.pos, got, tt.want)
			}
		})
	}
}
