# Editor Feature Implementation Plan

## Overview

This document provides a detailed implementation plan for four new features in the GoP level editor:

1. **Collision Layer Editing** - Enable direct editing of the collision layer
2. **Drag-to-Scroll** - Left-click drag to scroll the canvas when no tool is active
3. **Switch-to-Door Linking** - Visual mechanism to link switches to doors
4. **Platform Start/End Point Setting** - Interactive setting of platform movement paths

---

## Current Architecture Summary

### Editor Structure

The editor follows a modular architecture with clear separation of concerns:

```
internal/editor/
  app.go           - Main application, coordinates all components
  state.go         - Editor state management (tools, layers, selection)
  tools.go         - Tool implementations (Paint, Erase, Fill, Select, PlaceObject)
  canvas.go        - Canvas rendering and mouse interaction dispatch
  camera.go        - Camera pan/zoom controls
  schema.go        - Entity property schemas for UI generation
  properties.go    - Property panel rendering and editing
  selection.go     - Multi-selection and resize handle management
  object_actions.go - Undo/redo actions for object modifications
  validation.go    - Level validation logic
```

### Key Patterns

1. **Tool System** - Tools implement `ToolHandler` interface with `OnMouseDown`, `OnMouseMove`, `OnMouseUp` methods
2. **Action System** - All modifications use undoable actions via the History system
3. **Schema-Driven Properties** - Entity properties are defined in `schema.go` and auto-generated in UI
4. **Layer System** - Two tile layers: "Tiles" (visual) and "Collision" (solid tiles)

### Current Mouse Handling

From [`canvas.go:808-873`](internal/editor/canvas.go:808-873):

```go
func (c *Canvas) handleToolInput() {
    // Get cursor position
    mx, my := ebiten.CursorPosition()
    
    // Convert to world coordinates
    worldX, worldY := c.camera.ScreenToWorld(mx, my)
    
    // Convert to tile coordinates
    tileX := int(worldX) / tileW
    tileY := int(worldY) / tileH
    
    // Dispatch to current tool
    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
        c.tools.HandleMouseDown(c.state, tileX, tileY, worldX, worldY)
    }
    // ... mouse move and up handling
}
```

### Current Layer System

From [`state.go:124-135`](internal/editor/state.go:124-135):

```go
func (s *EditorState) CycleLayer() {
    layers := []string{"Tiles", "Collision"}
    for i, layer := range layers {
        if layer == s.CurrentLayer {
            next := (i + 1) % len(layers)
            s.CurrentLayer = layers[next]
            return
        }
    }
}
```

The editor already supports switching between layers, but the collision layer editing is not fully implemented.

---

## Feature 1: Collision Layer Editing

### Current State

The collision layer exists in the level format and is rendered as a semi-transparent red overlay. However, editing the collision layer has limitations:

- Paint tool uses `state.SelectedTile + 1` for tile IDs
- Collision layer should use binary values (0 = empty, 1 = solid)
- No visual distinction in the palette for collision editing mode

### Required Changes

#### 1.1 Modify Paint Tool for Collision Layer

**File:** [`internal/editor/tools.go:66-105`](internal/editor/tools.go:66-105)

The paint tool already has special handling for collision layer at lines 84-87:

```go
// For Tiles layer, use Tiled's 1-based tile IDs
tileID := state.SelectedTile + 1
if state.CurrentLayer == "Collision" {
    tileID = 1 // Collision layer uses 1 for solid
}
```

**Issue:** When painting on collision layer, the tile palette should show a simple "solid/empty" toggle rather than tile IDs.

**Solution:** Modify the tileset rendering when collision layer is active.

#### 1.2 Update Tileset Panel for Collision Mode

**File:** [`internal/editor/tileset.go`](internal/editor/tileset.go)

Add a method to render collision-specific palette:

```go
// DrawCollisionPalette renders a simple collision palette
// Shows: [0] Empty (transparent) | [1] Solid (red square)
func (t *Tileset) DrawCollisionPalette(screen *ebiten.Image, selectedTile int, x int) {
    // Draw "Empty" option
    // Draw "Solid" option with red fill
    // Highlight selected
}
```

#### 1.3 Update Canvas Preview for Collision

**File:** [`internal/editor/canvas.go:740-788`](internal/editor/canvas.go:740-788)

Modify `drawToolPreview()` to show collision preview when on collision layer:

```go
func (c *Canvas) drawToolPreview(screen *ebiten.Image, canvasWidth int) {
    // Add collision layer check
    if c.state.CurrentLayer == "Collision" {
        if c.state.SelectedTile == 0 {
            // Show "erase" preview (transparent with X pattern)
        } else {
            // Show "solid" preview (semi-transparent red)
        }
        return
    }
    // ... existing tile preview code
}
```

### Implementation Steps

1. **Add collision palette mode** to [`tileset.go`](internal/editor/tileset.go)
   - New method `DrawCollisionPalette()`
   - Show two options: Empty (checkerboard) and Solid (red)

2. **Update palette rendering** in [`app.go:131-159`](internal/editor/app.go:131-159)
   - Check current layer in `handlePaletteInput()`
   - Call collision palette when `CurrentLayer == "Collision"`

3. **Update tool preview** in [`canvas.go`](internal/editor/canvas.go)
   - Add collision-specific preview rendering

4. **Update status bar** in [`app.go:630-687`](internal/editor/app.go:630-687)
   - Show "Collision: Solid/Empty" instead of tile ID

### UI/UX Considerations

- **Visual Feedback:** Collision tiles should render with distinct red overlay
- **Palette Simplicity:** Only two options needed (solid/empty)
- **Layer Indicator:** Status bar should clearly show "Collision Layer" mode
- **Shortcut:** Press 'C' to toggle collision overlay visibility (already implemented)

### Files to Modify

| File | Changes |
|------|---------|
| `internal/editor/tileset.go` | Add `DrawCollisionPalette()` method |
| `internal/editor/app.go` | Update palette handling for collision mode |
| `internal/editor/canvas.go` | Add collision preview rendering |
| `internal/editor/state.go` | Add `SelectedCollisionTile` field (optional) |

### Potential Risks

- **Confusion between layers:** Users might paint visual tiles while collision layer is active
- **Mitigation:** Clear visual indicator in status bar and different palette appearance

---

## Feature 2: Drag-to-Scroll

### Current Mouse Handling Approach

From [`internal/editor/camera.go:45-77`](internal/editor/camera.go:45-77):

```go
func (c *Camera) Update() {
    // Handle middle mouse button panning
    if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
        // ... pan logic
    }
}
```

Currently, panning is done via:
- Middle mouse button drag
- Arrow keys

### How to Detect "No Tool Active" State

The editor always has a tool selected. For drag-to-scroll, we need to define when left-click should scroll:

**Option A:** Add a "Pan" tool (new tool type)
**Option B:** Use Space + Left-click drag (like many art programs)
**Option C:** Pan when clicking on empty space in Select tool mode

**Recommended:** Option B - Space + Left-click drag

This is a common pattern in graphics applications (Photoshop, Aseprite, Tiled) and doesn't require changing the tool system.

### Implementation Approach

#### 2.1 Add Pan State to Camera

**File:** [`internal/editor/camera.go`](internal/editor/camera.go)

```go
type Camera struct {
    // ... existing fields
    
    // Left-click pan state
    isLeftPanning   bool
    leftPanStartX   int
    leftPanStartY   int
    leftPanCameraX  float64
    leftPanCameraY  float64
    spaceWasPressed bool // Track space for pan mode
}
```

#### 2.2 Update Camera Pan Logic

**File:** [`internal/editor/camera.go:45-100`](internal/editor/camera.go:45-100)

```go
func (c *Camera) Update() {
    // Check for space key (pan mode)
    spacePressed := ebiten.IsKeyPressed(ebiten.KeySpace)
    
    // Handle left-click panning when space is held
    if spacePressed && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
        mx, my := ebiten.CursorPosition()
        if !c.isLeftPanning {
            c.isLeftPanning = true
            c.leftPanStartX = mx
            c.leftPanStartY = my
            c.leftPanCameraX = c.X
            c.leftPanCameraY = c.Y
        } else {
            dx := float64(c.leftPanStartX-mx) / c.Zoom
            dy := float64(c.leftPanStartY-my) / c.Zoom
            c.X = c.leftPanCameraX + dx
            c.Y = c.leftPanCameraY + dy
        }
    } else {
        c.isLeftPanning = false
    }
    
    // ... existing middle mouse and keyboard pan code
}
```

#### 2.3 Block Tool Input During Pan

**File:** [`internal/editor/canvas.go:808-873`](internal/editor/canvas.go:808-873)

```go
func (c *Canvas) handleToolInput() {
    // Block tool input when space is held (pan mode)
    if ebiten.IsKeyPressed(ebiten.KeySpace) {
        return
    }
    
    // ... existing tool input code
}
```

#### 2.4 Update Cursor Display

**File:** [`internal/editor/app.go`](internal/editor/app.go)

Add cursor feedback when in pan mode:

```go
func (a *App) Update() error {
    // ... existing code
    
    // Update cursor for pan mode
    if ebiten.IsKeyPressed(ebiten.KeySpace) {
        // Set cursor to hand/grab icon (if supported)
    }
}
```

### Implementation Steps

1. **Add pan state fields** to [`Camera`](internal/editor/camera.go) struct
2. **Implement left-click pan logic** in [`Camera.Update()`](internal/editor/camera.go:45)
3. **Block tool input** when space is held in [`canvas.handleToolInput()`](internal/editor/canvas.go:808)
4. **Add cursor feedback** (optional, requires platform-specific code)

### Files to Modify

| File | Changes |
|------|---------|
| `internal/editor/camera.go` | Add left-click pan state and logic |
| `internal/editor/canvas.go` | Block tool input during pan mode |
| `internal/editor/app.go` | Update cursor display (optional) |

### Potential Risks

- **Conflict with existing shortcuts:** Space might conflict with other features
- **Mitigation:** Check shortcut list; Space is not currently used

---

## Feature 3: Switch-to-Door Linking

### Current Property System

From [`internal/editor/schema.go:57-68`](internal/editor/schema.go:57-68):

```go
world.ObjectTypeSwitch: {
    Properties: []PropertySchema{
        {Name: "door_id", Type: "string", Required: false, Default: ""},
        {Name: "toggle", Type: "bool", Required: false, Default: true},
        {Name: "once", Type: "bool", Required: false, Default: false},
    },
},
```

Switches have a `door_id` property that references a door's `id` property.

### Current Linking Visualization

From [`internal/editor/canvas.go:488-551`](internal/editor/canvas.go:488-551):

The editor already draws switch-to-door links when either the switch or door is selected:

```go
func (c *Canvas) drawSwitchDoorLinks(...) {
    // ... find switch and door pairs
    isSelected := (selection != nil && (selection.IsSelected(switchIdx) || selection.IsSelected(doorIdx)))
    if !isSelected {
        continue // Only draw if selected
    }
    // Draw connection line
}
```

### How to Implement ID-Based Linking

The current system requires manually typing the door ID in the properties panel. We need a visual linking mechanism.

### Implementation Approach

#### 3.1 Add Link Tool Mode

**File:** [`internal/editor/tools.go`](internal/editor/tools.go)

Add a new tool or mode for linking:

```go
const (
    // ... existing tools
    ToolLink  // New tool for linking switches to doors
)

// LinkTool handles creating connections between switches and doors
type LinkTool struct {
    sourceObj  *world.ObjectData  // The switch being linked
    sourceIdx  int
    isLinking  bool
}

func (t *LinkTool) OnMouseDown(state *EditorState, tileX, tileY int, worldX, worldY float64) {
    // 1. Check if clicking on a switch
    // 2. If yes, start linking mode
    // 3. If already linking and clicking on a door, create the link
}
```

#### 3.2 Add Link Mode to Select Tool

**Alternative Approach:** Add link mode to existing Select tool with modifier key.

**File:** [`internal/editor/tools.go:287-463`](internal/editor/tools.go:287-463)

Add Ctrl+click behavior to Select tool:

```go
func (t *SelectTool) OnMouseDown(state *EditorState, tileX, tileY int, worldX, worldY float64) {
    // Check for Ctrl+click on switch (start linking)
    ctrlHeld := ebiten.IsKeyPressed(ebiten.KeyControl)
    
    if ctrlHeld && t.selection.HasSelection() {
        selectedObj := t.selection.GetSelectedObject(state.Objects)
        if selectedObj != nil && selectedObj.Type == world.ObjectTypeSwitch {
            // Start link mode
            t.linkMode = true
            t.linkSource = t.selection.SelectedIndex()
            return
        }
    }
    
    // If in link mode and clicking on a door
    if t.linkMode && ctrlHeld {
        hitIndex := t.selection.HitTest(worldX, worldY, state.Objects)
        if hitIndex >= 0 && state.Objects[hitIndex].Type == world.ObjectTypeDoor {
            // Create the link
            doorID := state.Objects[hitIndex].GetPropString("id", "")
            action := NewSetPropertyAction(t.linkSource, "door_id", 
                state.Objects[t.linkSource].GetPropString("door_id", ""), doorID)
            state.History.Do(action, state)
            t.linkMode = false
            return
        }
    }
    
    // ... existing selection logic
}
```

#### 3.3 Visual Feedback for Linking

**File:** [`internal/editor/canvas.go`](internal/editor/canvas.go)

Add visual feedback during link mode:

```go
func (c *Canvas) Draw(screen *ebiten.Image) {
    // ... existing drawing
    
    // Draw link preview if in link mode
    if c.isInLinkMode() {
        c.drawLinkPreview(screen)
    }
}

func (c *Canvas) drawLinkPreview(screen *ebiten.Image) {
    // Draw line from source switch to cursor position
    // Use dashed line style to indicate "linking in progress"
}
```

#### 3.4 Update Properties Panel

**File:** [`internal/editor/properties.go`](internal/editor/properties.go)

Add a "Pick Door" button for the door_id property:

```go
func (p *PropertiesPanel) drawCustomProperty(...) {
    // Special handling for door_id property
    if propSchema.Name == "door_id" && obj.Type == world.ObjectTypeSwitch {
        // Draw "Pick Door" button
        // On click, enter link mode
    }
}
```

### Implementation Steps

1. **Add link mode state** to SelectTool or create new LinkTool
2. **Implement Ctrl+click linking** in SelectTool
3. **Add visual feedback** for link mode in canvas
4. **Add "Pick Door" button** in properties panel
5. **Update validation** to show broken links

### UI for Selecting Target Door

**Option A:** Click on door in canvas (recommended)
- Visual: Line from switch to cursor while in link mode
- Interaction: Ctrl+click switch to start, Ctrl+click door to complete

**Option B:** Dropdown list of door IDs
- Simpler but less visual
- Good for levels with many doors

**Recommended:** Implement both - visual linking as primary, dropdown as fallback.

### Files to Modify

| File | Changes |
|------|---------|
| `internal/editor/tools.go` | Add link mode to SelectTool |
| `internal/editor/canvas.go` | Add link preview rendering |
| `internal/editor/properties.go` | Add "Pick Door" button |
| `internal/editor/validation.go` | Add broken link warnings |

### Potential Risks

- **Door without ID:** Doors need unique IDs for linking
- **Mitigation:** Auto-generate ID when door is created, validate before linking

---

## Feature 4: Platform Start/End Point Setting

### Current Platform Properties

From [`internal/editor/schema.go:41-55`](internal/editor/schema.go:41-55):

```go
world.ObjectTypePlatform: {
    Properties: []PropertySchema{
        {Name: "id", Type: "string", Required: false, Default: ""},
        {Name: "endX", Type: "float", Required: false, Default: 0.0, Min: 0, Max: 10000},
        {Name: "endY", Type: "float", Required: false, Default: 0.0, Min: 0, Max: 10000},
        {Name: "speed", Type: "float", Required: false, Default: 100.0, Min: 0, Max: 1000},
        {Name: "waitTime", Type: "float", Required: false, Default: 0.5, Min: 0, Max: 10},
        {Name: "pushPlayer", Type: "bool", Required: false, Default: false},
    },
},
```

The `endX` and `endY` properties define the movement offset from the platform's position.

### Current Visualization

From [`internal/editor/canvas.go:389-432`](internal/editor/canvas.go:389-432):

```go
func (c *Canvas) drawPlatformPaths(screen *ebiten.Image, ...) {
    for _, obj := range c.state.Objects {
        if obj.Type != world.ObjectTypePlatform {
            continue
        }
        endX := obj.GetPropFloat("endX", 0)
        endY := obj.GetPropFloat("endY", 0)
        // Draw dashed line from start to end
        // Draw endpoint marker
    }
}
```

The editor already draws platform paths, but there's no interactive way to set the endpoint.

### Visual Representation Approach

**Current:** Dashed line from platform position to endpoint
**Enhancement:** Add draggable endpoint handle

### Implementation Approach

#### 4.1 Add Endpoint Handle to Selection System

**File:** [`internal/editor/selection.go`](internal/editor/selection.go)

Add a new handle type for platform endpoints:

```go
const (
    // ... existing handles
    HandlePlatformEnd  // Special handle for platform endpoint
)

// Add to SelectionManager
type SelectionManager struct {
    // ... existing fields
    platformEndDragging bool
}

// Check if clicking on platform endpoint
func (sm *SelectionManager) GetPlatformEndHandle(worldX, worldY float64, obj *world.ObjectData, handleSize float64) bool {
    if obj.Type != world.ObjectTypePlatform {
        return false
    }
    endX := obj.GetPropFloat("endX", 0)
    endY := obj.GetPropFloat("endY", 0)
    // Check if click is near the endpoint marker
    endpointX := obj.X + endX
    endpointY := obj.Y + endY
    return sm.PointInRect(worldX, worldY, endpointX - handleSize/2, endpointY - handleSize/2, handleSize, handleSize)
}
```

#### 4.2 Update Select Tool for Endpoint Dragging

**File:** [`internal/editor/tools.go:323-389`](internal/editor/tools.go:323-389)

```go
func (t *SelectTool) OnMouseDown(state *EditorState, tileX, tileY int, worldX, worldY float64) {
    // Check for platform endpoint handle first
    if t.selection.HasSelection() {
        selectedObj := t.selection.GetSelectedObject(state.Objects)
        if selectedObj != nil && selectedObj.Type == world.ObjectTypePlatform {
            if t.selection.GetPlatformEndHandle(worldX, worldY, selectedObj, t.handleSize) {
                // Start endpoint drag
                t.selection.BeginEndpointDrag(worldX, worldY, selectedObj)
                t.dragStarted = true
                return
            }
        }
    }
    // ... existing selection logic
}
```

#### 4.3 Add Endpoint Drag Logic

**File:** [`internal/editor/selection.go`](internal/editor/selection.go)

```go
func (sm *SelectionManager) BeginEndpointDrag(worldX, worldY float64, obj *world.ObjectData) {
    sm.dragMode = DragModePlatformEnd
    sm.dragStartX = worldX
    sm.dragStartY = worldY
    sm.originalEndX = obj.GetPropFloat("endX", 0)
    sm.originalEndY = obj.GetPropFloat("endY", 0)
}

func (sm *SelectionManager) UpdateEndpointDrag(worldX, worldY float64, obj *world.ObjectData, snapToGrid bool, gridSize int) {
    dx := worldX - sm.dragStartX
    dy := worldY - sm.dragStartY
    
    newEndX := sm.originalEndX + dx
    newEndY := sm.originalEndY + dy
    
    // Optional: Snap to grid
    if snapToGrid && gridSize > 0 {
        newEndX = float64(int(newEndX/float64(gridSize))) * float64(gridSize)
        newEndY = float64(int(newEndY/float64(gridSize))) * float64(gridSize)
    }
    
    // Update the property
    if obj.Props == nil {
        obj.Props = make(map[string]any)
    }
    obj.Props["endX"] = newEndX
    obj.Props["endY"] = newEndY
}
```

#### 4.4 Add Endpoint Drag Action for Undo

**File:** [`internal/editor/object_actions.go`](internal/editor/object_actions.go)

```go
// SetPlatformEndpointAction represents changing a platform's endpoint
type SetPlatformEndpointAction struct {
    ObjectIndex int
    OldEndX, OldEndY float64
    NewEndX, NewEndY float64
}

func NewSetPlatformEndpointAction(objectIndex int, oldEndX, oldEndY, newEndX, newEndY float64) *SetPlatformEndpointAction {
    return &SetPlatformEndpointAction{
        ObjectIndex: objectIndex,
        OldEndX:     oldEndX,
        OldEndY:     oldEndY,
        NewEndX:     newEndX,
        NewEndY:     newEndY,
    }
}

func (a *SetPlatformEndpointAction) Do(state *EditorState) {
    if a.ObjectIndex >= 0 && a.ObjectIndex < len(state.Objects) {
        obj := &state.Objects[a.ObjectIndex]
        if obj.Props == nil {
            obj.Props = make(map[string]any)
        }
        obj.Props["endX"] = a.NewEndX
        obj.Props["endY"] = a.NewEndY
    }
}

func (a *SetPlatformEndpointAction) Undo(state *EditorState) {
    if a.ObjectIndex >= 0 && a.ObjectIndex < len(state.Objects) {
        obj := &state.Objects[a.ObjectIndex]
        if obj.Props == nil {
            obj.Props = make(map[string]any)
        }
        obj.Props["endX"] = a.OldEndX
        obj.Props["endY"] = a.OldEndY
    }
}

func (a *SetPlatformEndpointAction) Description() string {
    return fmt.Sprintf("Set platform endpoint to (%.0f, %.0f)", a.NewEndX, a.NewEndY)
}
```

#### 4.5 Draw Endpoint Handle

**File:** [`internal/editor/canvas.go:580-619`](internal/editor/canvas.go:580-619)

Add endpoint handle drawing to `drawSelectionHandles()`:

```go
func (c *Canvas) drawSelectionHandles(screen *ebiten.Image, screenX, screenY, w, h, zoom float64) {
    // ... existing handle drawing
    
    // For platforms, also draw endpoint handle
    selectedObj := c.state.GetSelectedObject()
    if selectedObj != nil && selectedObj.Type == world.ObjectTypePlatform {
        endX := selectedObj.GetPropFloat("endX", 0)
        endY := selectedObj.GetPropFloat("endY", 0)
        if endX != 0 || endY != 0 {
            // Draw endpoint handle at (screenX + endX*zoom, screenY + endY*zoom)
            endpointScreenX := screenX + endX*zoom
            endpointScreenY := screenY + endY*zoom
            // Draw with different color (e.g., cyan) to distinguish from resize handles
        }
    }
}
```

### Interaction Design for Setting Points

1. **Select platform** - Click on platform to select it
2. **Endpoint handle appears** - Cyan handle at current endpoint (if set)
3. **Drag endpoint** - Click and drag the endpoint handle to reposition
4. **Visual feedback** - Dashed line updates in real-time during drag
5. **Snap to grid** - Optional grid snapping (same as object movement)

**Alternative:** Double-click on platform to enter "set endpoint" mode, then click to place endpoint.

### Implementation Steps

1. **Add endpoint handle detection** to SelectionManager
2. **Add endpoint drag mode** to DragMode enum
3. **Update SelectTool** to handle endpoint dragging
4. **Create SetPlatformEndpointAction** for undo support
5. **Draw endpoint handle** in canvas
6. **Update cursor** when hovering over endpoint

### Files to Modify

| File | Changes |
|------|---------|
| `internal/editor/selection.go` | Add endpoint handle detection and drag logic |
| `internal/editor/tools.go` | Handle endpoint dragging in SelectTool |
| `internal/editor/object_actions.go` | Add SetPlatformEndpointAction |
| `internal/editor/canvas.go` | Draw endpoint handle |

### Potential Risks

- **Endpoint at origin:** User might set endpoint to (0,0), making platform stationary
- **Mitigation:** Show warning or allow setting endpoint relative to platform center

---

## Summary

### Implementation Priority

| Feature | Complexity | Dependencies | Priority |
|---------|------------|--------------|----------|
| Collision Layer Editing | Low | None | High |
| Drag-to-Scroll | Low | None | High |
| Switch-to-Door Linking | Medium | Validation system | Medium |
| Platform Endpoints | Medium | Selection system | Medium |

### Recommended Implementation Order

1. **Drag-to-Scroll** - Simplest, improves workflow immediately
2. **Collision Layer Editing** - Completes existing layer system
3. **Platform Endpoints** - Builds on existing selection system
4. **Switch-to-Door Linking** - Most complex, requires new interaction mode

### Architecture Considerations

All features follow existing patterns:
- Use the Action system for undo/redo support
- Integrate with existing tool system
- Maintain separation between state, rendering, and input handling
- Follow the schema-driven property approach

### Testing Considerations

For each feature, test:
- Undo/redo functionality
- Keyboard shortcuts don't conflict
- Visual feedback is clear
- Edge cases (empty levels, many objects, etc.)