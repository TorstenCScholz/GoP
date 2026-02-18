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
	state         *EditorState
	editorState   PropertyEditorState
	editingIndex  int               // Index of property being edited
	editingBuffer string            // Text buffer for editing
	editingProp   string            // Name of property being edited
	scrollOffset  int               // Scroll offset for long property lists
	hoveredRow    int               // Index of hovered property row (-1 if none)
	validation    *ValidationResult // Current validation result
}

// NewPropertiesPanel creates a new properties panel.
func NewPropertiesPanel(state *EditorState) *PropertiesPanel {
	return &PropertiesPanel{
		state:        state,
		editorState:  PropertyEditorIdle,
		editingIndex: -1,
		hoveredRow:   -1,
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
	propY = p.drawBuiltInProperty(screen, panelX, propY, "X", obj.X)
	propY = p.drawBuiltInProperty(screen, panelX, propY, "Y", obj.Y)
	propY = p.drawBuiltInProperty(screen, panelX, propY, "Width", obj.W)
	propY = p.drawBuiltInProperty(screen, panelX, propY, "Height", obj.H)

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
func (p *PropertiesPanel) drawBuiltInProperty(screen *ebiten.Image, panelX, y int, name string, value float64) int {
	labelX := panelX + PropertyPadding
	valueX := panelX + PropertyPadding + PropertyLabelWidth

	// Draw label
	ebitenutil.DebugPrintAt(screen, name+":", labelX, y)

	// Draw value
	valueText := fmt.Sprintf("%.0f", value)
	ebitenutil.DebugPrintAt(screen, valueText, valueX, y)

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

// cancelEdit cancels the current edit operation.
func (p *PropertiesPanel) cancelEdit() {
	p.editorState = PropertyEditorIdle
	p.editingIndex = -1
	p.editingProp = ""
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

// moveToNextProperty moves editing to the next property in the schema.
func (p *PropertiesPanel) moveToNextProperty() {
	obj := p.state.GetSelectedObject()
	if obj == nil {
		return
	}

	schema := GetSchema(obj.Type)
	if schema == nil || len(schema.Properties) == 0 {
		return
	}

	// Find next property index
	nextIndex := p.editingIndex + 1
	if nextIndex >= len(schema.Properties) {
		nextIndex = 0 // Wrap around
	}

	// Start editing the next property
	if nextIndex < len(schema.Properties) {
		p.startEdit(obj, schema.Properties[nextIndex], nextIndex)
	}
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
	propY := headerY + PropertyRowHeight + 5

	// Skip built-in properties (X, Y, Width, Height) - 4 rows
	propY += 4 * PropertyRowHeight

	// Skip separator
	propY += 15

	// Check each custom property
	for i, propSchema := range schema.Properties {
		rowTop := propY + i*PropertyRowHeight
		rowBottom := rowTop + PropertyRowHeight

		if screenY >= rowTop && screenY < rowBottom {
			// Click is on this property
			valueX := panelX + PropertyPadding + PropertyLabelWidth
			if screenX >= valueX {
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
	propY := headerY + PropertyRowHeight + 5

	// Skip built-in properties (X, Y, Width, Height) - 4 rows
	propY += 4 * PropertyRowHeight

	// Skip separator
	propY += 15

	// Check each custom property
	for i := range schema.Properties {
		rowTop := propY + i*PropertyRowHeight
		rowBottom := rowTop + PropertyRowHeight

		if screenY >= rowTop && screenY < rowBottom {
			valueX := panelX + PropertyPadding + PropertyLabelWidth
			if screenX >= valueX {
				p.hoveredRow = i
			}
			return
		}
	}
}

// IsEditing returns true if a property is currently being edited.
func (p *PropertiesPanel) IsEditing() bool {
	return p.editorState == PropertyEditorActive
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
