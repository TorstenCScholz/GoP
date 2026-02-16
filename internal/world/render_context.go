package world

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/torsten/GoP/internal/camera"
)

// RenderContext holds all render state needed for drawing.
// It consolidates camera coordinates and other render state into a single struct
// that can be passed to entity draw methods.
type RenderContext struct {
	Cam    *camera.Camera
	Debug  bool
	DT     float64 // Delta time for animations
	Screen *ebiten.Image
}

// NewRenderContext creates a new render context with the given parameters.
func NewRenderContext(cam *camera.Camera, screen *ebiten.Image, dt float64) *RenderContext {
	return &RenderContext{
		Cam:    cam,
		Debug:  false,
		DT:     dt,
		Screen: screen,
	}
}

// WorldToScreen converts world coordinates to screen coordinates.
// Returns the screen position where a world point should be rendered.
func (ctx *RenderContext) WorldToScreen(worldX, worldY float64) (screenX, screenY float64) {
	if ctx.Cam == nil {
		return worldX, worldY
	}
	return worldX - ctx.Cam.X, worldY - ctx.Cam.Y
}

// ScreenToWorld converts screen coordinates to world coordinates.
// Returns the world position at a given screen point.
func (ctx *RenderContext) ScreenToWorld(screenX, screenY float64) (worldX, worldY float64) {
	if ctx.Cam == nil {
		return screenX, screenY
	}
	return screenX + ctx.Cam.X, screenY + ctx.Cam.Y
}

// IsVisible checks if a world-space rect is visible on screen.
// worldX, worldY is the top-left corner of the rect in world coordinates.
// width, height is the size of the rect.
func (ctx *RenderContext) IsVisible(worldX, worldY, width, height float64) bool {
	if ctx.Cam == nil {
		return true
	}

	// Get camera bounds
	camX, camY := ctx.Cam.X, ctx.Cam.Y
	camW := float64(ctx.Cam.ViewportW)
	camH := float64(ctx.Cam.ViewportH)

	// Check if rect overlaps with camera viewport
	return worldX+width >= camX &&
		worldX <= camX+camW &&
		worldY+height >= camY &&
		worldY <= camY+camH
}

// CameraX returns the camera's X position (for backward compatibility).
func (ctx *RenderContext) CameraX() float64 {
	if ctx.Cam == nil {
		return 0
	}
	return ctx.Cam.X
}

// CameraY returns the camera's Y position (for backward compatibility).
func (ctx *RenderContext) CameraY() float64 {
	if ctx.Cam == nil {
		return 0
	}
	return ctx.Cam.Y
}
