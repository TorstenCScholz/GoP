package editor

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/torsten/GoP/internal/world"
)

// Link button dimensions
const (
	LinkButtonWidth  = 60
	LinkButtonHeight = PropertyRowHeight - 4
)

const (
	// PropertiesPanelHeight is the height of the properties panel in pixels.
	PropertiesPanelHeight = 250
	// PropertyRowHeight is the height of each property row in pixels.
	PropertyRowHeight = 22
	// PropertyLabelWidth is the width of the property label column.
	PropertyLabelWidth = 70
	// PropertyPadding is the padding inside the properties panel.
	PropertyPadding = 10
)

// PropertyEditorState represents the editing state for a property.
type PropertyEditorState int

const (
	// PropertyEditorIdle means no property is being edited.
	PropertyEditorIdle PropertyEditorState = iota
	// PropertyEditorActive means a property is being edited.
	PropertyEditorActive
)

// PropertiesPanel handles rendering and interaction for the property editing panel.
type PropertiesPanel struct {
	state             *EditorState
	editorState       PropertyEditorState
	editingIndex      int                   // Index of property being edited
	editingBuffer     string                // Text buffer for editing
	editingProp       string                // Name of property being edited
	editingBuiltIn    string                // Name of built-in property being edited ("X", "Y", "Width", "Height", or "")
	scrollOffset      int                   // Scroll offset for long property lists
	hoveredRow        int                   // Index of hovered property row (-1 if none)
	hoveredBuiltInRow int                   // Index of hovered built-in row (-1 if none)
	validation        *ValidationResult     // Current validation result
	linkButtonHovered bool                  // True if the "Link to Door" button is hovered
	OnStartLinkMode   func(switchIndex int) // Callback when link mode is requested
}

// NewPropertiesPanel creates a new properties panel.
func NewPropertiesPanel(state *EditorState) *PropertiesPanel {
	return &PropertiesPanel{
		state:             state,
		editorState:       PropertyEditorIdle,
		editingIndex:      -1,
		hoveredRow:        -1,
		hoveredBuiltInRow: -1,
	}
}

// Draw renders the properties panel to the screen.
// The panel is drawn at the bottom of the right sidebar.
func (p *PropertiesPanel) Draw(screen *ebiten.Image, startY int) {
	screenWidth, screenHeight := screen.Size()
	_ = screenHeight

	// Calculate panel position (right side, below object palette)
	panelX := screenWidth - ObjectPaletteWidth
	panelHeight := PropertiesPanelHeight

	// Draw panel background
	panelBg := ebiten.NewImage(ObjectPaletteWidth, panelHeight)
	panelBg.Fill(propertiesPanelBgColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(panelX), float64(startY))
	screen.DrawImage(panelBg, op)

	// Draw border at top
	borderImg := ebiten.NewImage(ObjectPaletteWidth, 1)
	borderImg.Fill(propertiesBorderColor)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(panelX), float64(startY))
	screen.DrawImage(borderImg, op)

	// Draw title
	titleY := startY + PropertyPadding
	ebitenutil.DebugPrintAt(screen, "Properties", panelX+PropertyPadding, titleY)

	// Check if an object is selected
	if !p.state.HasSelection() {
		// Show "No object selected" message
		msgY := titleY + 30
		ebitenutil.DebugPrintAt(screen, "No object selected", panelX+PropertyPadding, msgY)
		return
	}

	// Get selected object
	obj := p.state.GetSelectedObject()
	if obj == nil {
		return
	}

	// Draw object type header
	headerY := titleY + 20
	schema := GetSchema(obj.Type)
	typeName := "Unknown"
	if schema != nil {
		typeName = schema.Name
	}
	typeText := fmt.Sprintf("Type: %s", typeName)
	ebitenutil.DebugPrintAt(screen, typeText, panelX+PropertyPadding, headerY)

	// Draw built-in properties (X, Y, Width, Height)
	propY := headerY + PropertyRowHeight + 5
	propY = p.drawBuiltInProperty(screen, panelX, propY, "X", obj.X, 0)
	propY = p.drawBuiltInProperty(screen, panelX, propY, "Y", obj.Y, 1)
	propY = p.drawBuiltInProperty(screen, panelX, propY, "Width", obj.W, 2)
	propY = p.drawBuiltInProperty(screen, panelX, propY, "Height", obj.H, 3)

	// Draw separator
	sepY := propY + 5
	sepImg := ebiten.NewImage(ObjectPaletteWidth-2*PropertyPadding, 1)
	sepImg.Fill(propertiesSeparatorColor)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(panelX+PropertyPadding), float64(sepY))
	screen.DrawImage(sepImg, op)

	// Draw custom properties from schema
	propY = sepY + 10
	if schema != nil {
		for i, propSchema := range schema.Properties {
			propY = p.drawCustomProperty(screen, panelX, propY, obj, propSchema, i)
		}
	}

	// Draw validation errors for selected object
	if p.validation != nil && p.state.HasSelection() {
		issues := p.validation.GetObjectIssues(p.state.SelectedObject)
		if len(issues) > 0 {
			p.drawValidationIssues(screen, panelX, propY, issues)
		}
	}
}

// drawBuiltInProperty draws a built-in property row (X, Y, Width, Height).
func (p *PropertiesPanel) drawBuiltInProperty(screen *ebiten.Image, panelX, y int, name string, value float64, rowIndex int) int {
	labelX := panelX + PropertyPadding
	valueX := panelX + PropertyPadding + PropertyLabelWidth
	valueWidth := ObjectPaletteWidth - 2*PropertyPadding - PropertyLabelWidth

	// Draw label
	ebitenutil.DebugPrintAt(screen, name+":", labelX, y)

	// Check if this row is being edited
	if p.editorState == PropertyEditorActive && p.editingBuiltIn == name {
		// Draw input field background
		inputBg := ebiten.NewImage(valueWidth, PropertyRowHeight-4)
		inputBg.Fill(propertyInputBgColor)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(valueX), float64(y))
		screen.DrawImage(inputBg, op)

		// Draw editing buffer with cursor
		displayText := p.editingBuffer + "|"
		ebitenutil.DebugPrintAt(screen, displayText, valueX+2, y)
	} else {
		// Check if hovered
		if p.hoveredBuiltInRow == rowIndex {
			hoverBg := ebiten.NewImage(valueWidth, PropertyRowHeight-4)
			hoverBg.Fill(propertyHoverColor)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(valueX), float64(y))
			screen.DrawImage(hoverBg, op)
		}

		// Draw value
		valueText := fmt.Sprintf("%.0f", value)
		ebitenutil.DebugPrintAt(screen, valueText, valueX, y)
	}

	return y + PropertyRowHeight
}

// drawCustomProperty draws a custom property row with editing support.
func (p *PropertiesPanel) drawCustomProperty(screen *ebiten.Image, panelX, y int, obj *world.ObjectData, propSchema PropertySchema, index int) int {
	labelX := panelX + PropertyPadding
	valueX := panelX + PropertyPadding + PropertyLabelWidth
	valueWidth := ObjectPaletteWidth - 2*PropertyPadding - PropertyLabelWidth

	// Draw label
	ebitenutil.DebugPrintAt(screen, propSchema.Name+":", labelX, y)

	// Get current value
	value := p.getPropertyValue(obj, propSchema)

	// Check if this is the door_id property of a switch - add link button
	isDoorIDProp := propSchema.Name == "door_id" && obj.Type == world.ObjectTypeSwitch

	// Adjust value width if we need to add a button
	buttonWidth := 0
	if isDoorIDProp {
		buttonWidth = 60 // Width for "Link" button
		valueWidth -= buttonWidth + 5
	}

	// Check if this row is being edited
	if p.editorState == PropertyEditorActive && p.editingProp == propSchema.Name {
		// Draw input field background
		inputBg := ebiten.NewImage(valueWidth, PropertyRowHeight-4)
		inputBg.Fill(propertyInputBgColor)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(valueX), float64(y))
		screen.DrawImage(inputBg, op)

		// Draw editing buffer with cursor
		displayText := p.editingBuffer + "|"
		ebitenutil.DebugPrintAt(screen, displayText, valueX+2, y)
	} else {
		// Check if hovered
		if p.hoveredRow == index {
			// Draw hover background
			hoverBg := ebiten.NewImage(valueWidth, PropertyRowHeight-4)
			hoverBg.Fill(propertyHoverColor)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(valueX), float64(y))
			screen.DrawImage(hoverBg, op)
		}

		// Draw value based on type
		switch propSchema.Type {
		case "string":
			strVal, _ := value.(string)
			if strVal == "" {
				strVal = "(empty)"
			}
			ebitenutil.DebugPrintAt(screen, strVal, valueX, y)
		case "float":
			floatVal, _ := value.(float64)
			valueText := fmt.Sprintf("%.2f", floatVal)
			ebitenutil.DebugPrintAt(screen, valueText, valueX, y)
		case "bool":
			boolVal, _ := value.(bool)
			checkText := "[ ]"
			if boolVal {
				checkText = "[x]"
			}
			ebitenutil.DebugPrintAt(screen, checkText, valueX, y)
		}
	}

	// Draw "Link" button for door_id property on switches
	if isDoorIDProp {
		buttonX := valueX + valueWidth + 5
		buttonHeight := PropertyRowHeight - 4

		// Determine button color based on hover state
		buttonColor := linkButtonColor
		if p.linkButtonHovered {
			buttonColor = linkButtonHoverColor
		}

		// Draw button background
		buttonImg := ebiten.NewImage(buttonWidth, buttonHeight)
		buttonImg.Fill(buttonColor)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(buttonX), float64(y))
		screen.DrawImage(buttonImg, op)

		// Draw button border
		borderColor := color.RGBA{100, 100, 120, 255}
		ebitenutil.DrawRect(screen, float64(buttonX), float64(y), float64(buttonWidth), 1, borderColor)
		ebitenutil.DrawRect(screen, float64(buttonX), float64(y+buttonHeight-1), float64(buttonWidth), 1, borderColor)
		ebitenutil.DrawRect(screen, float64(buttonX), float64(y), 1, float64(buttonHeight), borderColor)
		ebitenutil.DrawRect(screen, float64(buttonX+buttonWidth-1), float64(y), 1, float64(buttonHeight), borderColor)

		// Draw button text
		ebitenutil.DebugPrintAt(screen, "Link", buttonX+8, y+2)
	}

	return y + PropertyRowHeight
}

// getPropertyValue returns the value of a property from the object.
func (p *PropertiesPanel) getPropertyValue(obj *world.ObjectData, propSchema PropertySchema) any {
	if obj.Props == nil {
		return propSchema.Default
	}
	if val, ok := obj.Props[propSchema.Name]; ok {
		return val
	}
	return propSchema.Default
}

// Update handles input for the properties panel.
func (p *PropertiesPanel) Update() bool {
	if p.editorState == PropertyEditorActive {
		return p.handleEditingInput()
	}
	return p.handleNavigationInput()
}

// handleNavigationInput handles input when not editing a property.
func (p *PropertiesPanel) handleNavigationInput() bool {
	mx, my := ebiten.CursorPosition()
	screenWidth := 0
	// We need to get screen width from somewhere - for now use a reasonable default
	// This will be properly handled when integrated with app.go

	// Check for mouse click to start editing or toggle bool
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Calculate if click is in properties panel area
		// This will be handled by the app.go integration
		_ = screenWidth
		_ = mx
		_ = my
	}
	return false
}

// handleEditingInput handles input when editing a property.
func (p *PropertiesPanel) handleEditingInput() bool {
	// Handle text input
	var inputChars []rune
	inputChars = ebiten.AppendInputChars(inputChars)
	for _, c := range inputChars {
		p.editingBuffer += string(c)
	}

	// Handle backspace
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if len(p.editingBuffer) > 0 {
			p.editingBuffer = p.editingBuffer[:len(p.editingBuffer)-1]
		}
	}

	// Handle Enter to confirm
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		p.confirmEdit()
		return true
	}

	// Handle Escape to cancel
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		p.cancelEdit()
		return false
	}

	// Handle Tab to move to next property
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		p.confirmEdit()
		p.moveToNextProperty()
		return true
	}

	return false
}

// startEdit begins editing a property.
func (p *PropertiesPanel) startEdit(obj *world.ObjectData, propSchema PropertySchema, index int) {
	p.editorState = PropertyEditorActive
	p.editingIndex = index
	p.editingProp = propSchema.Name

	// Get current value and convert to string for editing
	value := p.getPropertyValue(obj, propSchema)
	switch propSchema.Type {
	case "string":
		strVal, _ := value.(string)
		p.editingBuffer = strVal
	case "float":
		floatVal, _ := value.(float64)
		p.editingBuffer = fmt.Sprintf("%.2f", floatVal)
	case "bool":
		// Bool properties don't use text editing - they toggle directly
		p.toggleBoolProperty(obj, propSchema)
		p.editorState = PropertyEditorIdle
		p.editingIndex = -1
		p.editingProp = ""
	default:
		p.editingBuffer = fmt.Sprintf("%v", value)
	}
}

// startBuiltInEdit begins editing a built-in property (X, Y, Width, Height).
func (p *PropertiesPanel) startBuiltInEdit(obj *world.ObjectData, name string) {
	p.editorState = PropertyEditorActive
	p.editingBuiltIn = name
	p.editingProp = ""
	p.editingIndex = -1

	// Get current value
	var value float64
	switch name {
	case "X":
		value = obj.X
	case "Y":
		value = obj.Y
	case "Width":
		value = obj.W
	case "Height":
		value = obj.H
	}
	p.editingBuffer = fmt.Sprintf("%.0f", value)
}

// confirmEdit applies the edited value to the object.
func (p *PropertiesPanel) confirmEdit() {
	if p.editorState != PropertyEditorActive {
		return
	}

	obj := p.state.GetSelectedObject()
	if obj == nil {
		p.cancelEdit()
		return
	}

	// Handle built-in property editing
	if p.editingBuiltIn != "" {
		p.confirmBuiltInEdit(obj)
		return
	}

	schema := GetSchema(obj.Type)
	if schema == nil {
		p.cancelEdit()
		return
	}

	// Find the property schema
	var propSchema *PropertySchema
	for i, ps := range schema.Properties {
		if i == p.editingIndex || ps.Name == p.editingProp {
			propSchema = &ps
			break
		}
	}

	if propSchema == nil {
		p.cancelEdit()
		return
	}

	// Parse and apply the value based on type
	switch propSchema.Type {
	case "string":
		// Get old value
		var oldValue any
		if obj.Props != nil {
			oldValue = obj.Props[propSchema.Name]
		}
		// Create and execute action
		action := NewSetPropertyAction(p.state.SelectedObject, propSchema.Name, oldValue, p.editingBuffer)
		p.state.History.Do(action, p.state)
	case "float":
		floatVal, err := strconv.ParseFloat(p.editingBuffer, 64)
		if err == nil {
			// Clamp to min/max
			if propSchema.Min != 0 || propSchema.Max != 0 {
				if floatVal < propSchema.Min {
					floatVal = propSchema.Min
				}
				if floatVal > propSchema.Max {
					floatVal = propSchema.Max
				}
			}
			// Get old value
			var oldValue any
			if obj.Props != nil {
				oldValue = obj.Props[propSchema.Name]
			}
			// Create and execute action
			action := NewSetPropertyAction(p.state.SelectedObject, propSchema.Name, oldValue, floatVal)
			p.state.History.Do(action, p.state)
		}
	}

	p.editorState = PropertyEditorIdle
	p.editingIndex = -1
	p.editingProp = ""
	p.editingBuffer = ""
}

// confirmBuiltInEdit applies the edited value for a built-in property.
func (p *PropertiesPanel) confirmBuiltInEdit(obj *world.ObjectData) {
	floatVal, err := strconv.ParseFloat(strings.TrimSpace(p.editingBuffer), 64)
	if err != nil {
		p.cancelEdit()
		return
	}

	idx := p.state.SelectedObject

	switch p.editingBuiltIn {
	case "X":
		if floatVal != obj.X {
			action := NewMoveObjectAction(idx, obj.X, obj.Y, floatVal, obj.Y)
			p.state.History.Do(action, p.state)
		}
	case "Y":
		if floatVal != obj.Y {
			action := NewMoveObjectAction(idx, obj.X, obj.Y, obj.X, floatVal)
			p.state.History.Do(action, p.state)
		}
	case "Width":
		if floatVal > 0 && floatVal != obj.W {
			action := NewResizeObjectAction(idx, obj.W, obj.H, floatVal, obj.H)
			p.state.History.Do(action, p.state)
		}
	case "Height":
		if floatVal > 0 && floatVal != obj.H {
			action := NewResizeObjectAction(idx, obj.W, obj.H, obj.W, floatVal)
			p.state.History.Do(action, p.state)
		}
	}

	p.editorState = PropertyEditorIdle
	p.editingBuiltIn = ""
	p.editingIndex = -1
	p.editingProp = ""
	p.editingBuffer = ""
}

// cancelEdit cancels the current edit operation.
func (p *PropertiesPanel) cancelEdit() {
	p.editorState = PropertyEditorIdle
	p.editingIndex = -1
	p.editingProp = ""
	p.editingBuiltIn = ""
	p.editingBuffer = ""
}

// toggleBoolProperty toggles a boolean property value.
func (p *PropertiesPanel) toggleBoolProperty(obj *world.ObjectData, propSchema PropertySchema) {
	if obj.Props == nil {
		obj.Props = make(map[string]any)
	}

	currentVal := false
	if val, ok := obj.Props[propSchema.Name]; ok {
		if boolVal, ok := val.(bool); ok {
			currentVal = boolVal
		}
	}

	// Create and execute action
	action := NewSetPropertyAction(p.state.SelectedObject, propSchema.Name, currentVal, !currentVal)
	p.state.History.Do(action, p.state)
}

// moveToNextProperty moves editing to the next property.
// Cycles: X → Y → Width → Height → first custom property → ... → X
func (p *PropertiesPanel) moveToNextProperty() {
	obj := p.state.GetSelectedObject()
	if obj == nil {
		return
	}

	builtInNames := []string{"X", "Y", "Width", "Height"}

	schema := GetSchema(obj.Type)

	// If we were editing a built-in property, move to the next one
	if p.editingBuiltIn != "" {
		for i, name := range builtInNames {
			if name == p.editingBuiltIn {
				nextIdx := i + 1
				if nextIdx < len(builtInNames) {
					// Move to next built-in
					p.startBuiltInEdit(obj, builtInNames[nextIdx])
					return
				}
				// Past last built-in — move to first custom property
				if schema != nil && len(schema.Properties) > 0 {
					p.startEdit(obj, schema.Properties[0], 0)
					return
				}
				// No custom properties, wrap to first built-in
				p.startBuiltInEdit(obj, builtInNames[0])
				return
			}
		}
	}

	// We were editing a custom property
	if schema == nil || len(schema.Properties) == 0 {
		// No custom properties, wrap to first built-in
		p.startBuiltInEdit(obj, builtInNames[0])
		return
	}

	nextIndex := p.editingIndex + 1
	if nextIndex >= len(schema.Properties) {
		// Wrap to first built-in
		p.startBuiltInEdit(obj, builtInNames[0])
		return
	}

	p.startEdit(obj, schema.Properties[nextIndex], nextIndex)
}

// HandleClick processes mouse clicks in the properties panel.
// Returns true if the click was handled by the panel.
func (p *PropertiesPanel) HandleClick(screenX, screenY, screenWidth, startY int) bool {
	// Check if click is in properties panel area
	panelX := screenWidth - ObjectPaletteWidth
	panelEndY := startY + PropertiesPanelHeight

	if screenX < panelX || screenY < startY || screenY >= panelEndY {
		return false
	}

	// If currently editing, confirm the edit first
	if p.editorState == PropertyEditorActive {
		p.confirmEdit()
	}

	// Check if an object is selected
	if !p.state.HasSelection() {
		return true // Click was in panel but nothing to edit
	}

	obj := p.state.GetSelectedObject()
	if obj == nil {
		return true
	}

	schema := GetSchema(obj.Type)
	if schema == nil {
		return true
	}

	// Calculate property row positions
	titleY := startY + PropertyPadding
	headerY := titleY + 20
	builtInStartY := headerY + PropertyRowHeight + 5

	// Check built-in property clicks (X, Y, Width, Height)
	valueX := panelX + PropertyPadding + PropertyLabelWidth
	builtInNames := []string{"X", "Y", "Width", "Height"}
	for i, name := range builtInNames {
		rowTop := builtInStartY + i*PropertyRowHeight
		if screenY >= rowTop && screenY < rowTop+PropertyRowHeight {
			if screenX >= valueX {
				p.startBuiltInEdit(obj, name)
			}
			return true
		}
	}

	propY := builtInStartY + 4*PropertyRowHeight

	// Skip separator
	propY += 15

	// Check each custom property
	for i, propSchema := range schema.Properties {
		rowTop := propY + i*PropertyRowHeight
		rowBottom := rowTop + PropertyRowHeight

		if screenY >= rowTop && screenY < rowBottom {
			// Click is on this property
			valueX := panelX + PropertyPadding + PropertyLabelWidth
			valueWidth := ObjectPaletteWidth - 2*PropertyPadding - PropertyLabelWidth

			// Check if this is the door_id property with a link button
			isDoorIDProp := propSchema.Name == "door_id" && obj.Type == world.ObjectTypeSwitch
			if isDoorIDProp {
				buttonWidth := 60
				valueWidth -= buttonWidth + 5
				buttonX := valueX + valueWidth + 5

				// Check if click is on the link button
				if screenX >= buttonX && screenX < buttonX+buttonWidth {
					// Link button clicked - trigger callback
					if p.OnStartLinkMode != nil {
						p.OnStartLinkMode(p.state.SelectedObject)
					}
					return true
				}
			}

			if screenX >= valueX && screenX < valueX+valueWidth {
				// Click is on the value area - start editing
				p.startEdit(obj, propSchema, i)
			}
			return true
		}
	}

	return true
}

// HandleMouseMove processes mouse movement for hover effects.
func (p *PropertiesPanel) HandleMouseMove(screenX, screenY, screenWidth, startY int) {
	p.hoveredRow = -1
	p.hoveredBuiltInRow = -1
	p.linkButtonHovered = false

	// Check if mouse is in properties panel area
	panelX := screenWidth - ObjectPaletteWidth
	panelEndY := startY + PropertiesPanelHeight

	if screenX < panelX || screenY < startY || screenY >= panelEndY {
		return
	}

	// Check if an object is selected
	if !p.state.HasSelection() {
		return
	}

	obj := p.state.GetSelectedObject()
	if obj == nil {
		return
	}

	schema := GetSchema(obj.Type)
	if schema == nil {
		return
	}

	// Calculate property row positions
	titleY := startY + PropertyPadding
	headerY := titleY + 20
	builtInStartY := headerY + PropertyRowHeight + 5

	// Check built-in property rows
	valueX := panelX + PropertyPadding + PropertyLabelWidth
	for i := 0; i < 4; i++ {
		rowTop := builtInStartY + i*PropertyRowHeight
		if screenY >= rowTop && screenY < rowTop+PropertyRowHeight {
			if screenX >= valueX {
				p.hoveredBuiltInRow = i
			}
			return
		}
	}

	propY := builtInStartY + 4*PropertyRowHeight

	// Skip separator
	propY += 15

	// Check each custom property
	for i, propSchema := range schema.Properties {
		rowTop := propY + i*PropertyRowHeight
		rowBottom := rowTop + PropertyRowHeight

		if screenY >= rowTop && screenY < rowBottom {
			valueX := panelX + PropertyPadding + PropertyLabelWidth
			valueWidth := ObjectPaletteWidth - 2*PropertyPadding - PropertyLabelWidth

			// Check if this is the door_id property with a link button
			isDoorIDProp := propSchema.Name == "door_id" && obj.Type == world.ObjectTypeSwitch
			if isDoorIDProp {
				buttonWidth := 60
				valueWidth -= buttonWidth + 5
				buttonX := valueX + valueWidth + 5

				// Check if hovering over the link button
				if screenX >= buttonX && screenX < buttonX+buttonWidth {
					p.linkButtonHovered = true
					return
				}
			}

			if screenX >= valueX {
				p.hoveredRow = i
			}
			return
		}
	}
}

// IsEditing returns true if a property is currently being edited.
func (p *PropertiesPanel) IsEditing() bool {
	return p.editorState == PropertyEditorActive || p.editingBuiltIn != ""
}

// IsInPanel returns true if the given screen coordinates are within the properties panel area.
func (p *PropertiesPanel) IsInPanel(screenX, screenY, screenWidth, startY int) bool {
	panelX := screenWidth - ObjectPaletteWidth
	panelEndY := startY + PropertiesPanelHeight
	return screenX >= panelX && screenY >= startY && screenY < panelEndY
}

// SetState updates the editor state reference.
func (p *PropertiesPanel) SetState(state *EditorState) {
	p.state = state
}

// SetValidation sets the current validation result for error display.
func (p *PropertiesPanel) SetValidation(result *ValidationResult) {
	p.validation = result
}

// Validation returns the current validation result.
func (p *PropertiesPanel) Validation() *ValidationResult {
	return p.validation
}

// Colors for properties panel rendering
var (
	propertiesPanelBgColor   = color.RGBA{40, 40, 50, 255}
	propertiesBorderColor    = color.RGBA{60, 60, 70, 255}
	propertiesSeparatorColor = color.RGBA{80, 80, 90, 255}
	propertyInputBgColor     = color.RGBA{30, 30, 40, 255}
	propertyHoverColor       = color.RGBA{70, 70, 80, 255}
	linkButtonColor          = color.RGBA{60, 100, 140, 255}
	linkButtonHoverColor     = color.RGBA{80, 130, 180, 255}
)

// Helper function to check if a string is empty or whitespace only
func isEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

// drawValidationIssues draws validation errors/warnings for the selected object.
func (p *PropertiesPanel) drawValidationIssues(screen *ebiten.Image, panelX, y int, issues []ValidationError) int {
	// Draw separator
	sepY := y + 5
	sepImg := ebiten.NewImage(ObjectPaletteWidth-2*PropertyPadding, 1)
	sepImg.Fill(propertiesSeparatorColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(panelX+PropertyPadding), float64(sepY))
	screen.DrawImage(sepImg, op)

	// Draw "Issues" header
	headerY := sepY + 10
	ebitenutil.DebugPrintAt(screen, "Issues:", panelX+PropertyPadding, headerY)

	// Draw each issue
	issueY := headerY + PropertyRowHeight
	for _, issue := range issues {
		// Determine color based on type
		if issue.Type == TypeError {
			ebitenutil.DebugPrintAt(screen, "! "+issue.Message, panelX+PropertyPadding, issueY)
		} else {
			ebitenutil.DebugPrintAt(screen, "? "+issue.Message, panelX+PropertyPadding, issueY)
		}
		issueY += PropertyRowHeight

		// Stop if we run out of space
		if issueY > PropertiesPanelHeight-PropertyRowHeight {
			break
		}
	}

	return issueY
}
