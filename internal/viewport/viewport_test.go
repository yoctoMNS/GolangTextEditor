package viewport

import "testing"

func TestLineCount(t *testing.T) {
	tests := []struct {
		name            string
		availableHeight int
		lineHeight      int
		want            int
	}{
		{
			name:            "exact multiple",
			availableHeight: 160,
			lineHeight:      16,
			want:            10,
		},
		{
			name:            "rounds down for a partial trailing line",
			availableHeight: 165,
			lineHeight:      16,
			want:            10,
		},
		{
			name:            "clamps to 1 when the area is smaller than one line",
			availableHeight: 5,
			lineHeight:      16,
			want:            1,
		},
		{
			name:            "clamps to 1 for a negative or zero area",
			availableHeight: -20,
			lineHeight:      16,
			want:            1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LineCount(tt.availableHeight, tt.lineHeight); got != tt.want {
				t.Errorf("LineCount(%d, %d) = %d, want %d", tt.availableHeight, tt.lineHeight, got, tt.want)
			}
		})
	}
}

func TestGutterWidth(t *testing.T) {
	tests := []struct {
		name      string
		lineCount int
		charWidth int
		paddingX  int
		want      int
	}{
		{
			name:      "short file still gets a 2-digit-wide gutter",
			lineCount: 5,
			charWidth: 7,
			paddingX:  4,
			want:      2*7 + 4*2, // = 22
		},
		{
			name:      "3-digit file gets a wider gutter",
			lineCount: 100,
			charWidth: 7,
			paddingX:  4,
			want:      3*7 + 4*2, // = 29
		},
		{
			name:      "digit count follows the largest line number, not just magnitude",
			lineCount: 999,
			charWidth: 7,
			paddingX:  4,
			want:      3*7 + 4*2, // = 29
		},
		{
			name:      "4-digit file",
			lineCount: 1000,
			charWidth: 7,
			paddingX:  4,
			want:      4*7 + 4*2, // = 36
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GutterWidth(tt.lineCount, tt.charWidth, tt.paddingX)
			if got != tt.want {
				t.Errorf("GutterWidth(%d, %d, %d) = %d, want %d",
					tt.lineCount, tt.charWidth, tt.paddingX, got, tt.want)
			}
		})
	}
}

func TestClampScroll(t *testing.T) {
	tests := []struct {
		name         string
		scroll       int
		cursorLine   int
		visibleLines int
		lineCount    int
		want         int
	}{
		{
			name:         "cursor already visible does not scroll",
			scroll:       0,
			cursorLine:   3,
			visibleLines: 10,
			lineCount:    20,
			want:         0,
		},
		{
			name:         "cursor above viewport scrolls up to cursor",
			scroll:       10,
			cursorLine:   4,
			visibleLines: 10,
			lineCount:    20,
			want:         4,
		},
		{
			name:         "cursor below viewport scrolls down by the minimum amount",
			scroll:       0,
			cursorLine:   12,
			visibleLines: 10,
			lineCount:    20,
			want:         3,
		},
		{
			name:         "does not scroll past the last page when buffer is short",
			scroll:       0,
			cursorLine:   19,
			visibleLines: 10,
			lineCount:    20,
			want:         10,
		},
		{
			name:         "buffer shorter than viewport never scrolls",
			scroll:       5,
			cursorLine:   2,
			visibleLines: 10,
			lineCount:    3,
			want:         0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClampScroll(tt.scroll, tt.cursorLine, tt.visibleLines, tt.lineCount)
			if got != tt.want {
				t.Errorf("ClampScroll(%d, %d, %d, %d) = %d, want %d",
					tt.scroll, tt.cursorLine, tt.visibleLines, tt.lineCount, got, tt.want)
			}
		})
	}
}
