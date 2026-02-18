// Package editor provides the level editor functionality.
package editor

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/torsten/GoP/internal/world"
)

// ToolHandler defines the interface for editor tools.
type ToolHandler interface {
	// OnMouseDown is called when the mouse button is pressed.
	OnMouseDown(state *EditorState, tileX, tileY int, worldX, worldY float64)
	// OnMouseMove is called when the mouse is moved while pressed.
	OnMouseMove(state *EditorState, tileX, tileY int, worldX, worldY float64)
	// OnMouseUp is called when the mouse button is released.
	OnMouseUp(state *EditorState, tileX, tileY int, worldX, worldY float64)
}

// PaintTool handles painting tiles on the active layer.
type PaintTool struct {
	lastTileX, lastTileY int
	hasLast              bool
	changes              []TileChange // Collect changes during drag
}

// NewPaintTool creates a new paint tool.
func NewPaintTool() *PaintTool {
	return &PaintTool{}
}

// OnMouseDown paints a tile at the clicked position.
func (t *PaintTool) OnMouseDown(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	t.changes = nil // Reset changes for new operation
	t.recordPaintTile(state, tileX, tileY)
	t.lastTileX = tileX
	t.lastTileY = tileY
	t.hasLast = true
}

// OnMouseMove paints tiles while dragging.
func (t *PaintTool) OnMouseMove(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	if !t.hasLast {
		return
	}

	// Paint in a line from last position to current for smooth dragging
	t.drawLine(state, t.lastTileX, t.lastTileY, tileX, tileY)
	t.lastTileX = tileX
	t.lastTileY = tileY
}

// OnMouseUp finalizes the paint operation.
func (t *PaintTool) OnMouseUp(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	t.hasLast = false

	// Create a single action for all tile changes
	if len(t.changes) > 0 {
		action := NewPaintTilesAction(state.CurrentLayer, t.changes)
		state.History.Do(action, state)
		t.changes = nil
	}
}

// recordPaintTile records a single tile paint operation.
func (t *PaintTool) recordPaintTile(state *EditorState, tileX, tileY int) {
	if state.MapData == nil {
		return
	}

	// For Tiles layer, require a selected tile
	if state.CurrentLayer != "Collision" && state.SelectedTile < 0 {
		return
	}

	layer := state.MapData.Layer(state.CurrentLayer)
	if layer == nil {
		return
	}

	// Check bounds
	if tileX < 0 || tileX >= layer.Width() || tileY < 0 || tileY >= layer.Height() {
		return
	}

	// Determine tile ID based on layer type
	var tileID int
	if state.CurrentLayer == "Collision" {
		// Collision layer: use 1 for solid, 0 for empty
		if state.SelectedCollision {
			tileID = 1
		} else {
			tileID = 0
		}
	} else {
		// Tiles layer: use Tiled's 1-based tile IDs
		tileID = state.SelectedTile + 1 // Convert 0-based to 1-based
	}

	// Record the old value before changing
	oldTileID := layer.TileAt(tileX, tileY)

	// Only record if the tile is actually changing
	if oldTileID != tileID {
		// Apply the change immediately for visual feedback
		layer.SetTile(tileX, tileY, tileID)

		// Record the change for the action
		t.changes = append(t.changes, TileChange{
			TileX:     tileX,
			TileY:     tileY,
			OldTileID: oldTileID,
			NewTileID: tileID,
		})
	}
}

// drawLine draws a line of tiles using Bresenham's algorithm.
func (t *PaintTool) drawLine(state *EditorState, x0, y0, x1, y1 int) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	sy := 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	for {
		t.recordPaintTile(state, x0, y0)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// EraseTool handles erasing tiles from the active layer.
type EraseTool struct {
	lastTileX, lastTileY int
	hasLast              bool
	changes              []TileChange // Collect changes during drag
}

// NewEraseTool creates a new erase tool.
func NewEraseTool() *EraseTool {
	return &EraseTool{}
}

// OnMouseDown erases a tile at the clicked position.
func (t *EraseTool) OnMouseDown(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	t.changes = nil // Reset changes for new operation
	t.recordEraseTile(state, tileX, tileY)
	t.lastTileX = tileX
	t.lastTileY = tileY
	t.hasLast = true
}

// OnMouseMove erases tiles while dragging.
func (t *EraseTool) OnMouseMove(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	if !t.hasLast {
		return
	}

	// Erase in a line from last position to current for smooth dragging
	t.drawLine(state, t.lastTileX, t.lastTileY, tileX, tileY)
	t.lastTileX = tileX
	t.lastTileY = tileY
}

// OnMouseUp finalizes the erase operation.
func (t *EraseTool) OnMouseUp(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	t.hasLast = false

	// Create a single action for all tile changes
	if len(t.changes) > 0 {
		action := NewEraseTilesAction(state.CurrentLayer, t.changes)
		state.History.Do(action, state)
		t.changes = nil
	}
}

// recordEraseTile records a single tile erase operation.
func (t *EraseTool) recordEraseTile(state *EditorState, tileX, tileY int) {
	if state.MapData == nil {
		return
	}

	layer := state.MapData.Layer(state.CurrentLayer)
	if layer == nil {
		return
	}

	// Check bounds
	if tileX < 0 || tileX >= layer.Width() || tileY < 0 || tileY >= layer.Height() {
		return
	}

	// Record the old value before erasing
	oldTileID := layer.TileAt(tileX, tileY)

	// Only record if the tile is not already empty
	if oldTileID != 0 {
		// Apply the change immediately for visual feedback
		layer.SetTile(tileX, tileY, 0)

		// Record the change for the action
		t.changes = append(t.changes, TileChange{
			TileX:     tileX,
			TileY:     tileY,
			OldTileID: oldTileID,
			NewTileID: 0,
		})
	}
}

// drawLine draws a line of erasure using Bresenham's algorithm.
func (t *EraseTool) drawLine(state *EditorState, x0, y0, x1, y1 int) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	sy := 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	for {
		t.recordEraseTile(state, x0, y0)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// FillTool handles flood-filling tiles on the active layer.
type FillTool struct{}

// NewFillTool creates a new fill tool.
func NewFillTool() *FillTool {
	return &FillTool{}
}

// OnMouseDown performs a flood fill starting at the clicked position.
func (t *FillTool) OnMouseDown(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	if state.MapData == nil {
		return
	}

	// For Tiles layer, require a selected tile
	if state.CurrentLayer != "Collision" && state.SelectedTile < 0 {
		return
	}

	// Determine the new tile ID
	var newID int
	if state.CurrentLayer == "Collision" {
		// Collision layer: use 1 for solid, 0 for empty
		if state.SelectedCollision {
			newID = 1
		} else {
			newID = 0
		}
	} else {
		// Tiles layer: use Tiled's 1-based tile IDs
		newID = state.SelectedTile + 1 // Convert 0-based to 1-based
	}

	// Create the fill action (it records all changes internally)
	action := NewFillTilesAction(state, state.CurrentLayer, tileX, tileY, newID)

	// Only record if there were actual changes
	if len(action.Changes) > 0 {
		state.History.Do(action, state)
	}
}

// OnMouseMove does nothing for fill tool.
func (t *FillTool) OnMouseMove(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	// Fill tool doesn't support drag
}

// OnMouseUp does nothing for fill tool.
func (t *FillTool) OnMouseUp(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	// Fill tool doesn't support drag
}

// SelectTool handles selecting, moving, and resizing objects.
// Supports multi-selection with Shift key.
type SelectTool struct {
	selection  *SelectionManager
	handleSize float64 // Size of resize handles in screen pixels
	snapToGrid bool
	// Track original position for undo
	originalX, originalY float64
	originalW, originalH float64
	dragStarted          bool
}

// NewSelectTool creates a new select tool.
func NewSelectTool() *SelectTool {
	return &SelectTool{
		selection:  NewSelectionManager(),
		handleSize: 8.0, // 8x8 pixel handles
		snapToGrid: true,
	}
}

// Selection returns the selection manager for external access.
func (t *SelectTool) Selection() *SelectionManager {
	return t.selection
}

// SetSnapToGrid enables or disables grid snapping.
func (t *SelectTool) SetSnapToGrid(snap bool) {
	t.snapToGrid = snap
}

// SnapToGrid returns whether grid snapping is enabled.
func (t *SelectTool) SnapToGrid() bool {
	return t.snapToGrid
}

// OnMouseDown handles object selection and starts drag operations.
// Supports Shift for multi-selection.
func (t *SelectTool) OnMouseDown(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	// Check if shift is held for multi-selection
	shiftHeld := ebiten.IsKeyPressed(ebiten.KeyShift)

	// Check if we clicked on a resize handle of the primary selected object
	if t.selection.HasSelection() {
		selectedObj := t.selection.GetSelectedObject(state.Objects)
		if selectedObj != nil {
			handle := t.selection.GetHandleAtPosition(worldX, worldY, selectedObj, t.handleSize)
			if handle != HandleNone {
				// Start resize operation (only for primary object)
				t.selection.BeginResize(worldX, worldY, selectedObj, handle)
				t.originalX = selectedObj.X
				t.originalY = selectedObj.Y
				t.originalW = selectedObj.W
				t.originalH = selectedObj.H
				t.dragStarted = true
				return
			}

			// Check if we clicked on the object body (for moving)
			if t.selection.IsOnObjectBody(worldX, worldY, selectedObj, t.handleSize) {
				// Start move operation for all selected objects
				t.selection.BeginMove(worldX, worldY, state.Objects)
				t.originalX = selectedObj.X
				t.originalY = selectedObj.Y
				t.originalW = selectedObj.W
				t.originalH = selectedObj.H
				t.dragStarted = true
				return
			}
		}
	}

	// Check if we clicked on any object
	hitIndex := t.selection.HitTest(worldX, worldY, state.Objects)
	if hitIndex >= 0 {
		if shiftHeld {
			// Add/remove from selection (toggle behavior)
			t.selection.AddToSelection(hitIndex)
			state.SelectObject(hitIndex) // Update state's primary selection for backward compatibility
		} else {
			// Single selection - replace current selection
			t.selection.Select(hitIndex)
			state.SelectObject(hitIndex)
		}

		// Start move operation immediately for all selected objects
		t.selection.BeginMove(worldX, worldY, state.Objects)
		selectedObj := t.selection.GetSelectedObject(state.Objects)
		if selectedObj != nil {
			t.originalX = selectedObj.X
			t.originalY = selectedObj.Y
			t.originalW = selectedObj.W
			t.originalH = selectedObj.H
			t.dragStarted = true
		}
	} else {
		// Clicked on empty space - clear selection (unless shift is held)
		if !shiftHeld {
			t.selection.ClearSelection()
			state.ClearSelection()
		}
		t.dragStarted = false
	}
}

// OnMouseMove handles dragging objects.
func (t *SelectTool) OnMouseMove(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	if !t.selection.IsDragging() {
		return
	}

	// Get grid size for snapping
	gridSize := 16
	if state.MapData != nil {
		gridSize = state.MapData.TileWidth()
	}

	// Update the drag operation for all selected objects
	t.selection.UpdateDrag(worldX, worldY, state.Objects, t.snapToGrid, gridSize)
}

// OnMouseUp finalizes drag operations.
func (t *SelectTool) OnMouseUp(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	// Create action if a drag occurred
	if t.selection.IsDragging() && t.dragStarted {
		selectedObj := t.selection.GetSelectedObject(state.Objects)
		if selectedObj != nil && t.selection.SelectedIndex() >= 0 {
			// Check if this was a move operation with multiple objects
			if t.selection.SelectionCount() > 1 && t.selection.DragMode() == DragModeMove {
				// Create multi-object move action
				originalPositions := t.selection.GetOriginalPositions()
				action := NewMoveMultipleObjectsAction(state.Objects, originalPositions)
				state.History.Do(action, state)
			} else {
				// Single object operation
				// Check if position or size actually changed
				if selectedObj.X != t.originalX || selectedObj.Y != t.originalY ||
					selectedObj.W != t.originalW || selectedObj.H != t.originalH {
					// Determine which action to create based on what changed
					if selectedObj.W != t.originalW || selectedObj.H != t.originalH {
						// Size changed - could be resize or move+resize
						if selectedObj.X != t.originalX || selectedObj.Y != t.originalY {
							// Both position and size changed
							action := NewMoveAndResizeObjectAction(
								t.selection.SelectedIndex(),
								t.originalX, t.originalY,
								selectedObj.X, selectedObj.Y,
								t.originalW, t.originalH,
								selectedObj.W, selectedObj.H,
							)
							state.History.Do(action, state)
						} else {
							// Only size changed
							action := NewResizeObjectAction(
								t.selection.SelectedIndex(),
								t.originalW, t.originalH,
								selectedObj.W, selectedObj.H,
							)
							state.History.Do(action, state)
						}
					} else {
						// Only position changed
						action := NewMoveObjectAction(
							t.selection.SelectedIndex(),
							t.originalX, t.originalY,
							selectedObj.X, selectedObj.Y,
						)
						state.History.Do(action, state)
					}
				}
			}
		}
	}

	t.selection.EndDrag()
	t.dragStarted = false
}

// GetHoverHandle returns the handle at the given world position for cursor display.
func (t *SelectTool) GetHoverHandle(state *EditorState, worldX, worldY float64) HandlePosition {
	if !t.selection.HasSelection() {
		return HandleNone
	}

	selectedObj := t.selection.GetSelectedObject(state.Objects)
	if selectedObj == nil {
		return HandleNone
	}

	return t.selection.GetHandleAtPosition(worldX, worldY, selectedObj, t.handleSize)
}

// PlaceObjectTool handles placing new objects in the level.
type PlaceObjectTool struct {
	objectPalette *ObjectPalette
	state         *EditorState // Reference to editor state for tool switching
}

// NewPlaceObjectTool creates a new place object tool.
func NewPlaceObjectTool(palette *ObjectPalette) *PlaceObjectTool {
	return &PlaceObjectTool{
		objectPalette: palette,
	}
}

// SetState sets the editor state reference.
func (t *PlaceObjectTool) SetState(state *EditorState) {
	t.state = state
}

// OnMouseDown places a new object at the clicked position.
func (t *PlaceObjectTool) OnMouseDown(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	if t.objectPalette == nil {
		return
	}

	// Get the selected object type from the palette
	objType := t.objectPalette.SelectedType()

	// If no object type is selected, do nothing
	if objType == "" {
		return
	}

	// Create a new object with default properties
	obj := CreateDefaultObject(objType, worldX, worldY)

	// Generate a unique ID for the object
	obj.ID = t.generateObjectID(state)

	// Get the index where the object will be added
	index := len(state.Objects)

	// Create and execute the action
	action := NewAddObjectAction(obj, index)
	state.History.Do(action, state)

	log.Printf("Placed %s at (%.0f, %.0f)", objType, worldX, worldY)

	// Deselect the object type in the palette
	t.objectPalette.ClearSelection()

	// Switch back to Select tool
	state.SetTool(ToolSelect)
}

// OnMouseMove does nothing for place object tool.
func (t *PlaceObjectTool) OnMouseMove(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	// Place object tool doesn't support drag
}

// OnMouseUp does nothing for place object tool.
func (t *PlaceObjectTool) OnMouseUp(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	// Place object tool doesn't support drag
}

// generateObjectID generates a unique ID for a new object.
func (t *PlaceObjectTool) generateObjectID(state *EditorState) int {
	maxID := 0
	for _, obj := range state.Objects {
		if obj.ID > maxID {
			maxID = obj.ID
		}
	}
	return maxID + 1
}

// ToolManager manages the available tools and dispatches input to the active tool.
type ToolManager struct {
	paintTool       *PaintTool
	eraseTool       *EraseTool
	fillTool        *FillTool
	selectTool      *SelectTool
	placeObjectTool *PlaceObjectTool
}

// NewToolManager creates a new tool manager with all tools initialized.
func NewToolManager() *ToolManager {
	return &ToolManager{
		paintTool:  NewPaintTool(),
		eraseTool:  NewEraseTool(),
		fillTool:   NewFillTool(),
		selectTool: NewSelectTool(),
	}
}

// SetObjectPalette sets the object palette for the place object tool.
func (tm *ToolManager) SetObjectPalette(palette *ObjectPalette) {
	tm.placeObjectTool = NewPlaceObjectTool(palette)
}

// SetState sets the editor state for tools that need it.
func (tm *ToolManager) SetState(state *EditorState) {
	if tm.placeObjectTool != nil {
		tm.placeObjectTool.SetState(state)
	}
}

// SelectTool returns the select tool for external access.
func (tm *ToolManager) SelectTool() *SelectTool {
	return tm.selectTool
}

// GetTool returns the tool handler for the given tool type.
func (tm *ToolManager) GetTool(tool Tool) ToolHandler {
	switch tool {
	case ToolPaint:
		return tm.paintTool
	case ToolErase:
		return tm.eraseTool
	case ToolFill:
		return tm.fillTool
	case ToolSelect:
		return tm.selectTool
	case ToolPlaceObject:
		if tm.placeObjectTool != nil {
			return tm.placeObjectTool
		}
		return tm.selectTool
	default:
		return tm.selectTool
	}
}

// HandleMouseDown dispatches mouse down to the appropriate tool.
func (tm *ToolManager) HandleMouseDown(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	// Ensure state has reference to selection manager
	state.SetSelectionManager(tm.selectTool.Selection())

	tool := tm.GetTool(state.CurrentTool)
	tool.OnMouseDown(state, tileX, tileY, worldX, worldY)
}

// HandleMouseMove dispatches mouse move to the appropriate tool.
func (tm *ToolManager) HandleMouseMove(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	tool := tm.GetTool(state.CurrentTool)
	tool.OnMouseMove(state, tileX, tileY, worldX, worldY)
}

// HandleMouseUp dispatches mouse up to the appropriate tool.
func (tm *ToolManager) HandleMouseUp(state *EditorState, tileX, tileY int, worldX, worldY float64) {
	tool := tm.GetTool(state.CurrentTool)
	tool.OnMouseUp(state, tileX, tileY, worldX, worldY)
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Ensure TileLayer has SetTile method available from world package
var _ = (*world.TileLayer)(nil)
