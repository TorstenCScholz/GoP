package world

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Camera provides viewport offset for scrolling levels.
type Camera struct {
	X, Y       float64 // Camera position (top-left of viewport)
	ViewWidth  int     // Viewport width in pixels
	ViewHeight int     // Viewport height in pixels
}

// NewCamera creates a new camera with the given viewport size.
func NewCamera(viewWidth, viewHeight int) *Camera {
	return &Camera{
		X:          0,
		Y:          0,
		ViewWidth:  viewWidth,
		ViewHeight: viewHeight,
	}
}

// CenterOn centers the camera on the given world position.
// The camera is constrained to stay within map bounds.
func (c *Camera) CenterOn(x, y float64, mapPixelWidth, mapPixelHeight int) {
	// Center on target
	c.X = x - float64(c.ViewWidth)/2
	c.Y = y - float64(c.ViewHeight)/2

	// Constrain to map bounds
	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
	if c.X > float64(mapPixelWidth-c.ViewWidth) {
		c.X = float64(mapPixelWidth - c.ViewWidth)
	}
	if c.Y > float64(mapPixelHeight-c.ViewHeight) {
		c.Y = float64(mapPixelHeight - c.ViewHeight)
	}
}

// VisibleBounds returns the visible tile range (tx1, ty1, tx2, ty2).
// The range is inclusive for start, exclusive for end.
func (c *Camera) VisibleBounds(tileSize int) (int, int, int, int) {
	tx1 := int(c.X) / tileSize
	ty1 := int(c.Y) / tileSize
	tx2 := (int(c.X) + c.ViewWidth + tileSize - 1) / tileSize
	ty2 := (int(c.Y) + c.ViewHeight + tileSize - 1) / tileSize
	return tx1, ty1, tx2, ty2
}

// WorldToScreen converts world coordinates to screen coordinates.
func (c *Camera) WorldToScreen(wx, wy float64) (sx, sy float64) {
	return wx - c.X, wy - c.Y
}

// ScreenToWorld converts screen coordinates to world coordinates.
func (c *Camera) ScreenToWorld(sx, sy float64) (wx, wy float64) {
	return sx + c.X, sy + c.Y
}

// MapRenderer handles rendering a map with camera support.
type MapRenderer struct {
	m   *Map
	cam *Camera
}

// NewMapRenderer creates a new renderer for the given map.
func NewMapRenderer(m *Map) *MapRenderer {
	return &MapRenderer{
		m: m,
	}
}

// SetCamera updates the camera reference.
func (r *MapRenderer) SetCamera(cam *Camera) {
	r.cam = cam
}

// Draw renders all visible tiles to the screen.
// Only tiles within the camera viewport are drawn.
// camX and camY are the camera offset in world pixels.
func (r *MapRenderer) Draw(screen *ebiten.Image, camX, camY float64) {
	if r.m == nil || r.m.tileset == nil {
		return
	}

	tileW := r.m.tileWidth
	tileH := r.m.tileHeight

	// Calculate visible tile range
	tx1 := int(camX) / tileW
	ty1 := int(camY) / tileH

	// Add 1 to ensure we cover tiles partially visible
	tx2 := (int(camX) + screen.Bounds().Dx() + tileW - 1) / tileW
	ty2 := (int(camY) + screen.Bounds().Dy() + tileH - 1) / tileH

	// Clamp to map bounds
	if tx1 < 0 {
		tx1 = 0
	}
	if ty1 < 0 {
		ty1 = 0
	}
	if tx2 > r.m.width {
		tx2 = r.m.width
	}
	if ty2 > r.m.height {
		ty2 = r.m.height
	}

	// Draw each layer
	for _, layer := range r.m.layers {
		if layer.Name() == "Collision" {
			continue // Don't render collision layer
		}
		for ty := ty1; ty < ty2; ty++ {
			for tx := tx1; tx < tx2; tx++ {
				tileID := layer.TileAt(tx, ty)
				if tileID == 0 {
					continue // Empty tile
				}

				// Tiled uses 1-based IDs, convert to 0-based
				tile := r.m.tileset.Tile(tileID - 1)
				if tile == nil {
					continue
				}

				// Calculate screen position
				worldX := float64(tx * tileW)
				worldY := float64(ty * tileH)
				screenX := worldX - camX
				screenY := worldY - camY

				// Draw tile
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(screenX, screenY)
				op.Filter = ebiten.FilterNearest
				screen.DrawImage(tile, op)
			}
		}
	}
}

// DrawLayer renders a specific layer by name.
// camX and camY are the camera offset in world pixels.
func (r *MapRenderer) DrawLayer(screen *ebiten.Image, layerName string, camX, camY float64) {
	if r.m == nil || r.m.tileset == nil {
		return
	}

	layer := r.m.Layer(layerName)
	if layer == nil {
		return
	}

	tileW := r.m.tileWidth
	tileH := r.m.tileHeight

	// Calculate visible tile range
	tx1 := int(camX) / tileW
	ty1 := int(camY) / tileH
	tx2 := (int(camX) + screen.Bounds().Dx() + tileW - 1) / tileW
	ty2 := (int(camY) + screen.Bounds().Dy() + tileH - 1) / tileH

	// Clamp to map bounds
	if tx1 < 0 {
		tx1 = 0
	}
	if ty1 < 0 {
		ty1 = 0
	}
	if tx2 > r.m.width {
		tx2 = r.m.width
	}
	if ty2 > r.m.height {
		ty2 = r.m.height
	}

	for ty := ty1; ty < ty2; ty++ {
		for tx := tx1; tx < tx2; tx++ {
			tileID := layer.TileAt(tx, ty)
			if tileID == 0 {
				continue
			}

			tile := r.m.tileset.Tile(tileID - 1)
			if tile == nil {
				continue
			}

			worldX := float64(tx * tileW)
			worldY := float64(ty * tileH)
			screenX := worldX - camX
			screenY := worldY - camY

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(screenX, screenY)
			op.Filter = ebiten.FilterNearest
			screen.DrawImage(tile, op)
		}
	}
}

// DrawWithCamera renders the map using the Camera struct.
// This is an alternative to Draw that uses the Camera struct directly.
func (r *MapRenderer) DrawWithCamera(screen *ebiten.Image) {
	if r.cam == nil {
		return
	}
	r.Draw(screen, r.cam.X, r.cam.Y)
}

// DrawWithContext renders the map using a RenderContext.
func (r *MapRenderer) DrawWithContext(screen *ebiten.Image, ctx *RenderContext) {
	if ctx.Cam == nil {
		return
	}
	r.Draw(screen, ctx.Cam.X, ctx.Cam.Y)
}
