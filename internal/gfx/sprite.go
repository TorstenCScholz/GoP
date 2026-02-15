// Package gfx provides graphics utilities for sprite rendering and animation.
package gfx

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Sprite represents a drawable image with transformation properties.
// It supports position, scale, rotation, and origin point for transformations.
type Sprite struct {
	Image   *ebiten.Image
	X, Y    float64
	ScaleX  float64
	ScaleY  float64
	Rotation float64
	OriginX float64
	OriginY float64
}

// NewSprite creates a new sprite with the given image and default values.
// Default scale is 1.0, and default origin is center (0.5, 0.5).
func NewSprite(img *ebiten.Image) *Sprite {
	return &Sprite{
		Image:    img,
		X:        0,
		Y:        0,
		ScaleX:   1.0,
		ScaleY:   1.0,
		Rotation: 0,
		OriginX:  0.5,
		OriginY:  0.5,
	}
}

// SetPosition sets the sprite's position.
func (s *Sprite) SetPosition(x, y float64) {
	s.X = x
	s.Y = y
}

// SetScale sets the sprite's scale factors.
func (s *Sprite) SetScale(sx, sy float64) {
	s.ScaleX = sx
	s.ScaleY = sy
}

// SetRotation sets the sprite's rotation in radians.
func (s *Sprite) SetRotation(radians float64) {
	s.Rotation = radians
}

// SetOrigin sets the sprite's origin point for rotation and scale.
// Values are in 0-1 normalized range (0,0 = top-left, 1,1 = bottom-right).
func (s *Sprite) SetOrigin(ox, oy float64) {
	s.OriginX = ox
	s.OriginY = oy
}

// Draw renders the sprite to the target image with all transformations applied.
// Uses nearest-neighbor filtering for pixel art.
func (s *Sprite) Draw(screen *ebiten.Image) {
	if s.Image == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest

	// Get image bounds for origin calculation
	bounds := s.Image.Bounds()
	w, h := float64(bounds.Dx()), float64(bounds.Dy())

	// Calculate origin offset in pixels
	originOffsetX := w * s.OriginX
	originOffsetY := h * s.OriginY

	// Apply transformations in order:
	// 1. Translate so origin point is at (0,0) - for rotation/scale around origin
	op.GeoM.Translate(-originOffsetX, -originOffsetY)
	// 2. Scale
	op.GeoM.Scale(s.ScaleX, s.ScaleY)
	// 3. Rotate
	op.GeoM.Rotate(s.Rotation)
	// 4. Translate to final position (s.X, s.Y is where the origin point should be)
	op.GeoM.Translate(s.X, s.Y)

	screen.DrawImage(s.Image, op)
}
