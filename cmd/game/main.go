// Package main provides the entrypoint for the game.
package main

import (
	"log"

	"github.com/torsten/GoP/internal/app"
	"github.com/torsten/GoP/internal/scenes/sandbox"
)

func main() {
	// Create configuration
	cfg := &app.Config{
		WindowWidth:  640,
		WindowHeight: 360,
		WindowTitle:  "GoP Game",
		DebugMode:    false,
	}

	// Create app
	game := app.New(cfg)

	// Create and set initial scene
	scene := sandbox.New()
	game.SetScene(scene)

	// Run the game
	if err := game.Run(); err != nil {
		log.Fatal(err)
	}
}
