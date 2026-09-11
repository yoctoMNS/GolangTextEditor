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

func TestPositionAt(t *testing.T) {
	// originX=40, originY=4, scrollLine=0, charWidth=7, lineHeight=16
	// for every case unless noted otherwise.
	tests := []struct {
		name       string
		x, y       int
		scrollLine int
		wantLine   int
		wantCol    int
	}{
		{
			name: "click exactly at the text origin targets line 0, col 0",
			x:    40, y: 4,
			wantLine: 0, wantCol: 0,
		},
		{
			name: "click one full cell right and down targets line 1, col 1",
			x:    47, y: 20,
			wantLine: 1, wantCol: 1,
		},
		{
			name: "click mid-cell rounds down to that cell, not the next one",
			x:    44, y: 10, // 4px into col 0's 7px cell; 6px into row 0's 16px cell
			wantLine: 0, wantCol: 0,
		},
		{
			name: "click left of the text origin (e.g. in the gutter) gives a negative column",
			x:    10, y: 4,
			wantLine: 0, wantCol: -5, // (10-40)/7 floors to -5, not -4
		},
		{
			name: "click above the text origin gives a negative line",
			x:    40, y: -10,
			wantLine: -1, wantCol: 0, // (-10-4)/16 floors to -1, not 0
		},
		{
			name: "scrollLine offsets the resulting line",
			x:    40, y: 20,
			scrollLine: 5,
			wantLine:   6, // scrolled 5 + 1 row down from the click
			wantCol:    0,
		},
	}

	const originX, originY, charWidth, lineHeight = 40, 4, 7, 16

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLine, gotCol := PositionAt(tt.x, tt.y, originX, originY, tt.scrollLine, charWidth, lineHeight)
			if gotLine != tt.wantLine || gotCol != tt.wantCol {
				t.Errorf("PositionAt(%d, %d, scrollLine=%d) = (%d, %d), want (%d, %d)",
					tt.x, tt.y, tt.scrollLine, gotLine, gotCol, tt.wantLine, tt.wantCol)
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
