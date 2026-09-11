// Package app wires the UI-agnostic editor core to a concrete rendering
// and input backend. This is the only package in the module that may
// import a graphics library; internal/buffer and internal/editor stay
// backend-free so a different backend (raylib-go, go-sdl2, go-gl, ...)
// can be dropped in later behind the same App shape without touching
// the editor core (see CLAUDE.md).
package app

import (
	"fmt"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"

	"github.com/yoctoMNS/GolangTextEditor/internal/buffer"
	"github.com/yoctoMNS/GolangTextEditor/internal/editor"
	"github.com/yoctoMNS/GolangTextEditor/internal/viewport"
)

const (
	charWidth   = 7
	lineHeight  = 16
	marginX     = 4
	marginY     = 4
	statusBarPx = lineHeight + 4

	repeatDelayTicks    = 24 // ~0.4s at 60 TPS before a held key starts repeating
	repeatIntervalTicks = 3  // ~20 repeats/sec while held

	wheelScrollLines = 3 // lines scrolled per full wheel notch

	gutterPaddingX = 4 // horizontal padding on each side of the line numbers
)

var (
	gutterColor    = color.RGBA{R: 0x70, G: 0x70, B: 0x70, A: 0xff}
	selectionColor = color.RGBA{R: 0x26, G: 0x4f, B: 0x78, A: 0xff}
)

var face = text.NewGoXFace(basicfont.Face7x13)

// App implements the ebiten.Game interface (Update/Draw/Layout are the
// method names that interface requires), driving an *editor.Editor from
// keyboard input and rendering its buffer as plain monospaced text. The
// type is named App, not Game, because this is a text editor front-end,
// not a game.
type App struct {
	Ed *editor.Editor

	blinkTick      int
	keyHeldFor     map[ebiten.Key]int
	lastErr        string
	scrollLine     int
	prevCursorLine int
}

// New creates an App around ed.
func New(ed *editor.Editor) *App {
	return &App{Ed: ed, keyHeldFor: make(map[ebiten.Key]int)}
}

// Update advances the editor state by one tick in response to input.
func (a *App) Update() error {
	a.blinkTick++

	for _, r := range ebiten.AppendInputChars(nil) {
		a.Ed.InsertRune(r)
	}

	a.handleRepeatable(ebiten.KeyEnter, a.Ed.InsertNewline)
	a.handleRepeatable(ebiten.KeyBackspace, a.Ed.Backspace)
	a.handleRepeatable(ebiten.KeyDelete, a.Ed.Delete)
	a.handleMoveKey(ebiten.KeyArrowLeft, a.Ed.MoveLeft)
	a.handleMoveKey(ebiten.KeyArrowRight, a.Ed.MoveRight)
	a.handleMoveKey(ebiten.KeyArrowUp, a.Ed.MoveUp)
	a.handleMoveKey(ebiten.KeyArrowDown, a.Ed.MoveDown)
	a.handleRepeatable(ebiten.KeyHome, a.Ed.Home)
	a.handleRepeatable(ebiten.KeyEnd, a.Ed.End)

	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyS) {
		a.save()
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		a.handleClick()
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		a.handleDrag()
	}

	a.updateScroll()

	return nil
}

// handleClick moves the caret to wherever the left mouse button was just
// pressed, cancels any previously active selection, and anchors a new
// (as yet empty) selection there. Anchoring here rather than only in
// handleDrag means a plain click without any drag leaves no selection
// (StartSelection sets the anchor equal to the cursor, and HasSelection
// treats an anchor equal to the cursor as no selection), while a
// click-and-drag extends it (see handleDrag).
func (a *App) handleClick() {
	a.Ed.Cursor = a.cursorAtMouse()
	a.Ed.ClearSelection()
	a.Ed.StartSelection()
}

// handleDrag moves the caret to the current mouse position while the left
// button is held, extending the selection anchored by handleClick.
func (a *App) handleDrag() {
	a.Ed.Cursor = a.cursorAtMouse()
}

// cursorAtMouse converts the current mouse position to a valid caret
// position in the buffer.
func (a *App) cursorAtMouse() buffer.Position {
	x, y := ebiten.CursorPosition()
	gutterW := viewport.GutterWidth(a.Ed.Buf.LineCount(), charWidth, gutterPaddingX)
	textX := marginX + gutterW

	line, col := viewport.PositionAt(x, y, textX, marginY, a.scrollLine, charWidth, lineHeight)
	return a.Ed.Buf.Clamp(buffer.Position{Line: line, Col: col})
}

// updateScroll keeps the cursor in view whenever it has moved since the
// last tick, and applies mouse wheel input to scroll the view on its own.
// Wheel scrolling is deliberately not re-clamped to the cursor position
// every tick, so the user can scroll to look around without the view
// snapping back until the cursor itself moves.
func (a *App) updateScroll() {
	_, winH := ebiten.WindowSize()
	visibleLines := visibleLineCount(winH)
	lineCount := a.Ed.Buf.LineCount()

	if a.Ed.Cursor.Line != a.prevCursorLine {
		a.scrollLine = viewport.ClampScroll(a.scrollLine, a.Ed.Cursor.Line, visibleLines, lineCount)
		a.prevCursorLine = a.Ed.Cursor.Line
	}

	if _, wheelY := ebiten.Wheel(); wheelY != 0 {
		a.scrollLine -= int(wheelY * wheelScrollLines)
	}

	if maxScroll := lineCount - visibleLines; a.scrollLine > maxScroll {
		a.scrollLine = maxScroll
	}
	if a.scrollLine < 0 {
		a.scrollLine = 0
	}
}

func (a *App) save() {
	if a.Ed.Path == "" {
		a.lastErr = "no file path to save to (open editor with a file argument)"
		return
	}
	if err := a.Ed.Save(); err != nil {
		a.lastErr = fmt.Sprintf("save failed: %v", err)
		return
	}
	a.lastErr = ""
}

// handleMoveKey wraps a cursor-movement action so that holding Shift
// extends the current selection (starting one at the caret if none is
// active yet) instead of moving the caret alone, matching the
// Shift+Arrow convention of other editors. Moving without Shift cancels
// any active selection.
func (a *App) handleMoveKey(key ebiten.Key, move func()) {
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)
	a.handleRepeatable(key, func() {
		if shift {
			a.Ed.StartSelection()
		} else {
			a.Ed.ClearSelection()
		}
		move()
	})
}

// handleRepeatable calls action on the tick a key is first pressed, and
// again on a fixed interval while it is held, so navigation and deletion
// feel natural without needing per-key state machines at call sites.
func (a *App) handleRepeatable(key ebiten.Key, action func()) {
	if !ebiten.IsKeyPressed(key) {
		delete(a.keyHeldFor, key)
		return
	}
	held := a.keyHeldFor[key]
	if held == 0 {
		action()
	} else if held >= repeatDelayTicks && (held-repeatDelayTicks)%repeatIntervalTicks == 0 {
		action()
	}
	a.keyHeldFor[key] = held + 1
}

// Draw renders the buffer contents, a blinking caret and a status bar.
func (a *App) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 0x1e, G: 0x1e, B: 0x1e, A: 0xff})

	_, screenH := screen.Bounds().Dx(), screen.Bounds().Dy()
	visibleLines := visibleLineCount(screenH)

	buf := a.Ed.Buf
	gutterW := viewport.GutterWidth(buf.LineCount(), charWidth, gutterPaddingX)
	textX := marginX + gutterW

	lastVisible := a.scrollLine + visibleLines
	if lastVisible > buf.LineCount() {
		lastVisible = buf.LineCount()
	}
	for i := a.scrollLine; i < lastVisible; i++ {
		y := marginY + (i-a.scrollLine)*lineHeight

		if colStart, colEnd, ok := a.Ed.SelectionOnLine(i); ok {
			hx := float32(textX + colStart*charWidth)
			hw := float32((colEnd - colStart) * charWidth)
			vector.FillRect(screen, hx, float32(y), hw, float32(lineHeight), selectionColor, false)
		}

		lineNum := strconv.Itoa(i + 1)
		numX := marginX + gutterW - gutterPaddingX - len(lineNum)*charWidth
		numOp := &text.DrawOptions{}
		numOp.GeoM.Translate(float64(numX), float64(y))
		numOp.ColorScale.ScaleWithColor(gutterColor)
		text.Draw(screen, lineNum, face, numOp)

		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(textX), float64(y))
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, buf.Line(i), face, op)
	}

	if a.blinkTick%60 < 30 {
		cx := float32(textX + a.Ed.Cursor.Col*charWidth)
		cy := float32(marginY + (a.Ed.Cursor.Line-a.scrollLine)*lineHeight)
		vector.StrokeLine(screen, cx, cy, cx, cy+lineHeight-2, 1, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, false)
	}

	a.drawStatusBar(screen)
}

// visibleLineCount returns how many text lines fit in a window of the
// given height, always at least 1.
func visibleLineCount(height int) int {
	return viewport.LineCount(height-statusBarPx-marginY, lineHeight)
}

func (a *App) drawStatusBar(screen *ebiten.Image) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	barY := h - statusBarPx
	vector.FillRect(screen, 0, float32(barY), float32(w), float32(statusBarPx), color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)

	name := a.Ed.Path
	if name == "" {
		name = "[No Name]"
	}
	modified := ""
	if a.Ed.Modified {
		modified = " [+]"
	}
	status := fmt.Sprintf("%s%s  Ln %d, Col %d", name, modified, a.Ed.Cursor.Line+1, a.Ed.Cursor.Col+1)
	if a.lastErr != "" {
		status = a.lastErr
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(marginX, float64(barY+2))
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, status, face, op)
}

// Layout keeps the window's logical resolution equal to its actual size.
func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
