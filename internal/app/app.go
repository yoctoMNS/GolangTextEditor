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

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"

	"github.com/yoctoMNS/GolangTextEditor/internal/editor"
)

const (
	charWidth   = 7
	lineHeight  = 16
	marginX     = 4
	marginY     = 4
	statusBarPx = lineHeight + 4

	repeatDelayTicks    = 24 // ~0.4s at 60 TPS before a held key starts repeating
	repeatIntervalTicks = 3  // ~20 repeats/sec while held
)

var face = text.NewGoXFace(basicfont.Face7x13)

// App implements the ebiten.Game interface (Update/Draw/Layout are the
// method names that interface requires), driving an *editor.Editor from
// keyboard input and rendering its buffer as plain monospaced text. The
// type is named App, not Game, because this is a text editor front-end,
// not a game.
type App struct {
	Ed *editor.Editor

	blinkTick  int
	keyHeldFor map[ebiten.Key]int
	lastErr    string
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
	a.handleRepeatable(ebiten.KeyArrowLeft, a.Ed.MoveLeft)
	a.handleRepeatable(ebiten.KeyArrowRight, a.Ed.MoveRight)
	a.handleRepeatable(ebiten.KeyArrowUp, a.Ed.MoveUp)
	a.handleRepeatable(ebiten.KeyArrowDown, a.Ed.MoveDown)
	a.handleRepeatable(ebiten.KeyHome, a.Ed.Home)
	a.handleRepeatable(ebiten.KeyEnd, a.Ed.End)

	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyS) {
		a.save()
	}

	return nil
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

	buf := a.Ed.Buf
	for i := 0; i < buf.LineCount(); i++ {
		y := marginY + i*lineHeight
		op := &text.DrawOptions{}
		op.GeoM.Translate(marginX, float64(y))
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, buf.Line(i), face, op)
	}

	if a.blinkTick%60 < 30 {
		cx := float32(marginX + a.Ed.Cursor.Col*charWidth)
		cy := float32(marginY + a.Ed.Cursor.Line*lineHeight)
		vector.StrokeLine(screen, cx, cy, cx, cy+lineHeight-2, 1, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, false)
	}

	a.drawStatusBar(screen)
}

func (a *App) drawStatusBar(screen *ebiten.Image) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	barY := h - statusBarPx
	vector.DrawFilledRect(screen, 0, float32(barY), float32(w), float32(statusBarPx), color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)

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
