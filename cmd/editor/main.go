// Command editor is the entry point for the liightweight text editor.
// Usage:
//
// editor [file]
//
// If a file path is given it is loaded on startup; Ctrl+S saves back to that path.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yoctoMNS/GolangTextEditor/internal/app"
	"github.com/yoctoMNS/GolangTextEditor/internal/editor"
)

func main() {
	ed := editor.New()
	if len(os.Args) > 1 {
		path := os.Args[1]
		if err := ed.Load(path); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not load %q: %v (starting with an empty buffer)\n", path, err)
			ed.Path = path
		}
	}
	ebiten.SetWindowSize(960, 640)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("GolangTextEditor")
	if err := ebiten.RunGame(app.New(ed)); err != nil {
		log.Fatal(err)
	}
}
