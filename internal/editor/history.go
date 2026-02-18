// Package editor provides the level editor functionality.
package editor

// Action defines an interface for undoable/redoable operations.
type Action interface {
	// Do applies the action to the editor state.
	Do(state *EditorState)
	// Undo reverses the action on the editor state.
	Undo(state *EditorState)
	// Description returns a human-readable description for UI display.
	Description() string
}

// History manages the undo/redo stack for editor actions.
type History struct {
	actions []Action
	index   int // Current position in history (points to next action to redo)
}

// NewHistory creates a new history manager.
func NewHistory() *History {
	return &History{
		actions: make([]Action, 0),
		index:   0,
	}
}

// Do executes an action and adds it to the history.
// Any actions that were previously undone are discarded.
func (h *History) Do(action Action, state *EditorState) {
	// Truncate future history if we're not at the end
	h.actions = h.actions[:h.index]

	// Execute the action
	action.Do(state)

	// Add to history and advance index
	h.actions = append(h.actions, action)
	h.index++

	// Mark state as modified
	state.SetModified(true)
}

// Undo reverses the last action in history.
// Returns true if an action was undone, false if there's nothing to undo.
func (h *History) Undo(state *EditorState) bool {
	if h.index == 0 {
		return false
	}

	h.index--
	h.actions[h.index].Undo(state)
	state.SetModified(true)
	return true
}

// Redo re-applies the next action in history.
// Returns true if an action was redone, false if there's nothing to redo.
func (h *History) Redo(state *EditorState) bool {
	if h.index >= len(h.actions) {
		return false
	}

	h.actions[h.index].Do(state)
	h.index++
	state.SetModified(true)
	return true
}

// CanUndo returns true if there are actions that can be undone.
func (h *History) CanUndo() bool {
	return h.index > 0
}

// CanRedo returns true if there are actions that can be redone.
func (h *History) CanRedo() bool {
	return h.index < len(h.actions)
}

// Clear removes all actions from history.
func (h *History) Clear() {
	h.actions = make([]Action, 0)
	h.index = 0
}

// UndoDescription returns the description of the action that would be undone.
// Returns empty string if there's nothing to undo.
func (h *History) UndoDescription() string {
	if h.index == 0 {
		return ""
	}
	return h.actions[h.index-1].Description()
}

// RedoDescription returns the description of the action that would be redone.
// Returns empty string if there's nothing to redo.
func (h *History) RedoDescription() string {
	if h.index >= len(h.actions) {
		return ""
	}
	return h.actions[h.index].Description()
}

// Count returns the total number of actions in history.
func (h *History) Count() int {
	return len(h.actions)
}

// Index returns the current position in history.
func (h *History) Index() int {
	return h.index
}
