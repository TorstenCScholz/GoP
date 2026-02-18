// Package main provides the entry point for the GoP Level Editor.
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/torsten/GoP/internal/editor"
)

func main() {
	// Create the editor application
	app := editor.NewApp()

	// Configure the window
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("GoP Level Editor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)

	// Run the editor
	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
