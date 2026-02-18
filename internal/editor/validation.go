package editor

import (
	"fmt"

	"github.com/torsten/GoP/internal/world"
)

// ErrorType represents the type of validation issue.
type ErrorType string

const (
	// TypeError represents a critical error that must be fixed.
	TypeError ErrorType = "error"
	// TypeWarning represents a warning that should be fixed.
	TypeWarning ErrorType = "warning"
)

// ValidationError represents a single validation issue found in the level.
type ValidationError struct {
	Type        ErrorType // "error" or "warning"
	ObjectIndex int       // Index of the object with the issue (-1 for global errors)
	Message     string    // Human-readable description of the issue
	Property    string    // Property name if applicable (empty string if not)
}

// ValidationResult holds the result of level validation.
type ValidationResult struct {
	Errors   []ValidationError // Critical errors that must be fixed
	Warnings []ValidationError // Warnings that should be fixed
}

// HasErrors returns true if there are any critical errors.
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are any warnings.
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// HasIssues returns true if there are any errors or warnings.
func (r *ValidationResult) HasIssues() bool {
	return r.HasErrors() || r.HasWarnings()
}

// AllIssues returns all errors and warnings combined.
func (r *ValidationResult) AllIssues() []ValidationError {
	all := make([]ValidationError, 0, len(r.Errors)+len(r.Warnings))
	all = append(all, r.Errors...)
	all = append(all, r.Warnings...)
	return all
}

// ErrorCount returns the number of critical errors.
func (r *ValidationResult) ErrorCount() int {
	return len(r.Errors)
}

// WarningCount returns the number of warnings.
func (r *ValidationResult) WarningCount() int {
	return len(r.Warnings)
}

// GetObjectErrors returns all errors for a specific object index.
func (r *ValidationResult) GetObjectErrors(objectIndex int) []ValidationError {
	var errors []ValidationError
	for _, e := range r.Errors {
		if e.ObjectIndex == objectIndex {
			errors = append(errors, e)
		}
	}
	return errors
}

// GetObjectWarnings returns all warnings for a specific object index.
func (r *ValidationResult) GetObjectWarnings(objectIndex int) []ValidationError {
	var warnings []ValidationError
	for _, w := range r.Warnings {
		if w.ObjectIndex == objectIndex {
			warnings = append(warnings, w)
		}
	}
	return warnings
}

// GetObjectIssues returns all errors and warnings for a specific object index.
func (r *ValidationResult) GetObjectIssues(objectIndex int) []ValidationError {
	var issues []ValidationError
	for _, e := range r.Errors {
		if e.ObjectIndex == objectIndex {
			issues = append(issues, e)
		}
	}
	for _, w := range r.Warnings {
		if w.ObjectIndex == objectIndex {
			issues = append(issues, w)
		}
	}
	return issues
}

// ValidateLevel validates the level data and returns a validation result.
func ValidateLevel(state *EditorState) *ValidationResult {
	result := &ValidationResult{
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationError, 0),
	}

	if state == nil || state.Objects == nil {
		return result
	}

	// Check for spawn points
	validateSpawnPoints(state, result)

	// Check for duplicate IDs
	validateUniqueIDs(state, result)

	// Check switch references
	validateSwitchReferences(state, result)

	// Check for doors without switches
	validateDoorSwitches(state, result)

	// Check for platforms with no movement
	validatePlatforms(state, result)

	// Check for required properties
	validateRequiredProperties(state, result)

	return result
}

// validateSpawnPoints checks for player spawn points.
func validateSpawnPoints(state *EditorState, result *ValidationResult) {
	spawns := world.FilterObjectsByType(state.Objects, world.ObjectTypeSpawn)

	if len(spawns) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Type:        TypeError,
			ObjectIndex: -1,
			Message:     "No player spawn point defined",
			Property:    "",
		})
	} else if len(spawns) > 1 {
		result.Warnings = append(result.Warnings, ValidationError{
			Type:        TypeWarning,
			ObjectIndex: -1,
			Message:     fmt.Sprintf("Multiple spawn points defined (%d), only the first will be used", len(spawns)),
			Property:    "",
		})
	}
}

// validateUniqueIDs checks for duplicate entity IDs.
func validateUniqueIDs(state *EditorState, result *ValidationResult) {
	// Map from ID to the first object index that uses it
	idToIndex := make(map[string]int)

	for i, obj := range state.Objects {
		id := obj.GetPropString("id", "")
		if id == "" {
			continue
		}

		if firstIndex, exists := idToIndex[id]; exists {
			// Add error for the duplicate
			result.Errors = append(result.Errors, ValidationError{
				Type:        TypeError,
				ObjectIndex: i,
				Message:     fmt.Sprintf("Duplicate ID '%s' (first used at object %d)", id, firstIndex),
				Property:    "id",
			})
		} else {
			idToIndex[id] = i
		}
	}
}

// validateSwitchReferences checks that switches reference valid doors.
func validateSwitchReferences(state *EditorState, result *ValidationResult) {
	// Build a map of all door IDs
	doorIDs := make(map[string]int)
	for i, obj := range state.Objects {
		if obj.Type == world.ObjectTypeDoor {
			id := obj.GetPropString("id", "")
			if id != "" {
				doorIDs[id] = i
			}
		}
	}

	// Check each switch's door_id reference
	for i, obj := range state.Objects {
		if obj.Type != world.ObjectTypeSwitch {
			continue
		}

		doorID := obj.GetPropString("door_id", "")
		if doorID == "" {
			// Switch with no door_id - this could be a warning
			result.Warnings = append(result.Warnings, ValidationError{
				Type:        TypeWarning,
				ObjectIndex: i,
				Message:     "Switch has no door_id configured",
				Property:    "door_id",
			})
			continue
		}

		if _, exists := doorIDs[doorID]; !exists {
			result.Errors = append(result.Errors, ValidationError{
				Type:        TypeError,
				ObjectIndex: i,
				Message:     fmt.Sprintf("Switch references non-existent door '%s'", doorID),
				Property:    "door_id",
			})
		}
	}
}

// validateDoorSwitches checks for doors without any switch to control them.
func validateDoorSwitches(state *EditorState, result *ValidationResult) {
	// Build a map of doors that are referenced by switches
	referencedDoors := make(map[string]bool)

	for _, obj := range state.Objects {
		if obj.Type == world.ObjectTypeSwitch {
			doorID := obj.GetPropString("door_id", "")
			if doorID != "" {
				referencedDoors[doorID] = true
			}
		}
	}

	// Check each door
	for i, obj := range state.Objects {
		if obj.Type != world.ObjectTypeDoor {
			continue
		}

		id := obj.GetPropString("id", "")
		if id == "" {
			// Door with no ID - can't be referenced
			result.Warnings = append(result.Warnings, ValidationError{
				Type:        TypeWarning,
				ObjectIndex: i,
				Message:     "Door has no ID configured, cannot be controlled by switches",
				Property:    "id",
			})
			continue
		}

		if !referencedDoors[id] {
			result.Warnings = append(result.Warnings, ValidationError{
				Type:        TypeWarning,
				ObjectIndex: i,
				Message:     fmt.Sprintf("Door '%s' has no switch to control it", id),
				Property:    "id",
			})
		}
	}
}

// validatePlatforms checks for platforms with no movement.
func validatePlatforms(state *EditorState, result *ValidationResult) {
	for i, obj := range state.Objects {
		if obj.Type != world.ObjectTypePlatform {
			continue
		}

		// Get endX and endY relative to start position
		endX := obj.GetPropFloat("endX", 0)
		endY := obj.GetPropFloat("endY", 0)

		// If both are 0, the platform won't move
		if endX == 0 && endY == 0 {
			result.Warnings = append(result.Warnings, ValidationError{
				Type:        TypeWarning,
				ObjectIndex: i,
				Message:     "Platform has endX=0 and endY=0, it won't move",
				Property:    "endX",
			})
		}
	}
}

// validateRequiredProperties checks that all required properties are set.
func validateRequiredProperties(state *EditorState, result *ValidationResult) {
	for i, obj := range state.Objects {
		schema := GetSchema(obj.Type)
		if schema == nil {
			continue
		}

		for _, propSchema := range schema.Properties {
			if !propSchema.Required {
				continue
			}

			// Check if the property is set
			value := getPropertyValue(obj, propSchema.Name)

			// Check if the value is the default (unset) or empty
			if isZeroValue(value, propSchema) {
				result.Errors = append(result.Errors, ValidationError{
					Type:        TypeError,
					ObjectIndex: i,
					Message:     fmt.Sprintf("Required property '%s' is not set", propSchema.Name),
					Property:    propSchema.Name,
				})
			}
		}
	}
}

// getPropertyValue returns the value of a property from the object.
func getPropertyValue(obj world.ObjectData, propName string) any {
	if obj.Props == nil {
		return nil
	}
	return obj.Props[propName]
}

// isZeroValue checks if a value is considered "zero" or unset for its type.
func isZeroValue(value any, propSchema PropertySchema) bool {
	if value == nil {
		return true
	}

	switch propSchema.Type {
	case "string":
		if s, ok := value.(string); ok {
			return s == ""
		}
	case "float":
		if f, ok := value.(float64); ok {
			return f == 0
		}
	case "int":
		if n, ok := value.(int); ok {
			return n == 0
		}
		if n, ok := value.(float64); ok {
			return n == 0
		}
	case "bool":
		// Booleans are never considered "unset" - false is a valid value
		return false
	}

	return false
}

// FormatValidationError formats a validation error for display.
func FormatValidationError(err ValidationError) string {
	if err.Property != "" {
		return fmt.Sprintf("%s (property: %s)", err.Message, err.Property)
	}
	return err.Message
}
