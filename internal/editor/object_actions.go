// Package editor provides the level editor functionality.
package editor

import (
	"fmt"

	"github.com/torsten/GoP/internal/world"
)

// AddObjectAction represents adding a new object to the level.
type AddObjectAction struct {
	Object      world.ObjectData
	ObjectIndex int // Index where the object was added
}

// NewAddObjectAction creates a new add object action.
func NewAddObjectAction(obj world.ObjectData, index int) *AddObjectAction {
	return &AddObjectAction{
		Object:      obj,
		ObjectIndex: index,
	}
}

// Do adds the object to the state.
func (a *AddObjectAction) Do(state *EditorState) {
	// Ensure the object is at the correct index
	if a.ObjectIndex == len(state.Objects) {
		state.Objects = append(state.Objects, a.Object)
	} else if a.ObjectIndex < len(state.Objects) {
		state.Objects = append(state.Objects[:a.ObjectIndex+1], state.Objects[a.ObjectIndex:]...)
		state.Objects[a.ObjectIndex] = a.Object
	} else {
		// Index out of range, just append
		state.Objects = append(state.Objects, a.Object)
	}
}

// Undo removes the object from the state.
func (a *AddObjectAction) Undo(state *EditorState) {
	if a.ObjectIndex >= 0 && a.ObjectIndex < len(state.Objects) {
		state.Objects = append(state.Objects[:a.ObjectIndex], state.Objects[a.ObjectIndex+1:]...)
	}
}

// Description returns a human-readable description.
func (a *AddObjectAction) Description() string {
	return fmt.Sprintf("Add %s object", a.Object.Type)
}

// DeleteObjectAction represents deleting an object from the level.
type DeleteObjectAction struct {
	Object      world.ObjectData
	ObjectIndex int // Index where the object was located
}

// NewDeleteObjectAction creates a new delete object action.
func NewDeleteObjectAction(obj world.ObjectData, index int) *DeleteObjectAction {
	return &DeleteObjectAction{
		Object:      obj,
		ObjectIndex: index,
	}
}

// Do removes the object from the state.
func (a *DeleteObjectAction) Do(state *EditorState) {
	if a.ObjectIndex >= 0 && a.ObjectIndex < len(state.Objects) {
		state.Objects = append(state.Objects[:a.ObjectIndex], state.Objects[a.ObjectIndex+1:]...)
	}
}

// Undo restores the object to the state.
func (a *DeleteObjectAction) Undo(state *EditorState) {
	// Re-insert the object at its original position
	if a.ObjectIndex <= len(state.Objects) {
		state.Objects = append(state.Objects[:a.ObjectIndex], append([]world.ObjectData{a.Object}, state.Objects[a.ObjectIndex:]...)...)
	}
}

// Description returns a human-readable description.
func (a *DeleteObjectAction) Description() string {
	return fmt.Sprintf("Delete %s object", a.Object.Type)
}

// MoveObjectAction represents moving an object to a new position.
type MoveObjectAction struct {
	ObjectIndex int
	OldX, OldY  float64
	NewX, NewY  float64
}

// NewMoveObjectAction creates a new move object action.
func NewMoveObjectAction(objectIndex int, oldX, oldY, newX, newY float64) *MoveObjectAction {
	return &MoveObjectAction{
		ObjectIndex: objectIndex,
		OldX:        oldX,
		OldY:        oldY,
		NewX:        newX,
		NewY:        newY,
	}
}

// Do moves the object to the new position.
func (a *MoveObjectAction) Do(state *EditorState) {
	if a.ObjectIndex >= 0 && a.ObjectIndex < len(state.Objects) {
		state.Objects[a.ObjectIndex].X = a.NewX
		state.Objects[a.ObjectIndex].Y = a.NewY
	}
}

// Undo moves the object back to the old position.
func (a *MoveObjectAction) Undo(state *EditorState) {
	if a.ObjectIndex >= 0 && a.ObjectIndex < len(state.Objects) {
		state.Objects[a.ObjectIndex].X = a.OldX
		state.Objects[a.ObjectIndex].Y = a.OldY
	}
}

// Description returns a human-readable description.
func (a *MoveObjectAction) Description() string {
	return fmt.Sprintf("Move object to (%.0f, %.0f)", a.NewX, a.NewY)
}

// ResizeObjectAction represents resizing an object.
type ResizeObjectAction struct {
	ObjectIndex int
	OldW, OldH  float64
	NewW, NewH  float64
}

// NewResizeObjectAction creates a new resize object action.
func NewResizeObjectAction(objectIndex int, oldW, oldH, newW, newH float64) *ResizeObjectAction {
	return &ResizeObjectAction{
		ObjectIndex: objectIndex,
		OldW:        oldW,
		OldH:        oldH,
		NewW:        newW,
		NewH:        newH,
	}
}

// Do resizes the object to the new dimensions.
func (a *ResizeObjectAction) Do(state *EditorState) {
	if a.ObjectIndex >= 0 && a.ObjectIndex < len(state.Objects) {
		state.Objects[a.ObjectIndex].W = a.NewW
		state.Objects[a.ObjectIndex].H = a.NewH
	}
}

// Undo resizes the object back to the old dimensions.
func (a *ResizeObjectAction) Undo(state *EditorState) {
	if a.ObjectIndex >= 0 && a.ObjectIndex < len(state.Objects) {
		state.Objects[a.ObjectIndex].W = a.OldW
		state.Objects[a.ObjectIndex].H = a.OldH
	}
}

// Description returns a human-readable description.
func (a *ResizeObjectAction) Description() string {
	return fmt.Sprintf("Resize object to %.0fx%.0f", a.NewW, a.NewH)
}

// SetPropertyAction represents changing a property on an object.
type SetPropertyAction struct {
	ObjectIndex  int
	PropertyName string
	OldValue     any
	NewValue     any
}

// NewSetPropertyAction creates a new set property action.
func NewSetPropertyAction(objectIndex int, propertyName string, oldValue, newValue any) *SetPropertyAction {
	return &SetPropertyAction{
		ObjectIndex:  objectIndex,
		PropertyName: propertyName,
		OldValue:     oldValue,
		NewValue:     newValue,
	}
}

// Do sets the property to the new value.
func (a *SetPropertyAction) Do(state *EditorState) {
	if a.ObjectIndex >= 0 && a.ObjectIndex < len(state.Objects) {
		if state.Objects[a.ObjectIndex].Props == nil {
			state.Objects[a.ObjectIndex].Props = make(map[string]any)
		}
		state.Objects[a.ObjectIndex].Props[a.PropertyName] = a.NewValue
	}
}

// Undo sets the property back to the old value.
func (a *SetPropertyAction) Undo(state *EditorState) {
	if a.ObjectIndex >= 0 && a.ObjectIndex < len(state.Objects) {
		if state.Objects[a.ObjectIndex].Props == nil {
			state.Objects[a.ObjectIndex].Props = make(map[string]any)
		}
		if a.OldValue == nil {
			// Property didn't exist before, remove it
			delete(state.Objects[a.ObjectIndex].Props, a.PropertyName)
		} else {
			state.Objects[a.ObjectIndex].Props[a.PropertyName] = a.OldValue
		}
	}
}

// Description returns a human-readable description.
func (a *SetPropertyAction) Description() string {
	return fmt.Sprintf("Set %s property", a.PropertyName)
}

// CompositeAction represents multiple actions that should be undone/redone together.
type CompositeAction struct {
	actions []Action
	desc    string
}

// NewCompositeAction creates a new composite action from multiple actions.
func NewCompositeAction(desc string, actions ...Action) *CompositeAction {
	return &CompositeAction{
		actions: actions,
		desc:    desc,
	}
}

// Do applies all actions in order.
func (a *CompositeAction) Do(state *EditorState) {
	for _, action := range a.actions {
		action.Do(state)
	}
}

// Undo reverses all actions in reverse order.
func (a *CompositeAction) Undo(state *EditorState) {
	for i := len(a.actions) - 1; i >= 0; i-- {
		a.actions[i].Undo(state)
	}
}

// Description returns a human-readable description.
func (a *CompositeAction) Description() string {
	return a.desc
}

// MoveAndResizeObjectAction represents both moving and resizing an object in one operation.
type MoveAndResizeObjectAction struct {
	ObjectIndex int
	OldX, OldY  float64
	NewX, NewY  float64
	OldW, OldH  float64
	NewW, NewH  float64
}

// NewMoveAndResizeObjectAction creates a new move and resize object action.
func NewMoveAndResizeObjectAction(objectIndex int, oldX, oldY, newX, newY, oldW, oldH, newW, newH float64) *MoveAndResizeObjectAction {
	return &MoveAndResizeObjectAction{
		ObjectIndex: objectIndex,
		OldX:        oldX,
		OldY:        oldY,
		NewX:        newX,
		NewY:        newY,
		OldW:        oldW,
		OldH:        oldH,
		NewW:        newW,
		NewH:        newH,
	}
}

// Do applies both move and resize.
func (a *MoveAndResizeObjectAction) Do(state *EditorState) {
	if a.ObjectIndex >= 0 && a.ObjectIndex < len(state.Objects) {
		state.Objects[a.ObjectIndex].X = a.NewX
		state.Objects[a.ObjectIndex].Y = a.NewY
		state.Objects[a.ObjectIndex].W = a.NewW
		state.Objects[a.ObjectIndex].H = a.NewH
	}
}

// Undo reverses both move and resize.
func (a *MoveAndResizeObjectAction) Undo(state *EditorState) {
	if a.ObjectIndex >= 0 && a.ObjectIndex < len(state.Objects) {
		state.Objects[a.ObjectIndex].X = a.OldX
		state.Objects[a.ObjectIndex].Y = a.OldY
		state.Objects[a.ObjectIndex].W = a.OldW
		state.Objects[a.ObjectIndex].H = a.OldH
	}
}

// Description returns a human-readable description.
func (a *MoveAndResizeObjectAction) Description() string {
	return fmt.Sprintf("Move and resize object to (%.0f, %.0f) %.0fx%.0f", a.NewX, a.NewY, a.NewW, a.NewH)
}

// MoveMultipleObjectsAction represents moving multiple objects at once.
type MoveMultipleObjectsAction struct {
	// Maps object index to original position
	OriginalPositions map[int]ObjectPosition
	// Stores the final positions after move (computed during Do)
	finalPositions map[int]ObjectPosition
}

// NewMoveMultipleObjectsAction creates a new move multiple objects action.
// It captures the current positions from the objects slice.
func NewMoveMultipleObjectsAction(objects []world.ObjectData, originalPositions map[int]ObjectPosition) *MoveMultipleObjectsAction {
	// Capture final positions from current object state
	finalPositions := make(map[int]ObjectPosition)
	for idx := range originalPositions {
		if idx >= 0 && idx < len(objects) {
			finalPositions[idx] = ObjectPosition{
				X: objects[idx].X,
				Y: objects[idx].Y,
				W: objects[idx].W,
				H: objects[idx].H,
			}
		}
	}

	return &MoveMultipleObjectsAction{
		OriginalPositions: originalPositions,
		finalPositions:    finalPositions,
	}
}

// Do moves all objects to their final positions.
func (a *MoveMultipleObjectsAction) Do(state *EditorState) {
	for idx, pos := range a.finalPositions {
		if idx >= 0 && idx < len(state.Objects) {
			state.Objects[idx].X = pos.X
			state.Objects[idx].Y = pos.Y
		}
	}
}

// Undo moves all objects back to their original positions.
func (a *MoveMultipleObjectsAction) Undo(state *EditorState) {
	for idx, pos := range a.OriginalPositions {
		if idx >= 0 && idx < len(state.Objects) {
			state.Objects[idx].X = pos.X
			state.Objects[idx].Y = pos.Y
		}
	}
}

// Description returns a human-readable description.
func (a *MoveMultipleObjectsAction) Description() string {
	return fmt.Sprintf("Move %d objects", len(a.OriginalPositions))
}

// DeleteMultipleObjectsAction represents deleting multiple objects at once.
type DeleteMultipleObjectsAction struct {
	// Objects and their original indices (sorted by index descending for proper deletion)
	Objects []world.ObjectData
	Indices []int // Original indices, sorted descending
}

// NewDeleteMultipleObjectsAction creates a new delete multiple objects action.
func NewDeleteMultipleObjectsAction(objects []world.ObjectData, indices []int) *DeleteMultipleObjectsAction {
	// Sort indices descending for proper deletion order
	sortedIndices := make([]int, len(indices))
	copy(sortedIndices, indices)
	sortDesc(sortedIndices)

	// Collect objects with their data
	objs := make([]world.ObjectData, len(indices))
	for i, idx := range sortedIndices {
		if idx >= 0 && idx < len(objects) {
			objs[i] = objects[idx]
		}
	}

	return &DeleteMultipleObjectsAction{
		Objects: objs,
		Indices: sortedIndices,
	}
}

// sortDesc sorts a slice of ints in descending order.
func sortDesc(s []int) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] < s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// Do removes all objects from the state.
func (a *DeleteMultipleObjectsAction) Do(state *EditorState) {
	// Delete from highest index first to maintain correct indices
	for _, idx := range a.Indices {
		if idx >= 0 && idx < len(state.Objects) {
			state.Objects = append(state.Objects[:idx], state.Objects[idx+1:]...)
		}
	}
}

// Undo restores all objects to their original positions.
func (a *DeleteMultipleObjectsAction) Undo(state *EditorState) {
	// Restore in reverse order (lowest index first)
	for i := len(a.Indices) - 1; i >= 0; i-- {
		idx := a.Indices[i]
		obj := a.Objects[i]
		// Re-insert the object at its original position
		if idx <= len(state.Objects) {
			state.Objects = append(state.Objects[:idx], append([]world.ObjectData{obj}, state.Objects[idx:]...)...)
		}
	}
}

// Description returns a human-readable description.
func (a *DeleteMultipleObjectsAction) Description() string {
	return fmt.Sprintf("Delete %d objects", len(a.Objects))
}

// LinkSwitchToDoorAction represents linking a switch to a door.
type LinkSwitchToDoorAction struct {
	SwitchIndex int
	OldDoorID   string
	NewDoorID   string
}

// NewLinkSwitchToDoorAction creates a new link switch to door action.
func NewLinkSwitchToDoorAction(switchIndex int, oldDoorID, newDoorID string) *LinkSwitchToDoorAction {
	return &LinkSwitchToDoorAction{
		SwitchIndex: switchIndex,
		OldDoorID:   oldDoorID,
		NewDoorID:   newDoorID,
	}
}

// Do sets the switch's door_id to the new door ID.
func (a *LinkSwitchToDoorAction) Do(state *EditorState) {
	if a.SwitchIndex >= 0 && a.SwitchIndex < len(state.Objects) {
		if state.Objects[a.SwitchIndex].Props == nil {
			state.Objects[a.SwitchIndex].Props = make(map[string]any)
		}
		state.Objects[a.SwitchIndex].Props["door_id"] = a.NewDoorID
	}
}

// Undo restores the switch's door_id to the old value.
func (a *LinkSwitchToDoorAction) Undo(state *EditorState) {
	if a.SwitchIndex >= 0 && a.SwitchIndex < len(state.Objects) {
		if state.Objects[a.SwitchIndex].Props == nil {
			state.Objects[a.SwitchIndex].Props = make(map[string]any)
		}
		if a.OldDoorID == "" {
			delete(state.Objects[a.SwitchIndex].Props, "door_id")
		} else {
			state.Objects[a.SwitchIndex].Props["door_id"] = a.OldDoorID
		}
	}
}

// Description returns a human-readable description.
func (a *LinkSwitchToDoorAction) Description() string {
	return fmt.Sprintf("Link switch to door '%s'", a.NewDoorID)
}

// SetPlatformEndpointAction represents changing the endpoint of a moving platform.
type SetPlatformEndpointAction struct {
	ObjectIndex int
	OldEndX     float64
	OldEndY     float64
	NewEndX     float64
	NewEndY     float64
}

// NewSetPlatformEndpointAction creates a new set platform endpoint action.
func NewSetPlatformEndpointAction(objectIndex int, oldEndX, oldEndY, newEndX, newEndY float64) *SetPlatformEndpointAction {
	return &SetPlatformEndpointAction{
		ObjectIndex: objectIndex,
		OldEndX:     oldEndX,
		OldEndY:     oldEndY,
		NewEndX:     newEndX,
		NewEndY:     newEndY,
	}
}

// Do sets the platform endpoint to the new values.
func (a *SetPlatformEndpointAction) Do(state *EditorState) {
	if a.ObjectIndex >= 0 && a.ObjectIndex < len(state.Objects) {
		if state.Objects[a.ObjectIndex].Props == nil {
			state.Objects[a.ObjectIndex].Props = make(map[string]any)
		}
		state.Objects[a.ObjectIndex].Props["endX"] = a.NewEndX
		state.Objects[a.ObjectIndex].Props["endY"] = a.NewEndY
	}
}

// Undo restores the platform endpoint to the old values.
func (a *SetPlatformEndpointAction) Undo(state *EditorState) {
	if a.ObjectIndex >= 0 && a.ObjectIndex < len(state.Objects) {
		if state.Objects[a.ObjectIndex].Props == nil {
			state.Objects[a.ObjectIndex].Props = make(map[string]any)
		}
		state.Objects[a.ObjectIndex].Props["endX"] = a.OldEndX
		state.Objects[a.ObjectIndex].Props["endY"] = a.OldEndY
	}
}

// Description returns a human-readable description.
func (a *SetPlatformEndpointAction) Description() string {
	return fmt.Sprintf("Set platform endpoint to (%.0f, %.0f)", a.NewEndX, a.NewEndY)
}
