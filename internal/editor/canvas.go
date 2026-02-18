package editor

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/torsten/GoP/internal/world"
)

// Canvas handles tilemap rendering and interaction for the editor.
type Canvas struct {
	state         *EditorState
	camera        *Camera
	tileset       *Tileset
	tools         *ToolManager
	showGrid      bool
	showCollision bool
	mousePressed  bool
	hoverHandle   HandlePosition    // Current handle being hovered
	validation    *ValidationResult // Current validation result
	hoveredTileX  int               // Currently hovered tile X coordinate
	hoveredTileY  int               // Currently hovered tile Y coordinate
}

// NewCanvas creates a new canvas for rendering the tilemap.
func NewCanvas(state *EditorState, camera *Camera, tileset *Tileset) *Canvas {
	c := &Canvas{
		state:         state,
		camera:        camera,
		tileset:       tileset,
		tools:         NewToolManager(),
		showGrid:      true,
		showCollision: true,
		mousePressed:  false,
		validation:    nil,
		hoveredTileX:  -1,
		hoveredTileY:  -1,
	}
	// Set the state reference for tools that need it
	c.tools.SetState(state)
	return c
}

// Draw renders the tilemap canvas to the screen.
// The canvas is drawn in the left portion of the screen (excluding the palette sidebar).
func (c *Canvas) Draw(screen *ebiten.Image) {
	if c.state == nil || !c.state.HasLevel() {
		return
	}

	// Get screen dimensions
	screenWidth, screenHeight := screen.Size()
	canvasWidth := screenWidth - PaletteWidth - ObjectPaletteWidth

	// Create a sub-image for the canvas area
	canvasBounds := screen.Bounds()
	canvasBounds.Max.X = canvasWidth

	// Draw tile layers
	c.drawTileLayers(screen, canvasWidth)

	// Draw collision overlay
	if c.showCollision {
		c.drawCollisionOverlay(screen, canvasWidth)
	}

	// Draw objects
	c.drawObjects(screen, canvasWidth)

	// Draw grid overlay
	if c.showGrid {
		c.drawGrid(screen, canvasWidth, screenHeight)
	}

	// Draw tool preview
	c.drawToolPreview(screen, canvasWidth)
}

// drawTileLayers renders all visible tile layers.
func (c *Canvas) drawTileLayers(screen *ebiten.Image, canvasWidth int) {
	if c.state.MapData == nil || c.tileset == nil || !c.tileset.IsLoaded() {
		return
	}

	tileW := c.state.MapData.TileWidth()
	tileH := c.state.MapData.TileHeight()

	// Calculate visible tile range based on camera
	camX := c.camera.X
	camY := c.camera.Y
	zoom := c.camera.Zoom

	// Calculate visible tiles
	tx1 := int(camX) / tileW
	ty1 := int(camY) / tileH
	tx2 := (int(camX) + int(float64(canvasWidth)/zoom) + tileW) / tileW
	ty2 := (int(camY) + int(float64(screen.Bounds().Dy())/zoom) + tileH) / tileH

	// Clamp to map bounds
	if tx1 < 0 {
		tx1 = 0
	}
	if ty1 < 0 {
		ty1 = 0
	}
	if tx2 > c.state.MapData.Width() {
		tx2 = c.state.MapData.Width()
	}
	if ty2 > c.state.MapData.Height() {
		ty2 = c.state.MapData.Height()
	}

	// Draw each layer
	for _, layer := range c.state.MapData.Layers() {
		// Skip collision layer for normal rendering
		if layer.Name() == "Collision" {
			continue
		}
		// Skip invisible layers
		if !c.state.IsLayerVisible(layer.Name()) {
			continue
		}
		c.drawLayer(screen, layer, tx1, ty1, tx2, ty2, tileW, tileH, camX, camY, zoom)
	}
}

// drawLayer renders a single tile layer.
func (c *Canvas) drawLayer(screen *ebiten.Image, layer *world.TileLayer, tx1, ty1, tx2, ty2, tileW, tileH int, camX, camY, zoom float64) {
	for ty := ty1; ty < ty2; ty++ {
		for tx := tx1; tx < tx2; tx++ {
			tileID := layer.TileAt(tx, ty)
			if tileID == 0 {
				continue // Empty tile
			}

			// Tiled uses 1-based IDs, convert to 0-based
			tile := c.tileset.Tile(tileID - 1)
			if tile == nil {
				continue
			}

			// Calculate screen position with camera transform
			worldX := float64(tx * tileW)
			worldY := float64(ty * tileH)
			screenX := (worldX - camX) * zoom
			screenY := (worldY - camY) * zoom

			// Draw tile
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(zoom, zoom)
			op.GeoM.Translate(screenX, screenY)
			op.Filter = ebiten.FilterNearest
			screen.DrawImage(tile, op)
		}
	}
}

// drawCollisionOverlay renders the collision layer as a semi-transparent overlay.
func (c *Canvas) drawCollisionOverlay(screen *ebiten.Image, canvasWidth int) {
	if c.state.MapData == nil {
		return
	}

	// Skip if collision layer is not visible
	if !c.state.IsLayerVisible("Collision") {
		return
	}

	collisionLayer := c.state.MapData.Layer("Collision")
	if collisionLayer == nil {
		return
	}

	tileW := c.state.MapData.TileWidth()
	tileH := c.state.MapData.TileHeight()
	camX := c.camera.X
	camY := c.camera.Y
	zoom := c.camera.Zoom

	// Calculate visible tiles
	tx1 := int(camX) / tileW
	ty1 := int(camY) / tileH
	tx2 := (int(camX) + int(float64(canvasWidth)/zoom) + tileW) / tileW
	ty2 := (int(camY) + int(float64(screen.Bounds().Dy())/zoom) + tileH) / tileH

	// Clamp to map bounds
	if tx1 < 0 {
		tx1 = 0
	}
	if ty1 < 0 {
		ty1 = 0
	}
	if tx2 > c.state.MapData.Width() {
		tx2 = c.state.MapData.Width()
	}
	if ty2 > c.state.MapData.Height() {
		ty2 = c.state.MapData.Height()
	}

	// Create collision overlay tile
	overlayTile := ebiten.NewImage(tileW, tileH)
	overlayTile.Fill(collisionOverlayColor)

	for ty := ty1; ty < ty2; ty++ {
		for tx := tx1; tx < tx2; tx++ {
			tileID := collisionLayer.TileAt(tx, ty)
			if tileID == 0 {
				continue // Non-solid tile
			}

			// Calculate screen position
			worldX := float64(tx * tileW)
			worldY := float64(ty * tileH)
			screenX := (worldX - camX) * zoom
			screenY := (worldY - camY) * zoom

			// Draw collision overlay
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(zoom, zoom)
			op.GeoM.Translate(screenX, screenY)
			screen.DrawImage(overlayTile, op)
		}
	}
}

// drawObjects renders all objects as colored rectangles.
func (c *Canvas) drawObjects(screen *ebiten.Image, canvasWidth int) {
	if c.state == nil {
		return
	}

	camX := c.camera.X
	camY := c.camera.Y
	zoom := c.camera.Zoom

	// Get selection manager for multi-select
	selection := c.state.GetSelectionManager()

	// First pass: draw switch/door links
	c.drawSwitchDoorLinks(screen, canvasWidth, camX, camY, zoom)

	// Second pass: draw platform paths
	c.drawPlatformPaths(screen, canvasWidth, camX, camY, zoom)

	// Third pass: draw objects
	for i, obj := range c.state.Objects {
		// Convert world coordinates to screen coordinates
		screenX := (obj.X - camX) * zoom
		screenY := (obj.Y - camY) * zoom
		w := obj.W * zoom
		h := obj.H * zoom

		// Skip if outside visible area
		if screenX+w < 0 || screenX >= float64(canvasWidth) ||
			screenY+h < 0 || screenY >= float64(screen.Bounds().Dy()) {
			continue
		}

		// Get color from schema
		schema := GetSchema(obj.Type)
		objColor := objectDefaultColor
		if schema != nil {
			objColor = parseColor(schema.Color)
		}

		// Draw object rectangle
		objImg := ebiten.NewImage(int(w), int(h))
		objImg.Fill(objColor)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(screenX, screenY)
		screen.DrawImage(objImg, op)

		// Draw border
		borderColor := darkerColor(objColor, 0.6)
		ebitenutil.DrawRect(screen, screenX, screenY, w, 2, borderColor)
		ebitenutil.DrawRect(screen, screenX, screenY+h-2, w, 2, borderColor)
		ebitenutil.DrawRect(screen, screenX, screenY, 2, h, borderColor)
		ebitenutil.DrawRect(screen, screenX+w-2, screenY, 2, h, borderColor)

		// Check if this object is selected (single or multi)
		isSelected := selection != nil && selection.IsSelected(i)

		// Highlight selected object
		if isSelected {
			// Use different highlight for multi-select
			selectionColor := color.RGBA{255, 255, 255, 200}
			if selection.SelectionCount() > 1 {
				selectionColor = color.RGBA{0, 255, 255, 200} // Cyan for multi-select
			}
			ebitenutil.DrawRect(screen, screenX-2, screenY-2, w+4, 2, selectionColor)
			ebitenutil.DrawRect(screen, screenX-2, screenY+h, w+4, 2, selectionColor)
			ebitenutil.DrawRect(screen, screenX-2, screenY-2, 2, h+4, selectionColor)
			ebitenutil.DrawRect(screen, screenX+w, screenY-2, 2, h+4, selectionColor)

			// Draw resize handles only for primary selection
			if selection.SelectedIndex() == i {
				c.drawSelectionHandles(screen, screenX, screenY, w, h, zoom)
			}
		}

		// Draw error indicator if object has validation errors
		if c.validation != nil {
			issues := c.validation.GetObjectIssues(i)
			if len(issues) > 0 {
				// Determine color based on error vs warning
				indicatorColor := validationWarningColor
				for _, issue := range issues {
					if issue.Type == TypeError {
						indicatorColor = validationErrorColor
						break
					}
				}

				// Draw error/warning border around the object
				borderWidth := 3.0
				ebitenutil.DrawRect(screen, screenX-borderWidth, screenY-borderWidth, w+2*borderWidth, borderWidth, indicatorColor)
				ebitenutil.DrawRect(screen, screenX-borderWidth, screenY+h, w+2*borderWidth, borderWidth, indicatorColor)
				ebitenutil.DrawRect(screen, screenX-borderWidth, screenY-borderWidth, borderWidth, h+2*borderWidth, indicatorColor)
				ebitenutil.DrawRect(screen, screenX+w, screenY-borderWidth, borderWidth, h+2*borderWidth, indicatorColor)

				// Draw error count badge
				badgeX := screenX + w - 16
				badgeY := screenY - 16
				if badgeX < 0 {
					badgeX = screenX
				}
				if badgeY < 0 {
					badgeY = screenY
				}
				badgeImg := ebiten.NewImage(20, 20)
				badgeImg.Fill(indicatorColor)
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(badgeX, badgeY)
				screen.DrawImage(badgeImg, op)

				// Draw issue count on badge
				countText := fmt.Sprintf("%d", len(issues))
				ebitenutil.DebugPrintAt(screen, countText, int(badgeX)+4, int(badgeY)+4)
			}
		}

		// Draw entity icon (first letter of type)
		c.drawEntityIcon(screen, obj, screenX, screenY, w, h, zoom)

		// Draw object type label
		if schema != nil && zoom >= 0.5 {
			labelX := int(screenX) + 4
			labelY := int(screenY) + 4
			ebitenutil.DebugPrintAt(screen, schema.Name, labelX, labelY)
		}
	}
}

// drawEntityIcon draws a type indicator icon on each entity.
func (c *Canvas) drawEntityIcon(screen *ebiten.Image, obj world.ObjectData, screenX, screenY, w, h, zoom float64) {
	// Get the first letter of the type
	var letter string
	switch obj.Type {
	case world.ObjectTypeSpawn:
		letter = "S"
	case world.ObjectTypePlatform:
		letter = "P"
	case world.ObjectTypeSwitch:
		letter = "W"
	case world.ObjectTypeDoor:
		letter = "D"
	case world.ObjectTypeHazard:
		letter = "H"
	case world.ObjectTypeCheckpoint:
		letter = "C"
	case world.ObjectTypeGoal:
		letter = "G"
	default:
		return
	}

	// Draw the letter in the center of the object
	if zoom >= 0.5 && w >= 16 && h >= 16 {
		// Center the letter
		letterX := int(screenX + w/2 - 4)
		letterY := int(screenY + h/2 - 6)
		ebitenutil.DebugPrintAt(screen, letter, letterX, letterY)
	}
}

// drawPlatformPaths draws movement paths for platforms.
func (c *Canvas) drawPlatformPaths(screen *ebiten.Image, canvasWidth int, camX, camY, zoom float64) {
	for _, obj := range c.state.Objects {
		if obj.Type != world.ObjectTypePlatform {
			continue
		}

		// Get endX and endY from properties
		endX := obj.GetPropFloat("endX", 0)
		endY := obj.GetPropFloat("endY", 0)

		// Skip if no movement
		if endX == 0 && endY == 0 {
			continue
		}

		// Calculate start and end screen positions
		startX := (obj.X - camX) * zoom
		startY := (obj.Y - camY) * zoom
		endScreenX := (obj.X + endX - camX) * zoom
		endScreenY := (obj.Y + endY - camY) * zoom

		// Draw dashed line from start to end
		c.drawDashedLine(screen, startX, startY, endScreenX, endScreenY, platformPathColor)

		// Draw endpoint marker (small square at destination)
		markerSize := 8.0
		markerX := endScreenX
		markerY := endScreenY

		// Draw marker rectangle
		markerImg := ebiten.NewImage(int(markerSize), int(markerSize))
		markerImg.Fill(platformPathColor)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(markerX-markerSize/2, markerY-markerSize/2)
		screen.DrawImage(markerImg, op)

		// Draw marker border
		ebitenutil.DrawRect(screen, markerX-markerSize/2, markerY-markerSize/2, markerSize, 1, color.RGBA{255, 255, 255, 200})
		ebitenutil.DrawRect(screen, markerX-markerSize/2, markerY+markerSize/2-1, markerSize, 1, color.RGBA{255, 255, 255, 200})
		ebitenutil.DrawRect(screen, markerX-markerSize/2, markerY-markerSize/2, 1, markerSize, color.RGBA{255, 255, 255, 200})
		ebitenutil.DrawRect(screen, markerX+markerSize/2-1, markerY-markerSize/2, 1, markerSize, color.RGBA{255, 255, 255, 200})
	}
}

// drawDashedLine draws a dashed line between two points.
func (c *Canvas) drawDashedLine(screen *ebiten.Image, x1, y1, x2, y2 float64, col color.Color) {
	// Calculate line length and direction
	dx := x2 - x1
	dy := y2 - y1
	length := sqrt(dx*dx + dy*dy)

	if length == 0 {
		return
	}

	// Normalize direction
	nx := dx / length
	ny := dy / length

	// Draw dashes
	dashLength := 8.0
	gapLength := 4.0
	pos := 0.0

	for pos < length {
		// Start of dash
		startX := x1 + nx*pos
		startY := y1 + ny*pos

		// End of dash
		endPos := pos + dashLength
		if endPos > length {
			endPos = length
		}
		endX := x1 + nx*endPos
		endY := y1 + ny*endPos

		// Draw the dash segment
		ebitenutil.DrawLine(screen, startX, startY, endX, endY, col)

		// Move to next dash
		pos += dashLength + gapLength
	}
}

// sqrt returns the square root of a float64.
func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	// Simple Newton's method for square root
	z := x
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

// drawSwitchDoorLinks draws connection lines between switches and their target doors.
func (c *Canvas) drawSwitchDoorLinks(screen *ebiten.Image, canvasWidth int, camX, camY, zoom float64) {
	// Build a map of door IDs to door objects
	doorMap := make(map[int]int) // maps door index to door id string
	for i, obj := range c.state.Objects {
		if obj.Type == world.ObjectTypeDoor {
			id := obj.GetPropString("id", "")
			if id != "" {
				doorMap[i] = int(hashString(id) % 1000)
			}
		}
	}

	// Draw links from switches to doors
	for switchIdx, obj := range c.state.Objects {
		if obj.Type != world.ObjectTypeSwitch {
			continue
		}

		targetID := obj.GetPropString("door_id", "")
		if targetID == "" {
			continue
		}

		// Find the target door
		var doorObj *world.ObjectData
		var doorIdx int
		for i, d := range c.state.Objects {
			if d.Type == world.ObjectTypeDoor {
				id := d.GetPropString("id", "")
				if id == targetID {
					doorObj = &c.state.Objects[i]
					doorIdx = i
					break
				}
			}
		}

		if doorObj == nil {
			continue
		}

		// Check if switch or door is selected
		selection := c.state.GetSelectionManager()
		isSelected := (selection != nil && (selection.IsSelected(switchIdx) || selection.IsSelected(doorIdx)))

		// Only draw if selected or if showing all links
		if !isSelected {
			continue
		}

		// Calculate screen positions
		switchCenterX := (obj.X + obj.W/2 - camX) * zoom
		switchCenterY := (obj.Y + obj.H/2 - camY) * zoom
		doorCenterX := (doorObj.X + doorObj.W/2 - camX) * zoom
		doorCenterY := (doorObj.Y + doorObj.H/2 - camY) * zoom

		// Generate a color based on the door ID
		linkColor := generateLinkColor(targetID)

		// Draw the connection line
		ebitenutil.DrawLine(screen, switchCenterX, switchCenterY, doorCenterX, doorCenterY, linkColor)
	}
}

// hashString generates a simple hash from a string.
func hashString(s string) uint32 {
	var h uint32
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return h
}

// generateLinkColor generates a color based on a string ID.
func generateLinkColor(id string) color.Color {
	// Generate different colors for different connections
	colors := []color.Color{
		color.RGBA{255, 100, 100, 200}, // Red
		color.RGBA{100, 255, 100, 200}, // Green
		color.RGBA{100, 100, 255, 200}, // Blue
		color.RGBA{255, 255, 100, 200}, // Yellow
		color.RGBA{255, 100, 255, 200}, // Magenta
		color.RGBA{100, 255, 255, 200}, // Cyan
		color.RGBA{255, 165, 0, 200},   // Orange
		color.RGBA{128, 0, 255, 200},   // Purple
	}

	h := hashString(id)
	return colors[h%uint32(len(colors))]
}

// drawSelectionHandles draws the 8 resize handles on a selected object.
func (c *Canvas) drawSelectionHandles(screen *ebiten.Image, screenX, screenY, w, h, zoom float64) {
	handleSize := 8.0 // 8x8 pixel handles
	hs := handleSize / 2.0

	// Handle colors
	handleColor := color.RGBA{255, 255, 255, 255}
	handleBorder := color.RGBA{0, 0, 0, 255}

	// Handle positions (screen coordinates)
	handles := []struct {
		x, y float64
	}{
		// Corners
		{screenX - hs, screenY - hs},         // Top-left
		{screenX + w - hs, screenY - hs},     // Top-right
		{screenX - hs, screenY + h - hs},     // Bottom-left
		{screenX + w - hs, screenY + h - hs}, // Bottom-right
		// Edges
		{screenX + w/2 - hs, screenY - hs},     // Top
		{screenX + w/2 - hs, screenY + h - hs}, // Bottom
		{screenX - hs, screenY + h/2 - hs},     // Left
		{screenX + w - hs, screenY + h/2 - hs}, // Right
	}

	for _, handle := range handles {
		// Draw handle background
		handleImg := ebiten.NewImage(int(handleSize), int(handleSize))
		handleImg.Fill(handleColor)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(handle.x, handle.y)
		screen.DrawImage(handleImg, op)

		// Draw handle border
		ebitenutil.DrawRect(screen, handle.x, handle.y, handleSize, 1, handleBorder)
		ebitenutil.DrawRect(screen, handle.x, handle.y+handleSize-1, handleSize, 1, handleBorder)
		ebitenutil.DrawRect(screen, handle.x, handle.y, 1, handleSize, handleBorder)
		ebitenutil.DrawRect(screen, handle.x+handleSize-1, handle.y, 1, handleSize, handleBorder)
	}
}

// drawGrid renders the tile grid overlay.
func (c *Canvas) drawGrid(screen *ebiten.Image, canvasWidth, screenHeight int) {
	if c.state.MapData == nil {
		return
	}

	tileW := c.state.MapData.TileWidth()
	tileH := c.state.MapData.TileHeight()
	camX := c.camera.X
	camY := c.camera.Y
	zoom := c.camera.Zoom

	// Get map bounds
	mapWidth := c.state.MapData.Width()
	mapHeight := c.state.MapData.Height()

	// Calculate visible tile range
	startTileX := int(camX) / tileW
	startTileY := int(camY) / tileH
	endTileX := startTileX + int(float64(canvasWidth)/float64(tileW)/zoom) + 2
	endTileY := startTileY + int(float64(screenHeight)/float64(tileH)/zoom) + 2

	// Clamp to map bounds
	if startTileX < 0 {
		startTileX = 0
	}
	if startTileY < 0 {
		startTileY = 0
	}
	if endTileX > mapWidth {
		endTileX = mapWidth
	}
	if endTileY > mapHeight {
		endTileY = mapHeight
	}

	// Create grid line image
	gridLine := ebiten.NewImage(1, 1)
	gridLine.Fill(gridColor)

	// Calculate the screen position of the map boundaries
	mapRightWorld := float64(mapWidth * tileW)
	mapBottomWorld := float64(mapHeight * tileH)
	mapRightScreen := (mapRightWorld - camX) * zoom
	mapBottomScreen := (mapBottomWorld - camY) * zoom

	// Draw vertical lines (only within map bounds)
	for tx := startTileX; tx <= endTileX; tx++ {
		worldX := float64(tx * tileW)
		screenX := (worldX - camX) * zoom

		if screenX < 0 || screenX >= float64(canvasWidth) {
			continue
		}

		// Calculate line height (don't extend beyond map bottom)
		lineTop := 0.0
		lineBottom := float64(screenHeight)
		if mapBottomScreen < lineBottom {
			lineBottom = mapBottomScreen
		}
		if lineBottom <= 0 {
			continue
		}

		lineHeight := lineBottom - lineTop
		if lineHeight > 0 {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(1, lineHeight)
			op.GeoM.Translate(screenX, lineTop)
			screen.DrawImage(gridLine, op)
		}
	}

	// Draw horizontal lines (only within map bounds)
	for ty := startTileY; ty <= endTileY; ty++ {
		worldY := float64(ty * tileH)
		screenY := (worldY - camY) * zoom

		if screenY < 0 || screenY >= float64(screenHeight) {
			continue
		}

		// Calculate line width (don't extend beyond map right edge)
		lineLeft := 0.0
		lineRight := float64(canvasWidth)
		if mapRightScreen < lineRight {
			lineRight = mapRightScreen
		}
		if lineRight <= 0 {
			continue
		}

		lineWidth := lineRight - lineLeft
		if lineWidth > 0 {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(lineWidth, 1)
			op.GeoM.Translate(lineLeft, screenY)
			screen.DrawImage(gridLine, op)
		}
	}
}

// drawToolPreview renders a preview of the selected tile under the cursor.
func (c *Canvas) drawToolPreview(screen *ebiten.Image, canvasWidth int) {
	// Only show preview for paint tool with a selected tile
	if c.state.CurrentTool != ToolPaint || c.state.SelectedTile < 0 || c.tileset == nil || !c.tileset.IsLoaded() {
		return
	}

	// Get cursor position
	mx, my := ebiten.CursorPosition()

	// Check if cursor is in canvas area (not in palette)
	if mx >= canvasWidth {
		return
	}

	// Convert to world coordinates
	worldX, worldY := c.camera.ScreenToWorld(mx, my)

	// Convert to tile coordinates
	tileW := c.state.MapData.TileWidth()
	tileH := c.state.MapData.TileHeight()
	tileX := int(worldX) / tileW
	tileY := int(worldY) / tileH

	// Check if within map bounds
	if tileX < 0 || tileX >= c.state.MapData.Width() || tileY < 0 || tileY >= c.state.MapData.Height() {
		return
	}

	// Get the selected tile image
	tile := c.tileset.Tile(c.state.SelectedTile)
	if tile == nil {
		return
	}

	// Calculate screen position for preview
	previewWorldX := float64(tileX * tileW)
	previewWorldY := float64(tileY * tileH)
	screenX := (previewWorldX - c.camera.X) * c.camera.Zoom
	screenY := (previewWorldY - c.camera.Y) * c.camera.Zoom

	// Draw preview tile with transparency
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(c.camera.Zoom, c.camera.Zoom)
	op.GeoM.Translate(screenX, screenY)
	op.ColorM.Scale(1, 1, 1, 0.5) // Semi-transparent
	op.Filter = ebiten.FilterNearest
	screen.DrawImage(tile, op)
}

// Update handles input for the canvas.
func (c *Canvas) Update() error {
	// Toggle grid with G key
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		c.showGrid = !c.showGrid
	}

	// Toggle collision overlay with C key
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		c.showCollision = !c.showCollision
	}

	// Handle tool input
	c.handleToolInput()

	return nil
}

// handleToolInput processes mouse input for the current tool.
func (c *Canvas) handleToolInput() {
	if c.state == nil || !c.state.HasLevel() {
		return
	}

	// Get cursor position
	mx, my := ebiten.CursorPosition()

	// Check if cursor is in canvas area (not in palette)
	screenWidth := 1280 // Default, will be overridden by actual screen size
	if c.state.Zoom > 0 {
		// Approximate canvas width check (account for both palettes)
		canvasWidth := screenWidth - PaletteWidth - ObjectPaletteWidth
		if mx >= canvasWidth {
			c.hoveredTileX = -1
			c.hoveredTileY = -1
			return
		}
	}

	// Convert screen coordinates to world coordinates
	worldX, worldY := c.camera.ScreenToWorld(mx, my)

	// Convert to tile coordinates
	tileW := c.state.MapData.TileWidth()
	tileH := c.state.MapData.TileHeight()
	tileX := int(worldX) / tileW
	tileY := int(worldY) / tileH

	// Update hovered tile position (clamped to map bounds)
	mapWidth := c.state.MapData.Width()
	mapHeight := c.state.MapData.Height()
	if tileX >= 0 && tileX < mapWidth && tileY >= 0 && tileY < mapHeight {
		c.hoveredTileX = tileX
		c.hoveredTileY = tileY
	} else {
		c.hoveredTileX = -1
		c.hoveredTileY = -1
	}

	// Update hover handle for cursor display (only in select mode)
	if c.state.CurrentTool == ToolSelect {
		selectTool := c.tools.SelectTool()
		if selectTool != nil {
			c.hoverHandle = selectTool.GetHoverHandle(c.state, worldX, worldY)
		}
	} else {
		c.hoverHandle = HandleNone
	}

	// Handle mouse button events
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		c.mousePressed = true
		c.tools.HandleMouseDown(c.state, tileX, tileY, worldX, worldY)
	}

	if c.mousePressed && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		c.tools.HandleMouseMove(c.state, tileX, tileY, worldX, worldY)
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		c.mousePressed = false
		c.tools.HandleMouseUp(c.state, tileX, tileY, worldX, worldY)
	}
}

// ScreenToTile converts screen coordinates to tile coordinates.
func (c *Canvas) ScreenToTile(screenX, screenY int) (int, int) {
	if c.state == nil || !c.state.HasLevel() {
		return -1, -1
	}

	worldX, worldY := c.camera.ScreenToWorld(screenX, screenY)
	tileW := c.state.MapData.TileWidth()
	tileH := c.state.MapData.TileHeight()

	return int(worldX) / tileW, int(worldY) / tileH
}

// IsInCanvas returns true if the screen coordinates are within the canvas area.
func (c *Canvas) IsInCanvas(screenX, screenWidth int) bool {
	return screenX < screenWidth-PaletteWidth
}

// ShowGrid returns whether the grid is visible.
func (c *Canvas) ShowGrid() bool {
	return c.showGrid
}

// SetShowGrid sets the grid visibility.
func (c *Canvas) SetShowGrid(show bool) {
	c.showGrid = show
}

// ShowCollision returns whether the collision overlay is visible.
func (c *Canvas) ShowCollision() bool {
	return c.showCollision
}

// SetShowCollision sets the collision overlay visibility.
func (c *Canvas) SetShowCollision(show bool) {
	c.showCollision = show
}

// HoverHandle returns the current handle being hovered.
func (c *Canvas) HoverHandle() HandlePosition {
	return c.hoverHandle
}

// SetValidation sets the current validation result for error display.
func (c *Canvas) SetValidation(result *ValidationResult) {
	c.validation = result
}

// Validation returns the current validation result.
func (c *Canvas) Validation() *ValidationResult {
	return c.validation
}

// HoveredTile returns the currently hovered tile coordinates.
// Returns -1, -1 if no valid tile is hovered.
func (c *Canvas) HoveredTile() (int, int) {
	return c.hoveredTileX, c.hoveredTileY
}

// Colors for canvas rendering
var (
	gridColor              = color.RGBA{100, 100, 100, 100}
	collisionOverlayColor  = color.RGBA{255, 0, 0, 80}
	objectDefaultColor     = color.RGBA{128, 128, 128, 255}
	validationErrorColor   = color.RGBA{255, 0, 0, 255}    // Red for errors
	validationWarningColor = color.RGBA{255, 165, 0, 255}  // Orange for warnings
	platformPathColor      = color.RGBA{128, 64, 192, 200} // Purple for platform paths
)

// darkerColor returns a darker version of the given color.
func darkerColor(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * factor),
		G: uint8(float64(c.G) * factor),
		B: uint8(float64(c.B) * factor),
		A: c.A,
	}
}
