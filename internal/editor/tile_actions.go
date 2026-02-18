// Package editor provides the level editor functionality.
package editor

import (
	"fmt"
)

// PaintTileAction represents a single tile paint operation.
type PaintTileAction struct {
	LayerName string
	TileX     int
	TileY     int
	OldTileID int
	NewTileID int
}

// NewPaintTileAction creates a new paint tile action.
// It captures the current tile value before the change.
func NewPaintTileAction(state *EditorState, layerName string, tileX, tileY, newTileID int) *PaintTileAction {
	oldTileID := 0
	if state.MapData != nil {
		layer := state.MapData.Layer(layerName)
		if layer != nil {
			oldTileID = layer.TileAt(tileX, tileY)
		}
	}

	return &PaintTileAction{
		LayerName: layerName,
		TileX:     tileX,
		TileY:     tileY,
		OldTileID: oldTileID,
		NewTileID: newTileID,
	}
}

// Do applies the paint tile action.
func (a *PaintTileAction) Do(state *EditorState) {
	if state.MapData == nil {
		return
	}
	layer := state.MapData.Layer(a.LayerName)
	if layer == nil {
		return
	}
	layer.SetTile(a.TileX, a.TileY, a.NewTileID)
}

// Undo reverses the paint tile action.
func (a *PaintTileAction) Undo(state *EditorState) {
	if state.MapData == nil {
		return
	}
	layer := state.MapData.Layer(a.LayerName)
	if layer == nil {
		return
	}
	layer.SetTile(a.TileX, a.TileY, a.OldTileID)
}

// Description returns a human-readable description.
func (a *PaintTileAction) Description() string {
	return fmt.Sprintf("Paint tile (%d, %d) on %s layer", a.TileX, a.TileY, a.LayerName)
}

// EraseTileAction represents a single tile erase operation.
type EraseTileAction struct {
	LayerName string
	TileX     int
	TileY     int
	OldTileID int
}

// NewEraseTileAction creates a new erase tile action.
// It captures the current tile value before erasing.
func NewEraseTileAction(state *EditorState, layerName string, tileX, tileY int) *EraseTileAction {
	oldTileID := 0
	if state.MapData != nil {
		layer := state.MapData.Layer(layerName)
		if layer != nil {
			oldTileID = layer.TileAt(tileX, tileY)
		}
	}

	return &EraseTileAction{
		LayerName: layerName,
		TileX:     tileX,
		TileY:     tileY,
		OldTileID: oldTileID,
	}
}

// Do applies the erase tile action.
func (a *EraseTileAction) Do(state *EditorState) {
	if state.MapData == nil {
		return
	}
	layer := state.MapData.Layer(a.LayerName)
	if layer == nil {
		return
	}
	layer.SetTile(a.TileX, a.TileY, 0)
}

// Undo reverses the erase tile action.
func (a *EraseTileAction) Undo(state *EditorState) {
	if state.MapData == nil {
		return
	}
	layer := state.MapData.Layer(a.LayerName)
	if layer == nil {
		return
	}
	layer.SetTile(a.TileX, a.TileY, a.OldTileID)
}

// Description returns a human-readable description.
func (a *EraseTileAction) Description() string {
	return fmt.Sprintf("Erase tile (%d, %d) on %s layer", a.TileX, a.TileY, a.LayerName)
}

// TileChange represents a single tile modification for batch operations.
type TileChange struct {
	TileX     int
	TileY     int
	OldTileID int
	NewTileID int
}

// FillTilesAction represents a flood fill operation that may modify multiple tiles.
type FillTilesAction struct {
	LayerName string
	Changes   []TileChange
}

// NewFillTilesAction creates a new fill tiles action.
// It performs the fill and records all changes.
func NewFillTilesAction(state *EditorState, layerName string, startX, startY, newTileID int) *FillTilesAction {
	if state.MapData == nil {
		return &FillTilesAction{LayerName: layerName, Changes: nil}
	}

	layer := state.MapData.Layer(layerName)
	if layer == nil {
		return &FillTilesAction{LayerName: layerName, Changes: nil}
	}

	// Check bounds
	if startX < 0 || startX >= layer.Width() || startY < 0 || startY >= layer.Height() {
		return &FillTilesAction{LayerName: layerName, Changes: nil}
	}

	// Get the target tile ID to replace
	targetID := layer.TileAt(startX, startY)

	// Don't fill if the target is the same as the new value
	if targetID == newTileID {
		return &FillTilesAction{LayerName: layerName, Changes: nil}
	}

	// Collect all changes (we need to record old values before modifying)
	changes := make([]TileChange, 0)

	// Use a simple stack-based flood fill algorithm
	stack := []struct{ x, y int }{{startX, startY}}
	visited := make(map[[2]int]bool)

	for len(stack) > 0 {
		// Pop from stack
		pos := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		x, y := pos.x, pos.y

		// Check bounds
		if x < 0 || x >= layer.Width() || y < 0 || y >= layer.Height() {
			continue
		}

		// Check if already visited
		key := [2]int{x, y}
		if visited[key] {
			continue
		}
		visited[key] = true

		// Check if this tile matches the target
		if layer.TileAt(x, y) != targetID {
			continue
		}

		// Record the change
		changes = append(changes, TileChange{
			TileX:     x,
			TileY:     y,
			OldTileID: targetID,
			NewTileID: newTileID,
		})

		// Add neighbors to stack
		stack = append(stack,
			struct{ x, y int }{x + 1, y},
			struct{ x, y int }{x - 1, y},
			struct{ x, y int }{x, y + 1},
			struct{ x, y int }{x, y - 1},
		)
	}

	return &FillTilesAction{
		LayerName: layerName,
		Changes:   changes,
	}
}

// Do applies the fill tiles action.
func (a *FillTilesAction) Do(state *EditorState) {
	if state.MapData == nil {
		return
	}
	layer := state.MapData.Layer(a.LayerName)
	if layer == nil {
		return
	}
	for _, change := range a.Changes {
		layer.SetTile(change.TileX, change.TileY, change.NewTileID)
	}
}

// Undo reverses the fill tiles action.
func (a *FillTilesAction) Undo(state *EditorState) {
	if state.MapData == nil {
		return
	}
	layer := state.MapData.Layer(a.LayerName)
	if layer == nil {
		return
	}
	for _, change := range a.Changes {
		layer.SetTile(change.TileX, change.TileY, change.OldTileID)
	}
}

// Description returns a human-readable description.
func (a *FillTilesAction) Description() string {
	return fmt.Sprintf("Fill %d tiles on %s layer", len(a.Changes), a.LayerName)
}

// PaintTilesAction represents a batch paint operation (e.g., from dragging).
type PaintTilesAction struct {
	LayerName string
	Changes   []TileChange
}

// NewPaintTilesAction creates a new paint tiles action for batch operations.
func NewPaintTilesAction(layerName string, changes []TileChange) *PaintTilesAction {
	return &PaintTilesAction{
		LayerName: layerName,
		Changes:   changes,
	}
}

// Do applies the paint tiles action.
func (a *PaintTilesAction) Do(state *EditorState) {
	if state.MapData == nil {
		return
	}
	layer := state.MapData.Layer(a.LayerName)
	if layer == nil {
		return
	}
	for _, change := range a.Changes {
		layer.SetTile(change.TileX, change.TileY, change.NewTileID)
	}
}

// Undo reverses the paint tiles action.
func (a *PaintTilesAction) Undo(state *EditorState) {
	if state.MapData == nil {
		return
	}
	layer := state.MapData.Layer(a.LayerName)
	if layer == nil {
		return
	}
	for _, change := range a.Changes {
		layer.SetTile(change.TileX, change.TileY, change.OldTileID)
	}
}

// Description returns a human-readable description.
func (a *PaintTilesAction) Description() string {
	return fmt.Sprintf("Paint %d tiles on %s layer", len(a.Changes), a.LayerName)
}

// EraseTilesAction represents a batch erase operation (e.g., from dragging).
type EraseTilesAction struct {
	LayerName string
	Changes   []TileChange
}

// NewEraseTilesAction creates a new erase tiles action for batch operations.
func NewEraseTilesAction(layerName string, changes []TileChange) *EraseTilesAction {
	return &EraseTilesAction{
		LayerName: layerName,
		Changes:   changes,
	}
}

// Do applies the erase tiles action.
func (a *EraseTilesAction) Do(state *EditorState) {
	if state.MapData == nil {
		return
	}
	layer := state.MapData.Layer(a.LayerName)
	if layer == nil {
		return
	}
	for _, change := range a.Changes {
		layer.SetTile(change.TileX, change.TileY, 0)
	}
}

// Undo reverses the erase tiles action.
func (a *EraseTilesAction) Undo(state *EditorState) {
	if state.MapData == nil {
		return
	}
	layer := state.MapData.Layer(a.LayerName)
	if layer == nil {
		return
	}
	for _, change := range a.Changes {
		layer.SetTile(change.TileX, change.TileY, change.OldTileID)
	}
}

// Description returns a human-readable description.
func (a *EraseTilesAction) Description() string {
	return fmt.Sprintf("Erase %d tiles on %s layer", len(a.Changes), a.LayerName)
}
