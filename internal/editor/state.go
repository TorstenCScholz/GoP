// Package editor provides the level editor functionality.
package editor

import (
	"github.com/torsten/GoP/internal/world"
)

// Tool represents the currently selected editor tool.
type Tool int

const (
	// ToolSelect allows selecting and moving objects.
	ToolSelect Tool = iota
	// ToolPaint allows painting tiles on the active layer.
	ToolPaint
	// ToolErase allows erasing tiles from the active layer.
	ToolErase
	// ToolFill allows flood-filling tiles on the active layer.
	ToolFill
	// ToolPlaceObject allows placing new objects.
	ToolPlaceObject
)

// StatusMessage represents a status message to display to the user.
type StatusMessage struct {
	Text    string // The message text
	IsError bool   // True if this is an error message
	Timer   int    // Frames remaining to display the message
}

// EditorState holds all the state for the level editor.
type EditorState struct {
	// Level data
	FilePath string             // Path to the current level file
	MapData  *world.MapData     // Parsed map data
	Objects  []world.ObjectData // Objects from the level
	Tileset  *world.Tileset     // Loaded tileset

	// UI state
	CurrentTool       Tool            // Currently selected tool
	CurrentLayer      string          // Name of the active layer ("Tiles" or "Collision")
	LayerVisible      map[string]bool // Visibility state for each layer
	SelectedTile      int             // Tile ID for painting (-1 if none)
	SelectedCollision bool            // Collision value for painting (true = solid, false = empty)
	SelectedObject    int             // Object index for selection (-1 if none)
	SpacePressed      bool            // True when Space key is held (for drag-to-scroll)

	// View state
	CameraX float64 // Camera X position in world coordinates
	CameraY float64 // Camera Y position in world coordinates
	Zoom    float64 // Camera zoom level (1.0 = 100%)

	// Undo/Redo
	History *History // Undo/redo history manager

	// Modification tracking
	modified bool // True if there are unsaved changes

	// Selection manager reference (set by tool manager)
	selectionManager *SelectionManager

	// Status message for user feedback
	StatusMessage *StatusMessage

	// Link mode - when true, next click on a door links it to the selected switch
	LinkMode     bool
	LinkSourceID int // Index of the switch object being linked

	// Endpoint dragging for platforms
	IsDraggingEndpoint bool    // True when dragging a platform endpoint handle
	DraggingObjectIdx  int     // Index of the platform being edited
	DragStartEndpointX float64 // Original endX value when drag started
	DragStartEndpointY float64 // Original endY value when drag started
	DragStartWorldX    float64 // World X position where drag started
	DragStartWorldY    float64 // World Y position where drag started
}

// NewEditorState creates a new editor state with default values.
func NewEditorState() *EditorState {
	return &EditorState{
		FilePath:          "",
		MapData:           nil,
		Objects:           nil,
		Tileset:           nil,
		CurrentTool:       ToolSelect,
		CurrentLayer:      "Tiles",
		LayerVisible:      map[string]bool{"Tiles": true, "Collision": true},
		SelectedTile:      -1,
		SelectedCollision: true, // Default to solid
		SelectedObject:    -1,
		CameraX:           0,
		CameraY:           0,
		Zoom:              1.0,
		History:           NewHistory(),
	}
}

// SetTool changes the current tool.
func (s *EditorState) SetTool(tool Tool) {
	s.CurrentTool = tool
}

// SetLayer changes the current layer.
func (s *EditorState) SetLayer(layer string) {
	s.CurrentLayer = layer
}

// SetLayerVisible sets the visibility of a layer.
func (s *EditorState) SetLayerVisible(name string, visible bool) {
	if s.LayerVisible == nil {
		s.LayerVisible = make(map[string]bool)
	}
	s.LayerVisible[name] = visible
}

// IsLayerVisible returns whether a layer is visible.
// Returns true by default if the layer is not in the map.
func (s *EditorState) IsLayerVisible(name string) bool {
	if s.LayerVisible == nil {
		return true
	}
	visible, ok := s.LayerVisible[name]
	if !ok {
		return true // Default to visible
	}
	return visible
}

// ToggleLayerVisibility toggles the visibility of the current layer.
func (s *EditorState) ToggleLayerVisibility() {
	if s.LayerVisible == nil {
		s.LayerVisible = make(map[string]bool)
	}
	current := s.IsLayerVisible(s.CurrentLayer)
	s.LayerVisible[s.CurrentLayer] = !current
}

// CycleLayer cycles through available layers.
func (s *EditorState) CycleLayer() {
	layers := []string{"Tiles", "Collision"}
	for i, layer := range layers {
		if layer == s.CurrentLayer {
			next := (i + 1) % len(layers)
			s.CurrentLayer = layers[next]
			return
		}
	}
	// Default to Tiles if current layer not found
	s.CurrentLayer = "Tiles"
}

// SelectTile sets the selected tile for painting.
func (s *EditorState) SelectTile(tileID int) {
	s.SelectedTile = tileID
}

// SelectCollision sets the selected collision value for painting.
func (s *EditorState) SelectCollision(solid bool) {
	s.SelectedCollision = solid
}

// SelectObject sets the selected object index.
func (s *EditorState) SelectObject(index int) {
	s.SelectedObject = index
}

// ClearSelection clears any object selection.
func (s *EditorState) ClearSelection() {
	s.SelectedObject = -1
}

// GetSelectedObject returns the currently selected object, or nil if none is selected.
func (s *EditorState) GetSelectedObject() *world.ObjectData {
	if s.SelectedObject < 0 || s.SelectedObject >= len(s.Objects) {
		return nil
	}
	return &s.Objects[s.SelectedObject]
}

// HasSelection returns true if an object is currently selected.
func (s *EditorState) HasSelection() bool {
	return s.SelectedObject >= 0 && s.SelectedObject < len(s.Objects)
}

// DeleteSelectedObject removes the currently selected object from the Objects slice.
// Returns true if an object was deleted, false otherwise.
func (s *EditorState) DeleteSelectedObject() bool {
	if !s.HasSelection() {
		return false
	}

	// Remove the object at the selected index
	index := s.SelectedObject
	s.Objects = append(s.Objects[:index], s.Objects[index+1:]...)
	s.SelectedObject = -1
	s.SetModified(true)
	return true
}

// HasLevel returns true if a level is loaded.
func (s *EditorState) HasLevel() bool {
	return s.MapData != nil
}

// HasTileset returns true if a tileset is loaded.
func (s *EditorState) HasTileset() bool {
	return s.Tileset != nil
}

// SetModified sets the modification state for unsaved changes tracking.
func (s *EditorState) SetModified(modified bool) {
	s.modified = modified
}

// IsModified returns true if there are unsaved changes.
func (s *EditorState) IsModified() bool {
	return s.modified
}

// GetSelectionManager returns the selection manager for multi-select operations.
// This is a convenience method that returns the selection manager from the tool manager.
func (s *EditorState) GetSelectionManager() *SelectionManager {
	// The selection manager is accessed through the SelectTool
	// We need to create a temporary reference to access it
	// This will be properly initialized when tools are set up
	return s.selectionManager
}

// SetSelectionManager sets the selection manager reference.
func (s *EditorState) SetSelectionManager(sm *SelectionManager) {
	s.selectionManager = sm
}

// ShowStatusMessage displays a status message to the user.
func (s *EditorState) ShowStatusMessage(text string, isError bool) {
	s.StatusMessage = &StatusMessage{
		Text:    text,
		IsError: isError,
		Timer:   180, // Show for 3 seconds at 60 FPS
	}
}

// UpdateStatusMessage updates the status message timer.
func (s *EditorState) UpdateStatusMessage() {
	if s.StatusMessage != nil {
		s.StatusMessage.Timer--
		if s.StatusMessage.Timer <= 0 {
			s.StatusMessage = nil
		}
	}
}

// ClearStatusMessage clears the current status message.
func (s *EditorState) ClearStatusMessage() {
	s.StatusMessage = nil
}

// StartLinkMode begins link mode for connecting a switch to a door.
func (s *EditorState) StartLinkMode(switchIndex int) {
	s.LinkMode = true
	s.LinkSourceID = switchIndex
}

// EndLinkMode exits link mode.
func (s *EditorState) EndLinkMode() {
	s.LinkMode = false
	s.LinkSourceID = -1
}

// IsInLinkMode returns true if the editor is in link mode.
func (s *EditorState) IsInLinkMode() bool {
	return s.LinkMode
}

// GetLinkSource returns the index of the switch being linked.
func (s *EditorState) GetLinkSource() int {
	return s.LinkSourceID
}

// selectionManager is stored separately for clipboard access
// (will be set by the tool manager during initialization)
