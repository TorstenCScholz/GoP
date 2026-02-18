// Package editor provides the level editor functionality.
package editor

import (
	"encoding/json"
	"log"

	"github.com/torsten/GoP/internal/world"
)

// Clipboard stores copied object data for paste operations.
type Clipboard struct {
	// Serialized object data
	data []byte
	// Number of objects in clipboard
	count int
}

// NewClipboard creates a new clipboard.
func NewClipboard() *Clipboard {
	return &Clipboard{}
}

// Copy copies the selected objects to the clipboard.
func (c *Clipboard) Copy(state *EditorState) bool {
	// Get selected indices from the selection manager
	selection := state.GetSelectionManager()
	if selection == nil || !selection.HasSelection() {
		return false
	}

	indices := selection.SelectedIndices()
	if len(indices) == 0 {
		return false
	}

	// Collect objects to copy
	objects := make([]world.ObjectData, 0, len(indices))
	for _, idx := range indices {
		if idx >= 0 && idx < len(state.Objects) {
			objects = append(objects, state.Objects[idx])
		}
	}

	// Serialize the objects
	data, err := json.Marshal(objects)
	if err != nil {
		log.Printf("Failed to copy objects: %v", err)
		return false
	}

	c.data = data
	c.count = len(objects)
	log.Printf("Copied %d objects to clipboard", c.count)
	return true
}

// Paste pastes objects from the clipboard at an offset position.
// Returns the indices of the newly pasted objects.
func (c *Clipboard) Paste(state *EditorState) []int {
	if len(c.data) == 0 {
		return nil
	}

	// Deserialize the objects
	var objects []world.ObjectData
	if err := json.Unmarshal(c.data, &objects); err != nil {
		log.Printf("Failed to paste objects: %v", err)
		return nil
	}

	// Offset for paste (16px as specified)
	const pasteOffset = 16.0

	// Generate new IDs for pasted objects
	maxID := 0
	for _, obj := range state.Objects {
		if obj.ID > maxID {
			maxID = obj.ID
		}
	}

	// Create actions for each pasted object
	actions := make([]Action, 0, len(objects))
	newIndices := make([]int, 0, len(objects))

	for i, obj := range objects {
		// Offset position
		obj.X += pasteOffset
		obj.Y += pasteOffset

		// Assign new ID
		maxID++
		obj.ID = maxID

		// Add to end of objects list
		index := len(state.Objects) + i
		newIndices = append(newIndices, index)

		actions = append(actions, NewAddObjectAction(obj, index))
	}

	// Execute as composite action
	if len(actions) > 0 {
		action := NewCompositeAction("Paste objects", actions...)
		state.History.Do(action, state)
		log.Printf("Pasted %d objects from clipboard", len(objects))
	}

	return newIndices
}

// Cut copies selected objects to clipboard and then deletes them.
func (c *Clipboard) Cut(state *EditorState) bool {
	selection := state.GetSelectionManager()
	if selection == nil || !selection.HasSelection() {
		return false
	}

	// Copy first
	if !c.Copy(state) {
		return false
	}

	// Then delete
	indices := selection.SelectedIndices()
	action := NewDeleteMultipleObjectsAction(state.Objects, indices)
	state.History.Do(action, state)

	// Clear selection after cut
	selection.ClearSelection()
	state.ClearSelection()

	log.Printf("Cut %d objects", len(indices))
	return true
}

// HasContent returns true if there is content in the clipboard.
func (c *Clipboard) HasContent() bool {
	return len(c.data) > 0
}

// Count returns the number of objects in the clipboard.
func (c *Clipboard) Count() int {
	return c.count
}

// Clear clears the clipboard.
func (c *Clipboard) Clear() {
	c.data = nil
	c.count = 0
}
