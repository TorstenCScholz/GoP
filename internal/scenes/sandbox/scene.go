// Package sandbox provides a test scene for development and prototyping.
package sandbox

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/torsten/GoP/internal/assets"
	"github.com/torsten/GoP/internal/gfx"
	"github.com/torsten/GoP/internal/input"
)

const (
	// Movement speed in pixels per frame.
	movementSpeed = 4.0
	// Player size in pixels (for bounds checking).
	playerSize = 32
	// Sprite scale for better visibility.
	spriteScale = 2.0
	// Animation frame duration.
	frameDuration = 100 * time.Millisecond
)

// Colors for the scene.
var (
	backgroundColor = color.RGBA{0x10, 0x10, 0x20, 0xff}
)

// Scene represents the sandbox test scene.
type Scene struct {
	playerX  float64
	playerY  float64
	width    int
	height   int
	sprite   *gfx.Sprite
	animator *gfx.Animator
}

// New creates a new sandbox scene.
func New() *Scene {
	s := &Scene{
		playerX: 100,
		playerY: 100,
		width:   640,
		height:  360,
	}

	// Load spritesheet and create animation
	if err := s.initSprite(); err != nil {
		// Log error but continue - sprite will be nil
		fmt.Printf("Failed to load sprite: %v\n", err)
	}

	return s
}

// initSprite loads the spritesheet and creates the sprite animation.
func (s *Scene) initSprite() error {
	// Load spritesheet from assets/sprites directory
	fs, err := assets.SubFS("sprites")
	if err != nil {
		return fmt.Errorf("failed to open sprites directory: %w", err)
	}

	img, err := assets.LoadImage(fs, "test_sheet.png")
	if err != nil {
		return fmt.Errorf("failed to load spritesheet: %w", err)
	}

	// Create sheet with 32x32 frames
	sheet := assets.NewSheet(img, 32, 32)

	// Create animation from sheet
	anim := gfx.NewAnimationFromSheet(sheet, frameDuration)
	anim.Loop = true

	// Create animator and start playing
	s.animator = gfx.NewAnimator(anim)
	s.animator.Play()

	// Create sprite with scale for better visibility
	s.sprite = gfx.NewSprite(nil) // Image will be set in Draw
	s.sprite.SetScale(spriteScale, spriteScale)
	s.sprite.SetOrigin(0.5, 0.5) // Center origin for rotation

	return nil
}

// Update implements app.Scene.Update.
func (s *Scene) Update(inp *input.Input) error {
	// Update animation
	if s.animator != nil {
		// Use fixed timestep for animation (assuming 60 FPS)
		s.animator.Update(time.Second / 60)
	}

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

	// Calculate scaled size for bounds
	scaledSize := float64(playerSize) * spriteScale
	halfSize := scaledSize / 2 // Since origin is center (0.5, 0.5)

	// Keep player within bounds (playerX/Y is the CENTER of the sprite)
	if s.playerX < halfSize {
		s.playerX = halfSize
	}
	if s.playerX > float64(s.width)-halfSize {
		s.playerX = float64(s.width) - halfSize
	}
	if s.playerY < halfSize {
		s.playerY = halfSize
	}
	if s.playerY > float64(s.height)-halfSize {
		s.playerY = float64(s.height) - halfSize
	}

	return nil
}

// Draw implements app.Scene.Draw.
func (s *Scene) Draw(screen *ebiten.Image) {
	// Fill background
	screen.Fill(backgroundColor)

	// Draw animated sprite
	if s.sprite != nil && s.animator != nil {
		// Update sprite image from current animation frame
		s.sprite.Image = s.animator.CurrentFrame()
		// Update position
		s.sprite.SetPosition(s.playerX, s.playerY)
		// Draw sprite
		s.sprite.Draw(screen)
	} else {
		// Fallback: draw a rectangle if sprite failed to load
		ebitenutil.DrawRect(screen, s.playerX, s.playerY, playerSize, playerSize, color.RGBA{0xff, 0x00, 0x00, 0xff})
	}
}

// Layout implements app.Scene.Layout.
func (s *Scene) Layout(outsideW, outsideH int) (int, int) {
	s.width = outsideW
	s.height = outsideH
	return outsideW, outsideH
}

// DebugInfo implements app.Scene.DebugInfo.
func (s *Scene) DebugInfo() string {
	info := fmt.Sprintf("Player: (%.0f, %.0f)", s.playerX, s.playerY)
	if s.animator != nil {
		info += fmt.Sprintf(" | Frame: %d", s.animator.CurrentFrameIndex())
	}
	return info
}

// DrawDebug implements app.SceneDebugger.DrawDebug.
// Draws a white border around the player collision area for debugging.
func (s *Scene) DrawDebug(screen *ebiten.Image) {
	if s.sprite == nil {
		return
	}

	// Calculate scaled size and half-size
	scaledSize := float64(playerSize) * spriteScale
	halfSize := scaledSize / 2

	// Draw white border around collision area (playerX/Y is center)
	left := s.playerX - halfSize
	top := s.playerY - halfSize

	// Draw rectangle border using DrawRect (it draws a filled rect, so we draw 4 lines)
	borderColor := color.RGBA{255, 255, 255, 255}
	borderWidth := 2.0

	// Top border
	ebitenutil.DrawRect(screen, left, top, scaledSize, borderWidth, borderColor)
	// Bottom border
	ebitenutil.DrawRect(screen, left, top+scaledSize-borderWidth, scaledSize, borderWidth, borderColor)
	// Left border
	ebitenutil.DrawRect(screen, left, top, borderWidth, scaledSize, borderColor)
	// Right border
	ebitenutil.DrawRect(screen, left+scaledSize-borderWidth, top, borderWidth, scaledSize, borderColor)
}
