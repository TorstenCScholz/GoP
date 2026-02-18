// Package editor provides the level editor functionality.
package editor

import (
	"sort"

	"github.com/torsten/GoP/internal/world"
)

// HandlePosition represents the position of a resize handle.
type HandlePosition int

const (
	// HandleNone indicates no handle is being interacted with.
	HandleNone HandlePosition = iota
	// Corner handles - resize both dimensions
	HandleTopLeft
	HandleTopRight
	HandleBottomLeft
	HandleBottomRight
	// Edge handles - resize one dimension
	HandleTop
	HandleBottom
	HandleLeft
	HandleRight
)

// DragMode represents the current drag operation mode.
type DragMode int

const (
	// DragModeNone indicates no drag operation.
	DragModeNone DragMode = iota
	// DragModeMove indicates the object is being moved.
	DragModeMove
	// DragModeResize indicates the object is being resized.
	DragModeResize
)

// SelectionManager handles object selection and manipulation.
// Supports multi-selection for moving and deleting multiple objects.
type SelectionManager struct {
	// Selection state - now supports multiple indices
	selectedIndices map[int]bool // Set of selected object indices
	primaryIndex    int          // The primary selected object (for resize operations)

	// Drag state
	dragMode   DragMode
	dragStartX float64
	dragStartY float64
	dragHandle HandlePosition
	// Original positions for all selected objects (for undo)
	originalPositions map[int]ObjectPosition
	// Original dimensions for resize (primary object only)
	originalW float64
	originalH float64

	// Constraints
	minSize float64
}

// ObjectPosition stores the original position of an object for undo operations.
type ObjectPosition struct {
	X, Y float64
	W, H float64
}

// NewSelectionManager creates a new selection manager.
func NewSelectionManager() *SelectionManager {
	return &SelectionManager{
		selectedIndices:   make(map[int]bool),
		primaryIndex:      -1,
		dragMode:          DragModeNone,
		minSize:           16.0, // Minimum size constraint
		originalPositions: make(map[int]ObjectPosition),
	}
}

// SelectedIndex returns the primary selected object index, or -1 if none.
// For backward compatibility with single-selection code.
func (sm *SelectionManager) SelectedIndex() int {
	return sm.primaryIndex
}

// SelectedIndices returns all selected object indices as a sorted slice.
func (sm *SelectionManager) SelectedIndices() []int {
	indices := make([]int, 0, len(sm.selectedIndices))
	for idx := range sm.selectedIndices {
		indices = append(indices, idx)
	}
	sort.Ints(indices)
	return indices
}

// Select sets a single selected object index (clears multi-selection).
func (sm *SelectionManager) Select(index int) {
	sm.selectedIndices = make(map[int]bool)
	sm.selectedIndices[index] = true
	sm.primaryIndex = index
	sm.dragMode = DragModeNone
}

// AddToSelection adds an index to the current selection (for shift-click).
func (sm *SelectionManager) AddToSelection(index int) {
	if sm.selectedIndices == nil {
		sm.selectedIndices = make(map[int]bool)
	}

	// If already selected, deselect it (toggle behavior)
	if sm.selectedIndices[index] {
		delete(sm.selectedIndices, index)
		// Update primary index if needed
		if sm.primaryIndex == index {
			sm.primaryIndex = -1
			// Pick a new primary from remaining selections
			for idx := range sm.selectedIndices {
				sm.primaryIndex = idx
				break
			}
		}
	} else {
		sm.selectedIndices[index] = true
		// First selected becomes primary
		if sm.primaryIndex < 0 {
			sm.primaryIndex = index
		}
	}
	sm.dragMode = DragModeNone
}

// ClearSelection clears all selected objects.
func (sm *SelectionManager) ClearSelection() {
	sm.selectedIndices = make(map[int]bool)
	sm.primaryIndex = -1
	sm.dragMode = DragModeNone
}

// HasSelection returns true if at least one object is selected.
func (sm *SelectionManager) HasSelection() bool {
	return len(sm.selectedIndices) > 0
}

// SelectionCount returns the number of selected objects.
func (sm *SelectionManager) SelectionCount() int {
	return len(sm.selectedIndices)
}

// IsSelected returns true if the given index is selected.
func (sm *SelectionManager) IsSelected(index int) bool {
	return sm.selectedIndices[index]
}

// GetSelectedObject returns the primary selected object from the objects slice, or nil if none.
// For backward compatibility with single-selection code.
func (sm *SelectionManager) GetSelectedObject(objects []world.ObjectData) *world.ObjectData {
	if !sm.HasSelection() || sm.primaryIndex < 0 || sm.primaryIndex >= len(objects) {
		return nil
	}
	return &objects[sm.primaryIndex]
}

// GetSelectedObjects returns all selected objects from the objects slice.
func (sm *SelectionManager) GetSelectedObjects(objects []world.ObjectData) []*world.ObjectData {
	result := make([]*world.ObjectData, 0, len(sm.selectedIndices))
	for idx := range sm.selectedIndices {
		if idx >= 0 && idx < len(objects) {
			result = append(result, &objects[idx])
		}
	}
	return result
}

// HitTest checks if a point is inside any object and returns its index.
// Returns -1 if no object is hit.
func (sm *SelectionManager) HitTest(worldX, worldY float64, objects []world.ObjectData) int {
	// Check objects in reverse order (top-most first)
	for i := len(objects) - 1; i >= 0; i-- {
		obj := objects[i]
		if sm.PointInRect(worldX, worldY, obj.X, obj.Y, obj.W, obj.H) {
			return i
		}
	}
	return -1
}

// PointInRect checks if a point is inside a rectangle.
func (sm *SelectionManager) PointInRect(px, py, rx, ry, rw, rh float64) bool {
	return px >= rx && px < rx+rw && py >= ry && py < ry+rh
}

// GetHandleAtPosition checks if a point is on a resize handle of the primary selected object.
// Returns HandleNone if no handle is hit.
func (sm *SelectionManager) GetHandleAtPosition(worldX, worldY float64, obj *world.ObjectData, handleSize float64) HandlePosition {
	if obj == nil {
		return HandleNone
	}

	// Handle size in world coordinates (typically scaled by zoom)
	hs := handleSize / 2.0

	// Check corner handles first (they take priority)
	// Top-left
	if sm.PointInRect(worldX, worldY, obj.X-hs, obj.Y-hs, hs*2, hs*2) {
		return HandleTopLeft
	}
	// Top-right
	if sm.PointInRect(worldX, worldY, obj.X+obj.W-hs, obj.Y-hs, hs*2, hs*2) {
		return HandleTopRight
	}
	// Bottom-left
	if sm.PointInRect(worldX, worldY, obj.X-hs, obj.Y+obj.H-hs, hs*2, hs*2) {
		return HandleBottomLeft
	}
	// Bottom-right
	if sm.PointInRect(worldX, worldY, obj.X+obj.W-hs, obj.Y+obj.H-hs, hs*2, hs*2) {
		return HandleBottomRight
	}

	// Check edge handles
	// Top edge
	if sm.PointInRect(worldX, worldY, obj.X+hs, obj.Y-hs, obj.W-hs*2, hs*2) {
		return HandleTop
	}
	// Bottom edge
	if sm.PointInRect(worldX, worldY, obj.X+hs, obj.Y+obj.H-hs, obj.W-hs*2, hs*2) {
		return HandleBottom
	}
	// Left edge
	if sm.PointInRect(worldX, worldY, obj.X-hs, obj.Y+hs, hs*2, obj.H-hs*2) {
		return HandleLeft
	}
	// Right edge
	if sm.PointInRect(worldX, worldY, obj.X+obj.W-hs, obj.Y+hs, hs*2, obj.H-hs*2) {
		return HandleRight
	}

	return HandleNone
}

// IsOnObjectBody checks if a point is on the object body (not on handles).
// This is used to determine if we should start a move operation.
func (sm *SelectionManager) IsOnObjectBody(worldX, worldY float64, obj *world.ObjectData, handleSize float64) bool {
	if obj == nil {
		return false
	}

	// First check if we're on a handle
	if sm.GetHandleAtPosition(worldX, worldY, obj, handleSize) != HandleNone {
		return false
	}

	// Then check if we're on the object body
	return sm.PointInRect(worldX, worldY, obj.X, obj.Y, obj.W, obj.H)
}

// BeginMove starts a move operation for all selected objects.
func (sm *SelectionManager) BeginMove(worldX, worldY float64, objects []world.ObjectData) {
	sm.dragMode = DragModeMove
	sm.dragStartX = worldX
	sm.dragStartY = worldY

	// Store original positions for all selected objects
	sm.originalPositions = make(map[int]ObjectPosition)
	for idx := range sm.selectedIndices {
		if idx >= 0 && idx < len(objects) {
			obj := objects[idx]
			sm.originalPositions[idx] = ObjectPosition{X: obj.X, Y: obj.Y, W: obj.W, H: obj.H}
		}
	}
}

// BeginMoveSingle starts a move operation for a single object (backward compatibility).
func (sm *SelectionManager) BeginMoveSingle(worldX, worldY float64, obj *world.ObjectData) {
	if obj == nil {
		return
	}
	sm.dragMode = DragModeMove
	sm.dragStartX = worldX
	sm.dragStartY = worldY
	sm.originalPositions = make(map[int]ObjectPosition)
	sm.originalPositions[sm.primaryIndex] = ObjectPosition{X: obj.X, Y: obj.Y, W: obj.W, H: obj.H}
}

// BeginResize starts a resize operation (only for primary selected object).
func (sm *SelectionManager) BeginResize(worldX, worldY float64, obj *world.ObjectData, handle HandlePosition) {
	if obj == nil {
		return
	}
	sm.dragMode = DragModeResize
	sm.dragStartX = worldX
	sm.dragStartY = worldY
	sm.dragHandle = handle
	sm.originalPositions = make(map[int]ObjectPosition)
	sm.originalPositions[sm.primaryIndex] = ObjectPosition{X: obj.X, Y: obj.Y, W: obj.W, H: obj.H}
	sm.originalW = obj.W
	sm.originalH = obj.H
}

// UpdateDrag updates the object position or size during a drag operation.
// Returns the new position and size values.
func (sm *SelectionManager) UpdateDrag(worldX, worldY float64, objects []world.ObjectData, snapToGrid bool, gridSize int) {
	if sm.dragMode == DragModeNone {
		return
	}

	dx := worldX - sm.dragStartX
	dy := worldY - sm.dragStartY

	if sm.dragMode == DragModeMove {
		// Move all selected objects
		for idx := range sm.selectedIndices {
			if idx >= 0 && idx < len(objects) {
				orig, ok := sm.originalPositions[idx]
				if !ok {
					continue
				}
				newX := orig.X + dx
				newY := orig.Y + dy

				// Snap to grid if enabled
				if snapToGrid && gridSize > 0 {
					newX = float64(int(newX/float64(gridSize))) * float64(gridSize)
					newY = float64(int(newY/float64(gridSize))) * float64(gridSize)
				}

				objects[idx].X = newX
				objects[idx].Y = newY
			}
		}
	} else if sm.dragMode == DragModeResize {
		// Resize only the primary object
		if sm.primaryIndex >= 0 && sm.primaryIndex < len(objects) {
			sm.applyResize(&objects[sm.primaryIndex], dx, dy, snapToGrid, gridSize)
		}
	}
}

// UpdateDragSingle updates a single object during drag (backward compatibility).
func (sm *SelectionManager) UpdateDragSingle(worldX, worldY float64, obj *world.ObjectData, snapToGrid bool, gridSize int) {
	if obj == nil || sm.dragMode == DragModeNone {
		return
	}

	dx := worldX - sm.dragStartX
	dy := worldY - sm.dragStartY

	if sm.dragMode == DragModeMove {
		newX := sm.originalPositions[sm.primaryIndex].X + dx
		newY := sm.originalPositions[sm.primaryIndex].Y + dy

		// Snap to grid if enabled
		if snapToGrid && gridSize > 0 {
			newX = float64(int(newX/float64(gridSize))) * float64(gridSize)
			newY = float64(int(newY/float64(gridSize))) * float64(gridSize)
		}

		obj.X = newX
		obj.Y = newY
	} else if sm.dragMode == DragModeResize {
		sm.applyResize(obj, dx, dy, snapToGrid, gridSize)
	}
}

// applyResize applies the resize operation based on the active handle.
func (sm *SelectionManager) applyResize(obj *world.ObjectData, dx, dy float64, snapToGrid bool, gridSize int) {
	orig := sm.originalPositions[sm.primaryIndex]

	switch sm.dragHandle {
	case HandleTopLeft:
		// Move top-left corner
		newX := orig.X + dx
		newY := orig.Y + dy
		newW := orig.W - dx
		newH := orig.H - dy

		// Apply minimum size constraint
		if newW >= sm.minSize && newH >= sm.minSize {
			if snapToGrid && gridSize > 0 {
				newX = float64(int(newX/float64(gridSize))) * float64(gridSize)
				newY = float64(int(newY/float64(gridSize))) * float64(gridSize)
				newW = float64(int(newW/float64(gridSize))) * float64(gridSize)
				newH = float64(int(newH/float64(gridSize))) * float64(gridSize)
			}
			obj.X = newX
			obj.Y = newY
			obj.W = newW
			obj.H = newH
		}

	case HandleTopRight:
		// Move top-right corner
		newY := orig.Y + dy
		newW := orig.W + dx
		newH := orig.H - dy

		if newW >= sm.minSize && newH >= sm.minSize {
			if snapToGrid && gridSize > 0 {
				newY = float64(int(newY/float64(gridSize))) * float64(gridSize)
				newW = float64(int(newW/float64(gridSize))) * float64(gridSize)
				newH = float64(int(newH/float64(gridSize))) * float64(gridSize)
			}
			obj.Y = newY
			obj.W = newW
			obj.H = newH
		}

	case HandleBottomLeft:
		// Move bottom-left corner
		newX := orig.X + dx
		newW := orig.W - dx
		newH := orig.H + dy

		if newW >= sm.minSize && newH >= sm.minSize {
			if snapToGrid && gridSize > 0 {
				newX = float64(int(newX/float64(gridSize))) * float64(gridSize)
				newW = float64(int(newW/float64(gridSize))) * float64(gridSize)
				newH = float64(int(newH/float64(gridSize))) * float64(gridSize)
			}
			obj.X = newX
			obj.W = newW
			obj.H = newH
		}

	case HandleBottomRight:
		// Move bottom-right corner
		newW := orig.W + dx
		newH := orig.H + dy

		if newW >= sm.minSize && newH >= sm.minSize {
			if snapToGrid && gridSize > 0 {
				newW = float64(int(newW/float64(gridSize))) * float64(gridSize)
				newH = float64(int(newH/float64(gridSize))) * float64(gridSize)
			}
			obj.W = newW
			obj.H = newH
		}

	case HandleTop:
		// Move top edge (resize height only)
		newY := orig.Y + dy
		newH := orig.H - dy

		if newH >= sm.minSize {
			if snapToGrid && gridSize > 0 {
				newY = float64(int(newY/float64(gridSize))) * float64(gridSize)
				newH = float64(int(newH/float64(gridSize))) * float64(gridSize)
			}
			obj.Y = newY
			obj.H = newH
		}

	case HandleBottom:
		// Move bottom edge (resize height only)
		newH := orig.H + dy

		if newH >= sm.minSize {
			if snapToGrid && gridSize > 0 {
				newH = float64(int(newH/float64(gridSize))) * float64(gridSize)
			}
			obj.H = newH
		}

	case HandleLeft:
		// Move left edge (resize width only)
		newX := orig.X + dx
		newW := orig.W - dx

		if newW >= sm.minSize {
			if snapToGrid && gridSize > 0 {
				newX = float64(int(newX/float64(gridSize))) * float64(gridSize)
				newW = float64(int(newW/float64(gridSize))) * float64(gridSize)
			}
			obj.X = newX
			obj.W = newW
		}

	case HandleRight:
		// Move right edge (resize width only)
		newW := orig.W + dx

		if newW >= sm.minSize {
			if snapToGrid && gridSize > 0 {
				newW = float64(int(newW/float64(gridSize))) * float64(gridSize)
			}
			obj.W = newW
		}
	}
}

// EndDrag ends the current drag operation.
func (sm *SelectionManager) EndDrag() {
	sm.dragMode = DragModeNone
}

// IsDragging returns true if a drag operation is in progress.
func (sm *SelectionManager) IsDragging() bool {
	return sm.dragMode != DragModeNone
}

// DragMode returns the current drag mode.
func (sm *SelectionManager) DragMode() DragMode {
	return sm.dragMode
}

// GetOriginalPositions returns the original positions of all selected objects before drag.
func (sm *SelectionManager) GetOriginalPositions() map[int]ObjectPosition {
	return sm.originalPositions
}

// GetCursorForHandle returns the cursor type name for a handle position.
// Returns "default" for HandleNone.
func GetCursorForHandle(handle HandlePosition) string {
	switch handle {
	case HandleTopLeft, HandleBottomRight:
		return "nwse-resize"
	case HandleTopRight, HandleBottomLeft:
		return "nesw-resize"
	case HandleTop, HandleBottom:
		return "ns-resize"
	case HandleLeft, HandleRight:
		return "ew-resize"
	default:
		return "default"
	}
}
