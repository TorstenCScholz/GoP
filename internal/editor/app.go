package editor

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// App implements ebiten.Game interface for the level editor.
type App struct {
	state           *EditorState
	camera          *Camera
	tileset         *Tileset
	canvas          *Canvas
	objectPalette   *ObjectPalette
	propertiesPanel *PropertiesPanel
	screenWidth     int
	screenHeight    int
	validation      *ValidationResult   // Last validation result
	playtest        *PlaytestController // Playtest mode controller
	clipboard       *Clipboard          // Clipboard for copy/paste
	showHelp        bool                // Show keyboard shortcuts overlay
	minimap         *Minimap            // Minimap component
}

// NewApp creates a new editor application.
func NewApp() *App {
	// Create a new default level on startup
	state := NewLevel(DefaultLevelWidth, DefaultLevelHeight)

	// Load tileset
	tileset := NewTileset()

	// Create camera
	camera := NewCamera()

	// Create canvas
	canvas := NewCanvas(state, camera, tileset)

	// Create object palette
	objectPalette := NewObjectPalette()

	// Create properties panel
	propertiesPanel := NewPropertiesPanel(state)

	// Set up the tool manager with the object palette
	canvas.tools.SetObjectPalette(objectPalette)

	// Create the app
	app := &App{
		state:           state,
		camera:          camera,
		tileset:         tileset,
		canvas:          canvas,
		objectPalette:   objectPalette,
		propertiesPanel: propertiesPanel,
	}

	// Create playtest controller with reference to app
	app.playtest = NewPlaytestController(app)

	// Create clipboard
	app.clipboard = NewClipboard()

	// Create minimap
	app.minimap = NewMinimap()

	return app
}

// Update updates the editor state.
// This is called every tick (typically 60 times per second).
func (a *App) Update() error {
	// If playtest mode is active, delegate to playtest controller
	if a.playtest != nil && a.playtest.IsActive() {
		return a.playtest.Update()
	}

	// Handle playtest mode toggle (P key)
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		if !ebiten.IsKeyPressed(ebiten.KeyControl) {
			if err := a.playtest.StartPlaytest(); err != nil {
				log.Printf("Failed to start playtest: %v", err)
			}
			return nil
		}
	}

	// Handle keyboard shortcuts for file operations
	a.handleFileShortcuts()

	// Handle keyboard shortcuts for tools and layers
	a.handleToolShortcuts()

	// Handle validation shortcuts
	a.handleValidationShortcuts()

	// Update camera controls
	a.camera.Update()

	// Update editor state from camera
	a.state.CameraX = a.camera.X
	a.state.CameraY = a.camera.Y
	a.state.Zoom = a.camera.Zoom

	// Update canvas (handles grid/collision toggle, tool input, etc.)
	a.canvas.Update()

	// Handle tile selection from palette
	a.handlePaletteInput()

	// Handle object palette input
	a.handleObjectPaletteInput()

	// Handle properties panel input
	a.handlePropertiesInput()

	// Handle minimap click for navigation
	a.handleMinimapInput()

	// Update status message timer
	a.state.UpdateStatusMessage()

	return nil
}

// handlePaletteInput handles mouse input for tile selection from the palette.
func (a *App) handlePaletteInput() {
	if !a.tileset.IsLoaded() {
		return
	}

	// Check for mouse click in palette area
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		screenWidth := a.screenWidth
		if screenWidth == 0 {
			screenWidth = 1280 // Default width
		}
		if a.tileset.IsInPalette(mx, my, screenWidth) {
			// Calculate tile position within palette
			paletteX := screenWidth - PaletteWidth - ObjectPaletteWidth
			localX := mx - paletteX - PalettePadding
			localY := my - PalettePadding

			if localX >= 0 && localY >= 0 {
				tileID := a.tileset.TileAtPosition(mx-paletteX, my)
				if tileID >= 0 {
					a.state.SelectTile(tileID)
					a.state.SetTool(ToolPaint)
					log.Printf("Selected tile ID: %d", tileID)
				}
			}
		}
	}
}

// handleObjectPaletteInput handles mouse input for object type selection from the palette.
func (a *App) handleObjectPaletteInput() {
	mx, my := ebiten.CursorPosition()
	screenWidth := a.screenWidth
	if screenWidth == 0 {
		screenWidth = 1280 // Default width
	}

	// Update hover state
	a.objectPalette.HandleMouseMove(mx, my, screenWidth, 0)

	// Check for mouse click in object palette area
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if a.objectPalette.IsInPalette(mx, my, screenWidth, 0) {
			if a.objectPalette.HandleClick(mx, my, screenWidth, 0) {
				a.state.SetTool(ToolPlaceObject)
				log.Printf("Selected object type: %s", a.objectPalette.SelectedType())
			}
		}
	}
}

// handlePropertiesInput handles mouse input for the properties panel.
func (a *App) handlePropertiesInput() {
	mx, my := ebiten.CursorPosition()
	screenWidth := a.screenWidth
	screenHeight := a.screenHeight
	if screenWidth == 0 {
		screenWidth = 1280 // Default width
	}
	if screenHeight == 0 {
		screenHeight = 720 // Default height
	}

	// Calculate where properties panel starts (below object palette buttons)
	propertiesStartY := a.getPropertiesPanelStartY()

	// Update hover state
	a.propertiesPanel.HandleMouseMove(mx, my, screenWidth, propertiesStartY)

	// Handle property editing updates (text input, etc.)
	a.propertiesPanel.Update()

	// Check for mouse click in properties panel area
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if a.propertiesPanel.IsInPanel(mx, my, screenWidth, propertiesStartY) {
			a.propertiesPanel.HandleClick(mx, my, screenWidth, propertiesStartY)
		}
	}
}

// handleMinimapInput handles mouse input for the minimap.
func (a *App) handleMinimapInput() {
	if a.minimap == nil {
		return
	}

	mx, my := ebiten.CursorPosition()
	screenWidth := a.screenWidth
	screenHeight := a.screenHeight
	if screenWidth == 0 {
		screenWidth = 1280
	}
	if screenHeight == 0 {
		screenHeight = 720
	}

	canvasWidth := screenWidth - PaletteWidth - ObjectPaletteWidth

	// Check for click on minimap
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		a.minimap.HandleClick(mx, my, a.state, a.camera, canvasWidth, screenHeight)
	}
}

// getPropertiesPanelStartY calculates the Y position where the properties panel should start.
func (a *App) getPropertiesPanelStartY() int {
	// Properties panel starts at the bottom of the screen
	// Object palette is at the top, properties panel at the bottom
	screenHeight := a.screenHeight
	if screenHeight == 0 {
		screenHeight = 720
	}
	return screenHeight - PropertiesPanelHeight
}

// handleFileShortcuts processes keyboard shortcuts for file operations.
func (a *App) handleFileShortcuts() {
	// Ctrl+N: New level
	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyN) {
		a.newLevel()
	}

	// Ctrl+O: Open level
	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyO) {
		a.openLevel()
	}

	// Ctrl+S: Save level
	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyS) {
		// Check for Ctrl+Shift+S (Save As)
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			a.saveLevelAs()
		} else {
			a.saveLevel()
		}
	}
}

// handleToolShortcuts processes keyboard shortcuts for tool and layer selection.
func (a *App) handleToolShortcuts() {
	// Undo: Ctrl+Z
	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		if !ebiten.IsKeyPressed(ebiten.KeyShift) {
			if a.state.History.Undo(a.state) {
				log.Printf("Undo: %s", a.state.History.UndoDescription())
			}
			return
		}
	}

	// Redo: Ctrl+Y or Ctrl+Shift+Z
	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyY) {
		if a.state.History.Redo(a.state) {
			log.Printf("Redo: %s", a.state.History.RedoDescription())
		}
		return
	}
	if ebiten.IsKeyPressed(ebiten.KeyControl) && ebiten.IsKeyPressed(ebiten.KeyShift) && inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		if a.state.History.Redo(a.state) {
			log.Printf("Redo: %s", a.state.History.RedoDescription())
		}
		return
	}

	// Tool selection shortcuts
	// 1 or S - Select tool
	if inpututil.IsKeyJustPressed(ebiten.Key1) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		// Only if not pressing Ctrl (to avoid conflict with Ctrl+S)
		if !ebiten.IsKeyPressed(ebiten.KeyControl) {
			a.state.SetTool(ToolSelect)
			log.Println("Selected tool: Select")
		}
	}

	// 2 - Paint tool (removed P shortcut, now used for playtest)
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		if !ebiten.IsKeyPressed(ebiten.KeyControl) {
			a.state.SetTool(ToolPaint)
			log.Println("Selected tool: Paint")
		}
	}

	// 3 or E - Erase tool
	if inpututil.IsKeyJustPressed(ebiten.Key3) || inpututil.IsKeyJustPressed(ebiten.KeyE) {
		if !ebiten.IsKeyPressed(ebiten.KeyControl) {
			a.state.SetTool(ToolErase)
			log.Println("Selected tool: Erase")
		}
	}

	// 4 or F - Fill tool
	if inpututil.IsKeyJustPressed(ebiten.Key4) || inpututil.IsKeyJustPressed(ebiten.KeyF) {
		if !ebiten.IsKeyPressed(ebiten.KeyControl) {
			a.state.SetTool(ToolFill)
			log.Println("Selected tool: Fill")
		}
	}

	// 5 or O - Place Object tool
	if inpututil.IsKeyJustPressed(ebiten.Key5) || inpututil.IsKeyJustPressed(ebiten.KeyO) {
		if !ebiten.IsKeyPressed(ebiten.KeyControl) {
			a.state.SetTool(ToolPlaceObject)
			log.Println("Selected tool: Place Object")
		}
	}

	// Layer selection shortcuts
	// Tab - Cycle between layers
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		a.state.CycleLayer()
		log.Printf("Current layer: %s", a.state.CurrentLayer)
	}

	// H - Toggle current layer visibility
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		a.state.ToggleLayerVisibility()
		visible := a.state.IsLayerVisible(a.state.CurrentLayer)
		log.Printf("Layer %s visibility: %v", a.state.CurrentLayer, visible)
	}

	// Delete/Backspace - Delete selected object(s) or hovered tile in erase mode
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) || inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		selection := a.state.GetSelectionManager()
		if selection != nil && selection.HasSelection() {
			count := selection.SelectionCount()
			if count > 1 {
				// Multi-select delete
				indices := selection.SelectedIndices()
				action := NewDeleteMultipleObjectsAction(a.state.Objects, indices)
				a.state.History.Do(action, a.state)
				selection.ClearSelection()
				a.state.ClearSelection()
				log.Printf("Deleted %d objects", count)
			} else {
				// Single object delete
				obj := a.state.GetSelectedObject()
				objType := "unknown"
				if obj != nil {
					objType = string(obj.Type)
				}
				action := NewDeleteObjectAction(*obj, a.state.SelectedObject)
				a.state.History.Do(action, a.state)
				a.state.ClearSelection()
				selection.ClearSelection()
				log.Printf("Deleted selected object: %s", objType)
			}
		} else if a.state.CurrentTool == ToolErase {
			// In erase mode, delete the hovered tile
			tileX, tileY := a.canvas.HoveredTile()
			if tileX >= 0 && tileY >= 0 {
				// Create an erase action for the hovered tile
				action := NewEraseTileAction(a.state, a.state.CurrentLayer, tileX, tileY)
				a.state.History.Do(action, a.state)
				log.Printf("Erased tile at (%d, %d)", tileX, tileY)
			}
		}
	}

	// Escape - Clear selection and close help
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if a.showHelp {
			a.showHelp = false
		} else if a.state.HasSelection() {
			a.state.ClearSelection()
			selection := a.state.GetSelectionManager()
			if selection != nil {
				selection.ClearSelection()
			}
			log.Println("Cleared selection")
		}
	}

	// Copy: Ctrl+C
	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyC) {
		if a.clipboard.Copy(a.state) {
			log.Println("Copied selection to clipboard")
		}
	}

	// Paste: Ctrl+V
	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyV) {
		indices := a.clipboard.Paste(a.state)
		if len(indices) > 0 {
			// Select the newly pasted objects
			selection := a.state.GetSelectionManager()
			if selection != nil {
				selection.ClearSelection()
				for _, idx := range indices {
					selection.AddToSelection(idx)
				}
				if len(indices) > 0 {
					a.state.SelectObject(indices[0])
				}
			}
			log.Printf("Pasted %d objects from clipboard", len(indices))
		}
	}

	// Cut: Ctrl+X
	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyX) {
		if a.clipboard.Cut(a.state) {
			log.Println("Cut selection to clipboard")
		}
	}

	// Help: ? or F1
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) ||
		(ebiten.IsKeyPressed(ebiten.KeyShift) && inpututil.IsKeyJustPressed(ebiten.KeySlash)) {
		a.showHelp = !a.showHelp
	}
}

// handleValidationShortcuts processes keyboard shortcuts for validation.
func (a *App) handleValidationShortcuts() {
	// V - Run validation
	if inpututil.IsKeyJustPressed(ebiten.KeyV) {
		if !ebiten.IsKeyPressed(ebiten.KeyControl) {
			a.runValidation()
		}
	}
}

// runValidation validates the current level and logs the results.
func (a *App) runValidation() {
	a.validation = ValidateLevel(a.state)

	if !a.validation.HasIssues() {
		log.Println("Validation passed: No issues found")
		return
	}

	// Log all errors
	for _, err := range a.validation.Errors {
		log.Printf("ERROR: %s", FormatValidationError(err))
	}

	// Log all warnings
	for _, warn := range a.validation.Warnings {
		log.Printf("WARNING: %s", FormatValidationError(warn))
	}

	log.Printf("Validation complete: %d errors, %d warnings", a.validation.ErrorCount(), a.validation.WarningCount())
}

// newLevel creates a new empty level.
func (a *App) newLevel() {
	// TODO: Check for unsaved changes and prompt user
	a.state = NewLevel(DefaultLevelWidth, DefaultLevelHeight)
	a.camera.Reset()
	// Recreate canvas with new state
	a.canvas = NewCanvas(a.state, a.camera, a.tileset)
	a.canvas.tools.SetObjectPalette(a.objectPalette)
	// Update properties panel with new state
	a.propertiesPanel.SetState(a.state)
	log.Println("Created new level")
}

// openLevel opens an existing level file.
func (a *App) openLevel() {
	// TODO: Check for unsaved changes and prompt user

	// Use the current file path if set, otherwise use default
	path := a.state.FilePath
	if path == "" {
		path = DefaultLevelPath
	}

	state, err := OpenLevel(path)
	if err != nil {
		log.Printf("Failed to open level: %v", err)
		a.state.ShowStatusMessage(fmt.Sprintf("Failed to open: %v", err), true)
		return
	}

	a.state = state
	a.camera.Reset()
	// Recreate canvas with new state
	a.canvas = NewCanvas(a.state, a.camera, a.tileset)
	a.canvas.tools.SetObjectPalette(a.objectPalette)
	// Update properties panel with new state
	a.propertiesPanel.SetState(a.state)
	log.Printf("Opened level: %s", a.state.FilePath)
	a.state.ShowStatusMessage(fmt.Sprintf("Opened: %s", a.state.FilePath), false)
}

// saveLevel saves the current level.
func (a *App) saveLevel() {
	if !a.state.HasLevel() {
		log.Println("No level to save")
		a.state.ShowStatusMessage("No level to save", true)
		return
	}

	// Run validation before save
	a.validation = ValidateLevel(a.state)

	// Log validation issues
	if a.validation.HasIssues() {
		for _, err := range a.validation.Errors {
			log.Printf("ERROR: %s", FormatValidationError(err))
		}
		for _, warn := range a.validation.Warnings {
			log.Printf("WARNING: %s", FormatValidationError(warn))
		}

		// Still allow saving with warnings, but log the issues
		if a.validation.HasErrors() {
			log.Printf("Level has %d critical errors - consider fixing before playing", a.validation.ErrorCount())
		}
	}

	// If no file path, use save as
	if a.state.FilePath == "" {
		a.saveLevelAs()
		return
	}

	if err := SaveLevel(a.state); err != nil {
		log.Printf("Failed to save level: %v", err)
		a.state.ShowStatusMessage(fmt.Sprintf("Failed to save: %v", err), true)
		return
	}

	log.Printf("Saved level: %s", a.state.FilePath)
	a.state.ShowStatusMessage(fmt.Sprintf("Saved: %s", a.state.FilePath), false)
}

// saveLevelAs saves the current level to a new file.
func (a *App) saveLevelAs() {
	if !a.state.HasLevel() {
		log.Println("No level to save")
		a.state.ShowStatusMessage("No level to save", true)
		return
	}

	// Use hardcoded path for now (will add file dialogs later)
	// If there's already a path, use it; otherwise use default
	path := a.state.FilePath
	if path == "" {
		path = DefaultLevelPath
	}

	if err := SaveLevelAs(a.state, path); err != nil {
		log.Printf("Failed to save level: %v", err)
		a.state.ShowStatusMessage(fmt.Sprintf("Failed to save: %v", err), true)
		return
	}

	log.Printf("Saved level as: %s", a.state.FilePath)
	a.state.ShowStatusMessage(fmt.Sprintf("Saved: %s", a.state.FilePath), false)
}

// Draw renders the editor to the screen.
func (a *App) Draw(screen *ebiten.Image) {
	// If playtest mode is active, delegate to playtest controller
	if a.playtest != nil && a.playtest.IsActive() {
		a.playtest.Draw(screen)
		return
	}

	// Clear screen with a dark background
	screen.Fill(color.RGBA{30, 30, 40, 255})

	// Update canvas with validation result
	a.canvas.SetValidation(a.validation)

	// Draw tilemap canvas (left portion of screen)
	a.canvas.Draw(screen)

	// Draw tile palette (right sidebar, upper portion)
	screenWidth, screenHeight := screen.Size()
	tilePaletteX := screenWidth - PaletteWidth - ObjectPaletteWidth
	a.tileset.DrawPaletteAt(screen, a.state.SelectedTile, tilePaletteX)

	// Draw object palette (right sidebar, middle portion)
	a.objectPalette.Draw(screen, 0)

	// Draw properties panel (right sidebar, bottom portion)
	propertiesStartY := screenHeight - PropertiesPanelHeight
	a.propertiesPanel.SetValidation(a.validation)
	a.propertiesPanel.Draw(screen, propertiesStartY)

	// Draw minimap
	a.drawMinimap(screen)

	// Draw status bar at the bottom
	a.drawStatusBar(screen)

	// Draw status message if visible
	a.drawStatusMessage(screen)

	// Draw help overlay if visible
	if a.showHelp {
		a.drawHelpOverlay(screen)
	}
}

// drawStatusBar draws a simple status bar at the bottom of the screen.
func (a *App) drawStatusBar(screen *ebiten.Image) {
	// For now, just display basic info using the window title
	// A proper text rendering system will be added later
	width, height := screen.Size()
	_ = width
	_ = height

	// Update window title with status
	title := "GoP Level Editor"
	if a.state.HasLevel() {
		if a.state.FilePath != "" {
			title = fmt.Sprintf("GoP Level Editor - %s", a.state.FilePath)
		} else {
			title = "GoP Level Editor - Untitled"
		}
		if a.state.IsModified() {
			title += " *"
		}

		// Add current tool
		toolName := a.getToolName(a.state.CurrentTool)
		title += fmt.Sprintf(" | Tool: %s", toolName)

		// Add current layer with visibility indicator
		layerVisible := a.state.IsLayerVisible(a.state.CurrentLayer)
		visibility := "visible"
		if !layerVisible {
			visibility = "hidden"
		}
		title += fmt.Sprintf(" | Layer: %s (%s)", a.state.CurrentLayer, visibility)

		// Add grid and collision overlay status
		title += fmt.Sprintf(" | Grid: %v, Collision: %v", a.canvas.ShowGrid(), a.canvas.ShowCollision())

		// Add selected tile info
		if a.state.SelectedTile >= 0 {
			title += fmt.Sprintf(" | Tile: %d", a.state.SelectedTile)
		}

		// Add validation status
		if a.validation != nil && a.validation.HasIssues() {
			title += fmt.Sprintf(" | Issues: %d errors, %d warnings", a.validation.ErrorCount(), a.validation.WarningCount())
		}

		// Add undo/redo status
		if a.state.History.CanUndo() {
			title += fmt.Sprintf(" | Undo: %s", a.state.History.UndoDescription())
		}
		if a.state.History.CanRedo() {
			title += fmt.Sprintf(" | Redo: %s", a.state.History.RedoDescription())
		}

		// Add playtest hint
		title += " | P: Playtest"
	}
	ebiten.SetWindowTitle(title)
}

// getToolName returns a human-readable name for a tool.
func (a *App) getToolName(tool Tool) string {
	switch tool {
	case ToolSelect:
		return "Select"
	case ToolPaint:
		return "Paint"
	case ToolErase:
		return "Erase"
	case ToolFill:
		return "Fill"
	case ToolPlaceObject:
		return "Place Object"
	default:
		return "Unknown"
	}
}

// Layout returns the logical screen size.
// This controls the coordinate system for drawing.
func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	// Store screen dimensions for palette/panel calculations
	a.screenWidth = outsideWidth
	a.screenHeight = outsideHeight

	// Delegate to playtest controller if active
	if a.playtest != nil && a.playtest.IsActive() {
		return a.playtest.Layout(outsideWidth, outsideHeight)
	}

	// Use the actual window size for crisp rendering
	return outsideWidth, outsideHeight
}

// State returns the editor state for external access.
func (a *App) State() *EditorState {
	return a.state
}

// Camera returns the camera for external access.
func (a *App) Camera() *Camera {
	return a.camera
}

// Tileset returns the tileset for external access.
func (a *App) Tileset() *Tileset {
	return a.tileset
}

// Canvas returns the canvas for external access.
func (a *App) Canvas() *Canvas {
	return a.canvas
}

// PropertiesPanel returns the properties panel for external access.
func (a *App) PropertiesPanel() *PropertiesPanel {
	return a.propertiesPanel
}

// Validation returns the last validation result for external access.
func (a *App) Validation() *ValidationResult {
	return a.validation
}

// Playtest returns the playtest controller for external access.
func (a *App) Playtest() *PlaytestController {
	return a.playtest
}

// ScreenSize returns the current screen dimensions.
func (a *App) ScreenSize() (width, height int) {
	return a.screenWidth, a.screenHeight
}

// drawHelpOverlay draws the keyboard shortcuts reference panel.
func (a *App) drawHelpOverlay(screen *ebiten.Image) {
	screenWidth, screenHeight := screen.Size()

	// Semi-transparent background
	overlayWidth := 400
	overlayHeight := 450
	overlayX := (screenWidth - overlayWidth) / 2
	overlayY := (screenHeight - overlayHeight) / 2

	// Draw background
	overlayImg := ebiten.NewImage(overlayWidth, overlayHeight)
	overlayImg.Fill(color.RGBA{40, 40, 50, 240})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(overlayX), float64(overlayY))
	screen.DrawImage(overlayImg, op)

	// Draw border
	borderColor := color.RGBA{100, 100, 120, 255}
	ebitenutil.DrawRect(screen, float64(overlayX), float64(overlayY), float64(overlayWidth), 2, borderColor)
	ebitenutil.DrawRect(screen, float64(overlayX), float64(overlayY+overlayHeight-2), float64(overlayWidth), 2, borderColor)
	ebitenutil.DrawRect(screen, float64(overlayX), float64(overlayY), 2, float64(overlayHeight), borderColor)
	ebitenutil.DrawRect(screen, float64(overlayX+overlayWidth-2), float64(overlayY), 2, float64(overlayHeight), borderColor)

	// Title
	titleY := overlayY + 15
	ebitenutil.DebugPrintAt(screen, "KEYBOARD SHORTCUTS", overlayX+120, titleY)

	// Shortcuts list
	shortcuts := []struct {
		key    string
		action string
	}{
		{"--- File Operations ---", ""},
		{"Ctrl+N", "New Level"},
		{"Ctrl+O", "Open Level"},
		{"Ctrl+S", "Save Level"},
		{"Ctrl+Shift+S", "Save As"},
		{"--- Tools ---", ""},
		{"1 / S", "Select Tool"},
		{"2", "Paint Tool"},
		{"3 / E", "Erase Tool"},
		{"4 / F", "Fill Tool"},
		{"5 / O", "Place Object Tool"},
		{"--- Selection ---", ""},
		{"Shift+Click", "Add to Selection"},
		{"Ctrl+C", "Copy"},
		{"Ctrl+V", "Paste"},
		{"Ctrl+X", "Cut"},
		{"Del/Backspace", "Delete Selected"},
		{"Escape", "Clear Selection"},
		{"--- View ---", ""},
		{"G", "Toggle Grid"},
		{"C", "Toggle Collision"},
		{"H", "Toggle Layer Visibility"},
		{"Tab", "Cycle Layers"},
		{"--- Other ---", ""},
		{"P", "Playtest Mode"},
		{"V", "Validate Level"},
		{"Ctrl+Z", "Undo"},
		{"Ctrl+Y", "Redo"},
		{"F1 / ?", "Toggle This Help"},
	}

	y := titleY + 25
	for _, s := range shortcuts {
		if s.action == "" {
			// Section header
			ebitenutil.DebugPrintAt(screen, s.key, overlayX+20, y)
		} else {
			// Shortcut entry
			ebitenutil.DebugPrintAt(screen, s.key, overlayX+20, y)
			ebitenutil.DebugPrintAt(screen, s.action, overlayX+140, y)
		}
		y += 14
	}

	// Close hint
	ebitenutil.DebugPrintAt(screen, "Press F1 or ? to close", overlayX+120, overlayY+overlayHeight-25)
}

// drawMinimap draws a small overview of the level in the corner.
func (a *App) drawMinimap(screen *ebiten.Image) {
	if a.minimap != nil {
		a.minimap.Draw(screen, a.state, a.camera)
	}
}

// drawStatusMessage draws the status message overlay.
func (a *App) drawStatusMessage(screen *ebiten.Image) {
	if a.state.StatusMessage == nil {
		return
	}

	msg := a.state.StatusMessage
	screenWidth, screenHeight := screen.Size()

	// Calculate message width (approximate)
	msgWidth := len(msg.Text)*7 + 20
	msgHeight := 30
	msgX := (screenWidth - msgWidth) / 2
	msgY := screenHeight - 60

	// Choose color based on message type
	bgColor := color.RGBA{40, 120, 40, 220} // Green for success
	if msg.IsError {
		bgColor = color.RGBA{160, 40, 40, 220} // Red for errors
	}

	// Draw background
	msgImg := ebiten.NewImage(msgWidth, msgHeight)
	msgImg.Fill(bgColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(msgX), float64(msgY))
	screen.DrawImage(msgImg, op)

	// Draw border
	borderColor := color.RGBA{255, 255, 255, 200}
	ebitenutil.DrawRect(screen, float64(msgX), float64(msgY), float64(msgWidth), 2, borderColor)
	ebitenutil.DrawRect(screen, float64(msgX), float64(msgY+msgHeight-2), float64(msgWidth), 2, borderColor)
	ebitenutil.DrawRect(screen, float64(msgX), float64(msgY), 2, float64(msgHeight), borderColor)
	ebitenutil.DrawRect(screen, float64(msgX+msgWidth-2), float64(msgY), 2, float64(msgHeight), borderColor)

	// Draw text
	textX := msgX + 10
	textY := msgY + 8
	ebitenutil.DebugPrintAt(screen, msg.Text, textX, textY)
}
