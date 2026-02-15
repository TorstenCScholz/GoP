// Package sandbox provides a test scene for development and prototyping.
package sandbox

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/torsten/GoP/internal/input"
)

const (
	// Movement speed in pixels per frame.
	movementSpeed = 4.0
	// Player size in pixels.
	playerSize = 16
)

// Colors for the scene.
var (
	backgroundColor = color.RGBA{0x10, 0x10, 0x20, 0xff}
	playerColor     = color.RGBA{0x00, 0xff, 0x00, 0xff}
)

// Scene represents the sandbox test scene.
type Scene struct {
	playerX float64
	playerY float64
	width   int
	height  int
}

// New creates a new sandbox scene.
func New() *Scene {
	return &Scene{
		playerX: 100,
		playerY: 100,
		width:   640,
		height:  360,
	}
}

// Update implements app.Scene.Update.
func (s *Scene) Update(inp *input.Input) error {
	// Move player based on input
	if inp.Pressed(input.ActionMoveLeft) {
		s.playerX -= movementSpeed
	}
	if inp.Pressed(input.ActionMoveRight) {
		s.playerX += movementSpeed
	}
	if inp.Pressed(input.ActionMoveUp) {
		s.playerY -= movementSpeed
	}
	if inp.Pressed(input.ActionMoveDown) {
		s.playerY += movementSpeed
	}

	// Keep player within bounds
	if s.playerX < 0 {
		s.playerX = 0
	}
	if s.playerX > float64(s.width-playerSize) {
		s.playerX = float64(s.width - playerSize)
	}
	if s.playerY < 0 {
		s.playerY = 0
	}
	if s.playerY > float64(s.height-playerSize) {
		s.playerY = float64(s.height - playerSize)
	}

	return nil
}

// Draw implements app.Scene.Draw.
func (s *Scene) Draw(screen *ebiten.Image) {
	// Fill background
	screen.Fill(backgroundColor)

	// Draw player rectangle
	ebitenutil.DrawRect(screen, s.playerX, s.playerY, playerSize, playerSize, playerColor)
}

// Layout implements app.Scene.Layout.
func (s *Scene) Layout(outsideW, outsideH int) (int, int) {
	s.width = outsideW
	s.height = outsideH
	return outsideW, outsideH
}

// DebugInfo implements app.Scene.DebugInfo.
func (s *Scene) DebugInfo() string {
	return fmt.Sprintf("Player: (%.0f, %.0f)", s.playerX, s.playerY)
}
