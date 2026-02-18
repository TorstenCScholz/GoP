package editor

import (
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/torsten/GoP/internal/world"
)

const (
	// ObjectPaletteWidth is the width of the object palette sidebar in pixels.
	ObjectPaletteWidth = 150
	// ObjectPalettePadding is the padding inside the object palette area in pixels.
	ObjectPalettePadding = 10
	// ObjectButtonHeight is the height of each object type button.
	ObjectButtonHeight = 40
	// ObjectButtonSpacing is the spacing between object buttons.
	ObjectButtonSpacing = 4
)

// ObjectPalette handles rendering and interaction for the object type palette.
type ObjectPalette struct {
	selectedType world.ObjectType
	hoveredIndex int
	schemas      []*ObjectSchema
}

// NewObjectPalette creates a new object palette.
func NewObjectPalette() *ObjectPalette {
	return &ObjectPalette{
		selectedType: world.ObjectTypeSpawn,
		hoveredIndex: -1,
		schemas:      GetAllSchemas(),
	}
}

// SelectedType returns the currently selected object type.
func (p *ObjectPalette) SelectedType() world.ObjectType {
	return p.selectedType
}

// SetSelectedType sets the selected object type.
func (p *ObjectPalette) SetSelectedType(typ world.ObjectType) {
	p.selectedType = typ
}

// Draw renders the object palette to the screen.
// The palette is drawn in the right sidebar, below the tile palette.
func (p *ObjectPalette) Draw(screen *ebiten.Image, startY int) {
	screenWidth, screenHeight := screen.Size()
	_ = screenHeight

	// Calculate palette position (right side, below tile palette)
	paletteX := screenWidth - ObjectPaletteWidth

	// Draw palette background
	paletteBg := ebiten.NewImage(ObjectPaletteWidth, screenHeight-startY)
	paletteBg.Fill(objectPaletteBgColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(paletteX), float64(startY))
	screen.DrawImage(paletteBg, op)

	// Draw title
	titleY := startY + ObjectPalettePadding
	ebitenutil.DebugPrintAt(screen, "Objects", paletteX+ObjectPalettePadding, titleY)

	// Draw object type buttons
	buttonY := titleY + 20
	for i, schema := range p.schemas {
		y := buttonY + i*(ObjectButtonHeight+ObjectButtonSpacing)
		p.drawButton(screen, paletteX+ObjectPalettePadding, y, schema, i)
	}
}

// drawButton renders a single object type button.
func (p *ObjectPalette) drawButton(screen *ebiten.Image, x, y int, schema *ObjectSchema, index int) {
	buttonWidth := ObjectPaletteWidth - 2*ObjectPalettePadding

	// Determine button color based on state
	bgColor := objectButtonColor
	if p.selectedType == world.ObjectType(schema.Type) {
		bgColor = objectButtonSelectedColor
	} else if p.hoveredIndex == index {
		bgColor = objectButtonHoverColor
	}

	// Draw button background
	buttonImg := ebiten.NewImage(buttonWidth, ObjectButtonHeight)
	buttonImg.Fill(bgColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(buttonImg, op)

	// Draw color indicator (small colored rectangle)
	indicatorSize := 16
	indicatorX := x + 4
	indicatorY := y + (ObjectButtonHeight-indicatorSize)/2
	indicatorColor := parseColor(schema.Color)
	indicatorImg := ebiten.NewImage(indicatorSize, indicatorSize)
	indicatorImg.Fill(indicatorColor)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(indicatorX), float64(indicatorY))
	screen.DrawImage(indicatorImg, op)

	// Draw object name
	nameX := indicatorX + indicatorSize + 6
	nameY := y + (ObjectButtonHeight-16)/2 // Approximate vertical centering
	ebitenutil.DebugPrintAt(screen, schema.Name, nameX, nameY)
}

// HandleClick processes mouse clicks in the object palette.
// Returns true if the click was handled by the palette.
func (p *ObjectPalette) HandleClick(screenX, screenY, screenWidth, startY int) bool {
	// Check if click is in palette area
	paletteX := screenWidth - ObjectPaletteWidth
	if screenX < paletteX {
		return false
	}
	if screenY < startY {
		return false
	}

	// Calculate button positions
	titleY := startY + ObjectPalettePadding
	buttonY := titleY + 20

	// Check each button
	for i, schema := range p.schemas {
		y := buttonY + i*(ObjectButtonHeight+ObjectButtonSpacing)
		if screenY >= y && screenY < y+ObjectButtonHeight {
			p.selectedType = world.ObjectType(schema.Type)
			return true
		}
	}

	return false
}

// HandleMouseMove processes mouse movement for hover effects.
func (p *ObjectPalette) HandleMouseMove(screenX, screenY, screenWidth, startY int) {
	p.hoveredIndex = -1

	// Check if mouse is in palette area
	paletteX := screenWidth - ObjectPaletteWidth
	if screenX < paletteX {
		return
	}
	if screenY < startY {
		return
	}

	// Calculate button positions
	titleY := startY + ObjectPalettePadding
	buttonY := titleY + 20

	// Check each button
	for i := range p.schemas {
		y := buttonY + i*(ObjectButtonHeight+ObjectButtonSpacing)
		if screenY >= y && screenY < y+ObjectButtonHeight {
			p.hoveredIndex = i
			return
		}
	}
}

// IsInPalette returns true if the given screen coordinates are within the object palette area.
func (p *ObjectPalette) IsInPalette(screenX, screenY, screenWidth, startY int) bool {
	paletteX := screenWidth - ObjectPaletteWidth
	return screenX >= paletteX && screenY >= startY
}

// parseColor converts a hex color string to color.RGBA.
func parseColor(hex string) color.RGBA {
	if len(hex) == 0 || hex[0] != '#' {
		return color.RGBA{128, 128, 128, 255} // Default gray
	}

	hex = hex[1:] // Remove '#'
	if len(hex) != 6 {
		return color.RGBA{128, 128, 128, 255}
	}

	r, _ := strconv.ParseInt(hex[0:2], 16, 32)
	g, _ := strconv.ParseInt(hex[2:4], 16, 32)
	b, _ := strconv.ParseInt(hex[4:6], 16, 32)

	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

// Colors for object palette rendering
var (
	objectPaletteBgColor      = color.RGBA{50, 50, 60, 255}
	objectButtonColor         = color.RGBA{60, 60, 70, 255}
	objectButtonHoverColor    = color.RGBA{80, 80, 90, 255}
	objectButtonSelectedColor = color.RGBA{70, 100, 140, 255}
)
