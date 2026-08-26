// Package app wires the UI-agnostic editor core to a concrete rendering
// and input backend. This is the only package in the module that may
// import a graphics library; internal/buffer and internal/editor stay
// backend-free so a different backend (raylib-go, go-sdl2, go-gl, ...)
// can be dropped in later behind the same Game shape without touching
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

// Game implements ebiten.Game, driving an *editor.Editor from keyboard
// input and rendering its buffer as plain monospaced text.
type Game struct {
	Ed *editor.Editor

	blinkTick  int
	keyHeldFor map[ebiten.Key]int
	lastErr    string
}

// NewGame creates a Game around ed.
func NewGame(ed *editor.Editor) *Game {
	return &Game{Ed: ed, keyHeldFor: make(map[ebiten.Key]int)}
}

// Update advances the editor state by one tick in response to input.
func (g *Game) Update() error {
	g.blinkTick++

	for _, r := range ebiten.AppendInputChars(nil) {
		g.Ed.InsertRune(r)
	}

	g.handleRepeatable(ebiten.KeyEnter, g.Ed.InsertNewline)
	g.handleRepeatable(ebiten.KeyBackspace, g.Ed.Backspace)
	g.handleRepeatable(ebiten.KeyDelete, g.Ed.Delete)
	g.handleRepeatable(ebiten.KeyArrowLeft, g.Ed.MoveLeft)
	g.handleRepeatable(ebiten.KeyArrowRight, g.Ed.MoveRight)
	g.handleRepeatable(ebiten.KeyArrowUp, g.Ed.MoveUp)
	g.handleRepeatable(ebiten.KeyArrowDown, g.Ed.MoveDown)
	g.handleRepeatable(ebiten.KeyHome, g.Ed.Home)
	g.handleRepeatable(ebiten.KeyEnd, g.Ed.End)

	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.save()
	}

	return nil
}

func (g *Game) save() {
	if g.Ed.Path == "" {
		g.lastErr = "no file path to save to (open editor with a file argument)"
		return
	}
	if err := g.Ed.Save(); err != nil {
		g.lastErr = fmt.Sprintf("save failed: %v", err)
		return
	}
	g.lastErr = ""
}

// handleRepeatable calls action on the tick a key is first pressed, and
// again on a fixed interval while it is held, so navigation and deletion
// feel natural without needing per-key state machines at call sites.
func (g *Game) handleRepeatable(key ebiten.Key, action func()) {
	if !ebiten.IsKeyPressed(key) {
		delete(g.keyHeldFor, key)
		return
	}
	held := g.keyHeldFor[key]
	if held == 0 {
		action()
	} else if held >= repeatDelayTicks && (held-repeatDelayTicks)%repeatIntervalTicks == 0 {
		action()
	}
	g.keyHeldFor[key] = held + 1
}

// Draw renders the buffer contents, a blinking caret and a status bar.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 0x1e, G: 0x1e, B: 0x1e, A: 0xff})

	buf := g.Ed.Buf
	for i := 0; i < buf.LineCount(); i++ {
		y := marginY + i*lineHeight
		op := &text.DrawOptions{}
		op.GeoM.Translate(marginX, float64(y))
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, buf.Line(i), face, op)
	}

	if g.blinkTick%60 < 30 {
		cx := float32(marginX + g.Ed.Cursor.Col*charWidth)
		cy := float32(marginY + g.Ed.Cursor.Line*lineHeight)
		vector.StrokeLine(screen, cx, cy, cx, cy+lineHeight-2, 1, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, false)
	}

	g.drawStatusBar(screen)
}

func (g *Game) drawStatusBar(screen *ebiten.Image) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	barY := h - statusBarPx
	vector.DrawFilledRect(screen, 0, float32(barY), float32(w), float32(statusBarPx), color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, false)

	name := g.Ed.Path
	if name == "" {
		name = "[No Name]"
	}
	modified := ""
	if g.Ed.Modified {
		modified = " [+]"
	}
	status := fmt.Sprintf("%s%s  Ln %d, Col %d", name, modified, g.Ed.Cursor.Line+1, g.Ed.Cursor.Col+1)
	if g.lastErr != "" {
		status = g.lastErr
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(marginX, float64(barY+2))
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, status, face, op)
}

// Layout keeps the window's logical resolution equal to its actual size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
