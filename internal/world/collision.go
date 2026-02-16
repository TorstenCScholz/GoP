package world

// TileCoord represents a tile coordinate.
type TileCoord struct {
	X, Y int
}

// SolidGrid represents a grid of solid tiles for collision detection.
type SolidGrid struct {
	width  int    // Width in tiles
	height int    // Height in tiles
	data   []bool // true = solid, false = empty
}

// NewSolidGrid creates a new empty grid with the given dimensions.
// All tiles start as non-solid (false).
func NewSolidGrid(width, height int) *SolidGrid {
	return &SolidGrid{
		width:  width,
		height: height,
		data:   make([]bool, width*height),
	}
}

// NewSolidGridFromLayer creates a grid from a collision layer.
// Any non-zero tile ID is considered solid.
func NewSolidGridFromLayer(layer *TileLayer, width, height int) *SolidGrid {
	grid := NewSolidGrid(width, height)
	if layer == nil {
		return grid
	}

	for ty := 0; ty < height && ty < layer.height; ty++ {
		for tx := 0; tx < width && tx < layer.width; tx++ {
			tileID := layer.TileAt(tx, ty)
			if tileID != 0 {
				grid.SetSolid(tx, ty, true)
			}
		}
	}

	return grid
}

// Width returns the grid width in tiles.
func (g *SolidGrid) Width() int {
	return g.width
}

// Height returns the grid height in tiles.
func (g *SolidGrid) Height() int {
	return g.height
}

// IsSolid returns true if the tile at (tx, ty) is solid.
// Returns false for out-of-bounds coordinates.
func (g *SolidGrid) IsSolid(tx, ty int) bool {
	if tx < 0 || tx >= g.width || ty < 0 || ty >= g.height {
		return false
	}
	return g.data[ty*g.width+tx]
}

// SetSolid sets the solid state at the given tile coordinates.
// Does nothing if coordinates are out of bounds.
func (g *SolidGrid) SetSolid(tx, ty int, solid bool) {
	if tx < 0 || tx >= g.width || ty < 0 || ty >= g.height {
		return
	}
	g.data[ty*g.width+tx] = solid
}

// CollisionMap provides collision detection for a loaded map.
type CollisionMap struct {
	grid    *SolidGrid
	tileW   int
	tileH   int
}

// NewCollisionMap creates a collision map from a solid grid and tile dimensions.
func NewCollisionMap(grid *SolidGrid, tileW, tileH int) *CollisionMap {
	return &CollisionMap{
		grid:  grid,
		tileW: tileW,
		tileH: tileH,
	}
}

// NewCollisionMapFromMap creates a collision map from a Map's collision layer.
// If no collision layer exists, an empty collision map is returned.
func NewCollisionMapFromMap(m *Map, collisionLayerName string) *CollisionMap {
	grid := NewSolidGrid(m.Width(), m.Height())
	
	layer := m.Layer(collisionLayerName)
	if layer != nil {
		for ty := 0; ty < m.Height(); ty++ {
			for tx := 0; tx < m.Width(); tx++ {
				if layer.TileAt(tx, ty) != 0 {
					grid.SetSolid(tx, ty, true)
				}
			}
		}
	}

	return &CollisionMap{
		grid:  grid,
		tileW: m.TileWidth(),
		tileH: m.TileHeight(),
	}
}

// Grid returns the underlying solid grid.
func (c *CollisionMap) Grid() *SolidGrid {
	return c.grid
}

// TileWidth returns the tile width in pixels.
func (c *CollisionMap) TileWidth() int {
	return c.tileW
}

// TileHeight returns the tile height in pixels.
func (c *CollisionMap) TileHeight() int {
	return c.tileH
}

// IsSolidAtTile returns true if the tile at the given coordinates is solid.
// Returns false for out-of-bounds coordinates.
func (c *CollisionMap) IsSolidAtTile(tx, ty int) bool {
	return c.grid.IsSolid(tx, ty)
}

// IsSolidAtWorld returns true if the world position is in a solid tile.
// Returns false for out-of-bounds coordinates.
func (c *CollisionMap) IsSolidAtWorld(x, y float64) bool {
	tx := int(x) / c.tileW
	ty := int(y) / c.tileH
	return c.grid.IsSolid(tx, ty)
}

// OverlapsSolid checks if an AABB overlaps any solid tiles.
// The AABB is defined by top-left corner (x, y) and dimensions (w, h).
func (c *CollisionMap) OverlapsSolid(x, y, w, h float64) bool {
	// Get the tile range that the AABB overlaps
	tx1 := int(x) / c.tileW
	ty1 := int(y) / c.tileH
	// Use floor division to avoid including tiles at exact boundaries
	// Subtract a small epsilon to handle exact boundary cases
	tx2 := int(x+w-0.001) / c.tileW
	ty2 := int(y+h-0.001) / c.tileH

	// Check each tile in the range
	for ty := ty1; ty <= ty2; ty++ {
		for tx := tx1; tx <= tx2; tx++ {
			if c.grid.IsSolid(tx, ty) {
				return true
			}
		}
	}

	return false
}

// GetOverlappingTiles returns all tiles that overlap the given AABB.
// The AABB is defined by top-left corner (x, y) and dimensions (w, h).
func (c *CollisionMap) GetOverlappingTiles(x, y, w, h float64) []TileCoord {
	tx1 := int(x) / c.tileW
	ty1 := int(y) / c.tileH
	// Use floor division to avoid including tiles at exact boundaries
	tx2 := int(x+w-0.001) / c.tileW
	ty2 := int(y+h-0.001) / c.tileH

	var tiles []TileCoord
	for ty := ty1; ty <= ty2; ty++ {
		for tx := tx1; tx <= tx2; tx++ {
			tiles = append(tiles, TileCoord{X: tx, Y: ty})
		}
	}

	return tiles
}

// WorldToTile converts world coordinates to tile coordinates.
func WorldToTile(x, y float64, tileW, tileH int) (tx, ty int) {
	return int(x) / tileW, int(y) / tileH
}

// TileToWorld converts tile coordinates to world coordinates (top-left corner).
func TileToWorld(tx, ty, tileW, tileH int) (x, y float64) {
	return float64(tx * tileW), float64(ty * tileH)
}

// TileCenter returns the world position of a tile's center.
func TileCenter(tx, ty, tileW, tileH int) (x, y float64) {
	return float64(tx*tileW + tileW/2), float64(ty*tileH + tileH/2)
}
