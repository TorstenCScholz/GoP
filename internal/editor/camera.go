package editor

import (
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// CameraPanSpeed is the speed of camera panning in pixels per frame.
	CameraPanSpeed = 5.0
	// ZoomSpeed is the multiplier for zoom changes.
	ZoomSpeed = 0.1
	// MinZoom is the minimum zoom level.
	MinZoom = 0.25
	// MaxZoom is the maximum zoom level.
	MaxZoom = 4.0
)

// Camera handles the viewport transformation for the editor canvas.
type Camera struct {
	X    float64 // Camera position in world coordinates
	Y    float64
	Zoom float64 // Zoom level (1.0 = 100%)

	// Interaction state
	isPanning  bool
	panStartX  int
	panStartY  int
	panCameraX float64
	panCameraY float64
}

// NewCamera creates a new camera with default values.
func NewCamera() *Camera {
	return &Camera{
		X:         0,
		Y:         0,
		Zoom:      1.0,
		isPanning: false,
	}
}

// Update processes input for camera control.
// Pan with middle mouse button or arrow keys.
// Zoom with mouse wheel.
func (c *Camera) Update() {
	// Handle keyboard panning
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		c.X -= CameraPanSpeed / c.Zoom
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		c.X += CameraPanSpeed / c.Zoom
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		c.Y -= CameraPanSpeed / c.Zoom
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		c.Y += CameraPanSpeed / c.Zoom
	}

	// Handle middle mouse button panning
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		mx, my := ebiten.CursorPosition()
		if !c.isPanning {
			c.isPanning = true
			c.panStartX = mx
			c.panStartY = my
			c.panCameraX = c.X
			c.panCameraY = c.Y
		} else {
			dx := float64(c.panStartX-mx) / c.Zoom
			dy := float64(c.panStartY-my) / c.Zoom
			c.X = c.panCameraX + dx
			c.Y = c.panCameraY + dy
		}
	} else {
		c.isPanning = false
	}

	// Handle zoom with mouse wheel
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		// Get mouse position before zoom for zoom-at-cursor
		mx, my := ebiten.CursorPosition()
		worldX, worldY := c.ScreenToWorld(mx, my)

		// Apply zoom
		c.Zoom += wheelY * ZoomSpeed
		if c.Zoom < MinZoom {
			c.Zoom = MinZoom
		}
		if c.Zoom > MaxZoom {
			c.Zoom = MaxZoom
		}

		// Adjust camera to zoom at cursor position
		newWorldX, newWorldY := c.ScreenToWorld(mx, my)
		c.X += worldX - newWorldX
		c.Y += worldY - newWorldY
	}
}

// ScreenToWorld converts screen coordinates to world coordinates.
func (c *Camera) ScreenToWorld(screenX, screenY int) (worldX, worldY float64) {
	worldX = float64(screenX)/c.Zoom + c.X
	worldY = float64(screenY)/c.Zoom + c.Y
	return
}

// WorldToScreen converts world coordinates to screen coordinates.
func (c *Camera) WorldToScreen(worldX, worldY float64) (screenX, screenY int) {
	screenX = int((worldX - c.X) * c.Zoom)
	screenY = int((worldY - c.Y) * c.Zoom)
	return
}

// ScreenToWorldTile converts screen coordinates to tile coordinates.
func (c *Camera) ScreenToWorldTile(screenX, screenY, tileWidth, tileHeight int) (tileX, tileY int) {
	worldX, worldY := c.ScreenToWorld(screenX, screenY)
	tileX = int(worldX) / tileWidth
	tileY = int(worldY) / tileHeight
	return
}

// SetPosition sets the camera position.
func (c *Camera) SetPosition(x, y float64) {
	c.X = x
	c.Y = y
}

// SetZoom sets the camera zoom level.
func (c *Camera) SetZoom(zoom float64) {
	if zoom < MinZoom {
		zoom = MinZoom
	}
	if zoom > MaxZoom {
		zoom = MaxZoom
	}
	c.Zoom = zoom
}

// Reset resets the camera to default position and zoom.
func (c *Camera) Reset() {
	c.X = 0
	c.Y = 0
	c.Zoom = 1.0
	c.isPanning = false
}
