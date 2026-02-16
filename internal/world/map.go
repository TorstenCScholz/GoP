// Package world provides tilemap loading and collision detection.
package world

import (
	"encoding/json"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Tileset represents a tileset image with sliced tiles.
type Tileset struct {
	image      *ebiten.Image
	tileWidth  int
	tileHeight int
	tiles      []*ebiten.Image // Pre-sliced tile images
	columns    int             // Number of columns in tileset
}

// NewTilesetFromImage creates a tileset from an image with uniform tile size.
// The image is sliced into tiles of the specified dimensions.
func NewTilesetFromImage(img image.Image, tileW, tileH int) *Tileset {
	ebitenImg := ebiten.NewImageFromImage(img)
	bounds := img.Bounds()
	imgW, imgH := bounds.Dx(), bounds.Dy()

	columns := imgW / tileW
	rows := imgH / tileH
	totalTiles := columns * rows

	tiles := make([]*ebiten.Image, totalTiles)
	for i := 0; i < totalTiles; i++ {
		tx := i % columns
		ty := i / columns
		x := tx * tileW
		y := ty * tileH
		tile := ebitenImg.SubImage(image.Rect(x, y, x+tileW, y+tileH)).(*ebiten.Image)
		tiles[i] = tile
	}

	return &Tileset{
		image:      ebitenImg,
		tileWidth:  tileW,
		tileHeight: tileH,
		tiles:      tiles,
		columns:    columns,
	}
}

// Tile returns the tile image at the given local tile ID (0-indexed).
// Returns nil if the ID is out of range.
func (t *Tileset) Tile(id int) *ebiten.Image {
	if id < 0 || id >= len(t.tiles) {
		return nil
	}
	return t.tiles[id]
}

// TileCount returns the total number of tiles in the tileset.
func (t *Tileset) TileCount() int {
	return len(t.tiles)
}

// TileWidth returns the tile width in pixels.
func (t *Tileset) TileWidth() int {
	return t.tileWidth
}

// TileHeight returns the tile height in pixels.
func (t *Tileset) TileHeight() int {
	return t.tileHeight
}

// Columns returns the number of columns in the tileset.
func (t *Tileset) Columns() int {
	return t.columns
}

// TileLayer represents a single tile layer from the map.
type TileLayer struct {
	name   string
	width  int
	height int
	data   []int // Global tile IDs, 0 = empty
}

// Name returns the layer name.
func (l *TileLayer) Name() string {
	return l.name
}

// Width returns the layer width in tiles.
func (l *TileLayer) Width() int {
	return l.width
}

// Height returns the layer height in tiles.
func (l *TileLayer) Height() int {
	return l.height
}

// TileAt returns the tile ID at the given tile coordinates.
// Returns 0 if coordinates are out of bounds.
func (l *TileLayer) TileAt(tx, ty int) int {
	if tx < 0 || tx >= l.width || ty < 0 || ty >= l.height {
		return 0
	}
	return l.data[ty*l.width+tx]
}

// SetTile sets the tile ID at the given tile coordinates.
// Does nothing if coordinates are out of bounds.
func (l *TileLayer) SetTile(tx, ty, id int) {
	if tx < 0 || tx >= l.width || ty < 0 || ty >= l.height {
		return
	}
	l.data[ty*l.width+tx] = id
}

// Data returns the raw tile data array.
func (l *TileLayer) Data() []int {
	return l.data
}

// MapData represents the parsed Tiled JSON data before tileset association.
type MapData struct {
	width      int
	height     int
	tileWidth  int
	tileHeight int
	layers     []*TileLayer
	layerIndex map[string]int
}

// Width returns the map width in tiles.
func (m *MapData) Width() int {
	return m.width
}

// Height returns the map height in tiles.
func (m *MapData) Height() int {
	return m.height
}

// TileWidth returns the tile width in pixels.
func (m *MapData) TileWidth() int {
	return m.tileWidth
}

// TileHeight returns the tile height in pixels.
func (m *MapData) TileHeight() int {
	return m.tileHeight
}

// Layer returns the layer by name, or nil if not found.
func (m *MapData) Layer(name string) *TileLayer {
	idx, ok := m.layerIndex[name]
	if !ok {
		return nil
	}
	return m.layers[idx]
}

// Layers returns all layers.
func (m *MapData) Layers() []*TileLayer {
	return m.layers
}

// tiledMap represents the JSON structure from Tiled.
type tiledMap struct {
	Width      int            `json:"width"`
	Height     int            `json:"height"`
	TileWidth  int            `json:"tilewidth"`
	TileHeight int            `json:"tileheight"`
	Layers     []tiledLayer   `json:"layers"`
	Tilesets   []tiledTileset `json:"tilesets"`
}

type tiledLayer struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Data   []int  `json:"data"`
}

type tiledTileset struct {
	FirstGID    int    `json:"firstgid"`
	Image       string `json:"image"`
	TileWidth   int    `json:"tilewidth"`
	TileHeight  int    `json:"tileheight"`
	ImageWidth  int    `json:"imagewidth"`
	ImageHeight int    `json:"imageheight"`
}

// ParseTiledJSON parses a Tiled JSON export into MapData.
// The MapData can then be used with a tileset to create a renderable Map.
func ParseTiledJSON(data []byte) (*MapData, error) {
	var tm tiledMap
	if err := json.Unmarshal(data, &tm); err != nil {
		return nil, fmt.Errorf("failed to parse Tiled JSON: %w", err)
	}

	mapData := &MapData{
		width:      tm.Width,
		height:     tm.Height,
		tileWidth:  tm.TileWidth,
		tileHeight: tm.TileHeight,
		layers:     make([]*TileLayer, 0, len(tm.Layers)),
		layerIndex: make(map[string]int),
	}

	for i, tl := range tm.Layers {
		// Skip non-tile layers (e.g., object layers)
		if tl.Type != "tilelayer" {
			continue
		}

		// Use map dimensions if layer dimensions are not set
		layerW := tl.Width
		layerH := tl.Height
		if layerW == 0 {
			layerW = tm.Width
		}
		if layerH == 0 {
			layerH = tm.Height
		}

		layer := &TileLayer{
			name:   tl.Name,
			width:  layerW,
			height: layerH,
			data:   tl.Data,
		}

		mapData.layerIndex[tl.Name] = len(mapData.layers)
		mapData.layers = append(mapData.layers, layer)

		// Store the original index for reference
		_ = i
	}

	return mapData, nil
}

// Map represents a loaded tilemap with all layers and tileset.
type Map struct {
	width      int
	height     int
	tileWidth  int
	tileHeight int
	layers     []*TileLayer
	tileset    *Tileset
	layerIndex map[string]int
}

// NewMap creates a new map from MapData and a tileset.
func NewMap(mapData *MapData, tileset *Tileset) *Map {
	return &Map{
		width:      mapData.width,
		height:     mapData.height,
		tileWidth:  mapData.tileWidth,
		tileHeight: mapData.tileHeight,
		layers:     mapData.layers,
		tileset:    tileset,
		layerIndex: mapData.layerIndex,
	}
}

// Width returns the map width in tiles.
func (m *Map) Width() int {
	return m.width
}

// Height returns the map height in tiles.
func (m *Map) Height() int {
	return m.height
}

// TileWidth returns the tile width in pixels.
func (m *Map) TileWidth() int {
	return m.tileWidth
}

// TileHeight returns the tile height in pixels.
func (m *Map) TileHeight() int {
	return m.tileHeight
}

// PixelWidth returns the map width in pixels.
func (m *Map) PixelWidth() int {
	return m.width * m.tileWidth
}

// PixelHeight returns the map height in pixels.
func (m *Map) PixelHeight() int {
	return m.height * m.tileHeight
}

// Layer returns the layer by name, or nil if not found.
func (m *Map) Layer(name string) *TileLayer {
	idx, ok := m.layerIndex[name]
	if !ok {
		return nil
	}
	return m.layers[idx]
}

// Layers returns all layers.
func (m *Map) Layers() []*TileLayer {
	return m.layers
}

// Tileset returns the map's tileset.
func (m *Map) Tileset() *Tileset {
	return m.tileset
}

// AddLayer adds a new layer to the map.
func (m *Map) AddLayer(name string, data []int) *TileLayer {
	layer := &TileLayer{
		name:   name,
		width:  m.width,
		height: m.height,
		data:   data,
	}
	m.layerIndex[name] = len(m.layers)
	m.layers = append(m.layers, layer)
	return layer
}
