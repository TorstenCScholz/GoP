// Package editor provides the level editor functionality.
package editor

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Minimap provides a small overview of the entire level.
type Minimap struct {
	width  int
	height int
	x      int
	y      int
}

// NewMinimap creates a new minimap component.
func NewMinimap() *Minimap {
	return &Minimap{
		width:  150,
		height: 100,
		x:      10,
		y:      10,
	}
}

// Draw renders the minimap to the screen.
func (m *Minimap) Draw(screen *ebiten.Image, state *EditorState, camera *Camera) {
	if state == nil || !state.HasLevel() {
		return
	}

	screenWidth, _ := screen.Size()

	// Position minimap in top-left corner of canvas area
	m.x = screenWidth - PaletteWidth - ObjectPaletteWidth - m.width - 10
	m.y = 10

	// Draw background
	minimapImg := ebiten.NewImage(m.width, m.height)
	minimapImg.Fill(color.RGBA{20, 20, 30, 200})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(m.x), float64(m.y))
	screen.DrawImage(minimapImg, op)

	// Draw border
	borderColor := color.RGBA{80, 80, 100, 255}
	ebitenutil.DrawRect(screen, float64(m.x), float64(m.y), float64(m.width), 1, borderColor)
	ebitenutil.DrawRect(screen, float64(m.x), float64(m.y+m.height-1), float64(m.width), 1, borderColor)
	ebitenutil.DrawRect(screen, float64(m.x), float64(m.y), 1, float64(m.height), borderColor)
	ebitenutil.DrawRect(screen, float64(m.x+m.width-1), float64(m.y), 1, float64(m.height), borderColor)

	// Calculate scale
	mapWidth := state.MapData.Width() * state.MapData.TileWidth()
	mapHeight := state.MapData.Height() * state.MapData.TileHeight()

	scaleX := float64(m.width) / float64(mapWidth)
	scaleY := float64(m.height) / float64(mapHeight)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Draw objects as small rectangles
	for _, obj := range state.Objects {
		// Get color from schema
		schema := GetSchema(obj.Type)
		objColor := objectDefaultColor
		if schema != nil {
			objColor = parseColor(schema.Color)
		}

		// Calculate minimap position
		mx := m.x + int(obj.X*scale)
		my := m.y + int(obj.Y*scale)
		mw := int(obj.W * scale)
		mh := int(obj.H * scale)

		// Ensure minimum size of 2 pixels
		if mw < 2 {
			mw = 2
		}
		if mh < 2 {
			mh = 2
		}

		// Draw object
		objRect := ebiten.NewImage(mw, mh)
		objRect.Fill(objColor)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(mx), float64(my))
		screen.DrawImage(objRect, op)
	}

	// Draw viewport rectangle
	screenWidthActual, screenHeightActual := screen.Size()
	viewportX := m.x + int(camera.X*scale)
	viewportY := m.y + int(camera.Y*scale)
	viewportW := int(float64(screenWidthActual-PaletteWidth-ObjectPaletteWidth) * scale / camera.Zoom)
	viewportH := int(float64(screenHeightActual) * scale / camera.Zoom)

	// Clamp viewport to minimap bounds
	if viewportX < m.x {
		viewportW -= m.x - viewportX
		viewportX = m.x
	}
	if viewportY < m.y {
		viewportH -= m.y - viewportY
		viewportY = m.y
	}
	if viewportX+viewportW > m.x+m.width {
		viewportW = m.x + m.width - viewportX
	}
	if viewportY+viewportH > m.y+m.height {
		viewportH = m.y + m.height - viewportY
	}

	// Draw viewport rectangle
	if viewportW > 0 && viewportH > 0 {
		viewportColor := color.RGBA{255, 255, 255, 150}
		ebitenutil.DrawRect(screen, float64(viewportX), float64(viewportY), float64(viewportW), 1, viewportColor)
		ebitenutil.DrawRect(screen, float64(viewportX), float64(viewportY+viewportH-1), float64(viewportW), 1, viewportColor)
		ebitenutil.DrawRect(screen, float64(viewportX), float64(viewportY), 1, float64(viewportH), viewportColor)
		ebitenutil.DrawRect(screen, float64(viewportX+viewportW-1), float64(viewportY), 1, float64(viewportH), viewportColor)
	}
}

// HandleClick processes a click on the minimap to jump to that location.
// Returns true if the click was handled by the minimap.
func (m *Minimap) HandleClick(clickX, clickY int, state *EditorState, camera *Camera, canvasWidth, canvasHeight int) bool {
	if state == nil || !state.HasLevel() {
		return false
	}

	// Check if click is within minimap bounds
	if clickX < m.x || clickX >= m.x+m.width {
		return false
	}
	if clickY < m.y || clickY >= m.y+m.height {
		return false
	}

	// Calculate scale
	mapWidth := state.MapData.Width() * state.MapData.TileWidth()
	mapHeight := state.MapData.Height() * state.MapData.TileHeight()

	scaleX := float64(m.width) / float64(mapWidth)
	scaleY := float64(m.height) / float64(mapHeight)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Calculate position within minimap
	minimapClickX := float64(clickX - m.x)
	minimapClickY := float64(clickY - m.y)

	// Convert to world coordinates
	worldX := minimapClickX / scale
	worldY := minimapClickY / scale

	// Center camera on this position
	camera.X = worldX - float64(canvasWidth)/2/camera.Zoom
	camera.Y = worldY - float64(canvasHeight)/2/camera.Zoom

	return true
}

// Bounds returns the minimap's screen bounds.
func (m *Minimap) Bounds() (x, y, width, height int) {
	return m.x, m.y, m.width, m.height
}
