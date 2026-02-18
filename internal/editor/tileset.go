package editor

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/torsten/GoP/internal/assets"
	"github.com/torsten/GoP/internal/world"
)

const (
	// PaletteWidth is the width of the tile palette sidebar in pixels.
	PaletteWidth = 200
	// PalettePadding is the padding inside the palette area in pixels.
	PalettePadding = 10
	// PaletteTileSize is the display size of tiles in the palette (zoomed for visibility).
	PaletteTileSize = 32
)

// Tileset handles loading and rendering the tileset for the editor.
type Tileset struct {
	tileset    *world.Tileset
	paletteImg *ebiten.Image // Pre-rendered palette image
}

// NewTileset loads the tileset from the assets directory.
func NewTileset() *Tileset {
	// Load tileset as raw image for pixel access (needed before game loop starts)
	rawImg, err := assets.LoadTilesetRaw()
	if err != nil {
		log.Printf("Failed to load tileset: %v", err)
		return &Tileset{}
	}

	// Create world.Tileset (16x16 pixel tiles)
	ts := world.NewTilesetFromImage(rawImg, DefaultTileSize, DefaultTileSize)

	t := &Tileset{
		tileset: ts,
	}

	// Pre-render the palette
	t.paletteImg = t.createPaletteImage()

	return t
}

// createPaletteImage creates a pre-rendered image of all tiles for the palette.
func (t *Tileset) createPaletteImage() *ebiten.Image {
	if t.tileset == nil {
		return nil
	}

	// Calculate palette dimensions
	tileCount := t.tileset.TileCount()
	columns := t.tileset.Columns()
	rows := (tileCount + columns - 1) / columns

	// Create image for the palette
	paletteWidth := columns * PaletteTileSize
	paletteHeight := rows * PaletteTileSize
	palette := ebiten.NewImage(paletteWidth, paletteHeight)

	// Draw each tile
	for i := 0; i < tileCount; i++ {
		tile := t.tileset.Tile(i)
		if tile == nil {
			continue
		}

		// Calculate position in palette
		tx := i % columns
		ty := i / columns
		x := tx * PaletteTileSize
		y := ty * PaletteTileSize

		// Draw tile scaled up
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(PaletteTileSize)/float64(DefaultTileSize), float64(PaletteTileSize)/float64(DefaultTileSize))
		op.GeoM.Translate(float64(x), float64(y))
		op.Filter = ebiten.FilterNearest
		palette.DrawImage(tile, op)
	}

	return palette
}

// Tileset returns the underlying world.Tileset.
func (t *Tileset) Tileset() *world.Tileset {
	return t.tileset
}

// IsLoaded returns true if the tileset is loaded.
func (t *Tileset) IsLoaded() bool {
	return t.tileset != nil
}

// TileCount returns the number of tiles in the tileset.
func (t *Tileset) TileCount() int {
	if t.tileset == nil {
		return 0
	}
	return t.tileset.TileCount()
}

// Tile returns the tile image at the given ID (0-indexed).
func (t *Tileset) Tile(id int) *ebiten.Image {
	if t.tileset == nil {
		return nil
	}
	return t.tileset.Tile(id)
}

// DrawPalette renders the tile palette to the screen.
// The palette is drawn in the right sidebar area.
// selectedTile is the currently selected tile ID (-1 if none).
func (t *Tileset) DrawPalette(screen *ebiten.Image, selectedTile int) {
	if t.paletteImg == nil {
		return
	}

	// Get screen dimensions
	screenWidth, _ := screen.Size()

	// Calculate palette position (right side of screen)
	paletteX := screenWidth - PaletteWidth - ObjectPaletteWidth

	t.DrawPaletteAt(screen, selectedTile, paletteX)
}

// DrawPaletteAt renders the tile palette at a specific X position.
func (t *Tileset) DrawPaletteAt(screen *ebiten.Image, selectedTile int, paletteX int) {
	if t.paletteImg == nil {
		return
	}

	// Draw palette background
	paletteBg := ebiten.NewImage(PaletteWidth, screen.Bounds().Dy())
	paletteBg.Fill(paletteBgColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(paletteX), 0)
	screen.DrawImage(paletteBg, op)

	// Draw the pre-rendered palette image
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(paletteX+PalettePadding), float64(PalettePadding))
	screen.DrawImage(t.paletteImg, op)

	// Draw selection highlight
	if selectedTile >= 0 && selectedTile < t.TileCount() {
		t.drawSelectionHighlight(screen, selectedTile, paletteX)
	}
}

// drawSelectionHighlight draws a highlight around the selected tile.
func (t *Tileset) drawSelectionHighlight(screen *ebiten.Image, selectedTile int, paletteX int) {
	columns := t.tileset.Columns()
	tx := selectedTile % columns
	ty := selectedTile / columns

	// Calculate screen position
	x := paletteX + PalettePadding + tx*PaletteTileSize
	y := PalettePadding + ty*PaletteTileSize

	// Draw highlight rectangle
	highlight := ebiten.NewImage(PaletteTileSize, PaletteTileSize)
	highlight.Fill(selectionColor)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorM.Scale(1, 1, 1, 0.3) // Semi-transparent
	screen.DrawImage(highlight, op)
}

// TileAtPosition returns the tile ID at the given screen position within the palette.
// Returns -1 if the position is outside the palette area or no tile at that position.
func (t *Tileset) TileAtPosition(screenX, screenY int) int {
	if t.tileset == nil {
		return -1
	}

	// Get screen dimensions (we need to know the palette start position)
	// This will be calculated by the caller and passed in
	// For now, assume the caller handles this

	columns := t.tileset.Columns()
	tileCount := t.tileset.TileCount()

	// Calculate tile position
	tx := (screenX - PalettePadding) / PaletteTileSize
	ty := (screenY - PalettePadding) / PaletteTileSize

	// Validate position
	if tx < 0 || ty < 0 {
		return -1
	}

	// Calculate tile ID
	tileID := ty*columns + tx
	if tileID >= tileCount {
		return -1
	}

	return tileID
}

// PaletteRect returns the screen rectangle (x, y, width, height) of the palette area.
func (t *Tileset) PaletteRect(screenWidth int) (x, y, w, h int) {
	x = screenWidth - PaletteWidth
	y = 0
	w = PaletteWidth
	h = -1 // Full height (caller should use screen height)
	return
}

// IsInPalette returns true if the given screen coordinates are within the palette area.
func (t *Tileset) IsInPalette(screenX, screenY, screenWidth int) bool {
	paletteX := screenWidth - PaletteWidth - ObjectPaletteWidth
	return screenX >= paletteX && screenX < paletteX+PaletteWidth
}

// Colors for palette rendering
var (
	paletteBgColor = color.RGBA{40, 40, 50, 255}
	selectionColor = color.RGBA{100, 200, 255, 255}
)
