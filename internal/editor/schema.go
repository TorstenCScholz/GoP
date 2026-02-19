package editor

import (
	"fmt"
	"strings"

	"github.com/torsten/GoP/internal/world"
)

// PropertySchema defines the schema for a single object property.
type PropertySchema struct {
	Name     string  // Property name
	Type     string  // Property type: "string", "float", "bool", "int"
	Required bool    // Whether the property is required
	Default  any     // Default value if not specified
	Min      float64 // Minimum value for float/int types
	Max      float64 // Maximum value for float/int types
}

// ObjectSchema defines the schema for an object type.
type ObjectSchema struct {
	Type       string           // Object type (e.g., "spawn", "platform")
	Name       string           // Human-readable name (e.g., "Player Spawn")
	Icon       string           // Icon identifier (for future use)
	DefaultW   float64          // Default width in pixels
	DefaultH   float64          // Default height in pixels
	Properties []PropertySchema // Property schemas
	Color      string           // Color for rendering (hex string)
}

// SchemaRegistry holds all object schemas.
var SchemaRegistry = map[world.ObjectType]*ObjectSchema{
	world.ObjectTypeSpawn: {
		Type:       string(world.ObjectTypeSpawn),
		Name:       "Player Spawn",
		Icon:       "spawn",
		DefaultW:   32,
		DefaultH:   32,
		Color:      "#00FF00", // Green
		Properties: []PropertySchema{
			// Spawn has no optional properties
		},
	},
	world.ObjectTypePlatform: {
		Type:     string(world.ObjectTypePlatform),
		Name:     "Platform",
		Icon:     "platform",
		DefaultW: 64,
		DefaultH: 16,
		Color:    "#8040C0", // Purple
		Properties: []PropertySchema{
			{Name: "id", Type: "string", Required: false, Default: ""},
			{Name: "endX", Type: "float", Required: false, Default: 0.0, Min: 0, Max: 10000},
			{Name: "endY", Type: "float", Required: false, Default: 0.0, Min: 0, Max: 10000},
			{Name: "speed", Type: "float", Required: false, Default: 100.0, Min: 0, Max: 1000},
			{Name: "waitTime", Type: "float", Required: false, Default: 0.5, Min: 0, Max: 10},
			{Name: "pushPlayer", Type: "bool", Required: false, Default: false},
		},
	},
	world.ObjectTypeSwitch: {
		Type:     string(world.ObjectTypeSwitch),
		Name:     "Switch",
		Icon:     "switch",
		DefaultW: 32,
		DefaultH: 32,
		Color:    "#FFC800", // Yellow/Orange
		Properties: []PropertySchema{
			{Name: "door_id", Type: "string", Required: false, Default: ""},
			{Name: "toggle", Type: "bool", Required: false, Default: true},
			{Name: "once", Type: "bool", Required: false, Default: false},
		},
	},
	world.ObjectTypeDoor: {
		Type:     string(world.ObjectTypeDoor),
		Name:     "Door",
		Icon:     "door",
		DefaultW: 32,
		DefaultH: 64,
		Color:    "#0080FF", // Blue
		Properties: []PropertySchema{
			{Name: "id", Type: "string", Required: false, Default: ""},
			{Name: "startOpen", Type: "bool", Required: false, Default: false},
		},
	},
	world.ObjectTypeHazard: {
		Type:       string(world.ObjectTypeHazard),
		Name:       "Hazard",
		Icon:       "hazard",
		DefaultW:   32,
		DefaultH:   32,
		Color:      "#FF0000", // Red
		Properties: []PropertySchema{
			// Hazard has no optional properties
		},
	},
	world.ObjectTypeCheckpoint: {
		Type:     string(world.ObjectTypeCheckpoint),
		Name:     "Checkpoint",
		Icon:     "checkpoint",
		DefaultW: 32,
		DefaultH: 48,
		Color:    "#00FFFF", // Cyan
		Properties: []PropertySchema{
			{Name: "id", Type: "string", Required: false, Default: ""},
		},
	},
	world.ObjectTypeGoal: {
		Type:       string(world.ObjectTypeGoal),
		Name:       "Goal",
		Icon:       "goal",
		DefaultW:   48,
		DefaultH:   64,
		Color:      "#FFD700", // Gold
		Properties: []PropertySchema{
			// Goal has no optional properties
		},
	},
}

// GetSchema returns the schema for an object type.
func GetSchema(typ world.ObjectType) *ObjectSchema {
	return SchemaRegistry[typ]
}

// GetSchemaByString returns the schema for an object type string.
func GetSchemaByString(typ string) *ObjectSchema {
	return SchemaRegistry[world.ObjectType(typ)]
}

// GetAllSchemas returns all object schemas in a consistent order.
func GetAllSchemas() []*ObjectSchema {
	// Return in a consistent order for UI display
	order := []world.ObjectType{
		world.ObjectTypeSpawn,
		world.ObjectTypePlatform,
		world.ObjectTypeSwitch,
		world.ObjectTypeDoor,
		world.ObjectTypeHazard,
		world.ObjectTypeCheckpoint,
		world.ObjectTypeGoal,
	}

	schemas := make([]*ObjectSchema, 0, len(order))
	for _, typ := range order {
		if schema, ok := SchemaRegistry[typ]; ok {
			schemas = append(schemas, schema)
		}
	}
	return schemas
}

// CreateDefaultObject creates an ObjectData with default values for the given type.
func CreateDefaultObject(typ world.ObjectType, x, y float64) world.ObjectData {
	schema := GetSchema(typ)
	if schema == nil {
		return world.ObjectData{
			Type:  typ,
			X:     x,
			Y:     y,
			W:     32,
			H:     32,
			Props: make(map[string]any),
		}
	}

	// Create properties with default values
	props := make(map[string]any)
	for _, propSchema := range schema.Properties {
		if propSchema.Default != nil {
			props[propSchema.Name] = propSchema.Default
		}
	}

	return world.ObjectData{
		Type:  typ,
		X:     x,
		Y:     y,
		W:     schema.DefaultW,
		H:     schema.DefaultH,
		Props: props,
	}
}

// NeedsAutoID returns true if the object type needs an auto-generated ID property.
// These are entities that can be targeted by other entities (e.g., doors, platforms, checkpoints).
func NeedsAutoID(typ world.ObjectType) bool {
	switch typ {
	case world.ObjectTypeDoor, world.ObjectTypePlatform, world.ObjectTypeCheckpoint:
		return true
	default:
		return false
	}
}

// GenerateUniqueID generates a unique string ID for an object type based on existing objects.
// The ID format is: <type>_<number> (e.g., "door_1", "platform_2", "checkpoint_1")
func GenerateUniqueID(typ world.ObjectType, existingObjects []world.ObjectData) string {
	// Determine the prefix based on object type
	prefix := string(typ) + "_"

	// Collect all existing IDs that match this prefix
	existingIDs := make(map[string]bool)
	for _, obj := range existingObjects {
		if obj.Type == typ {
			if id, ok := obj.Props["id"].(string); ok && id != "" {
				existingIDs[id] = true
			}
		}
	}

	// Find the next available number
	for i := 1; ; i++ {
		candidateID := fmt.Sprintf("%s%d", prefix, i)
		if !existingIDs[candidateID] {
			return candidateID
		}
	}
}

// GenerateUniqueIDWithCustomPrefix generates a unique ID with a custom prefix.
// This is useful for creating more semantic IDs like "gate_a" instead of "door_1".
func GenerateUniqueIDWithCustomPrefix(prefix string, existingObjects []world.ObjectData) string {
	// Collect all existing IDs
	existingIDs := make(map[string]bool)
	for _, obj := range existingObjects {
		if id, ok := obj.Props["id"].(string); ok && id != "" {
			existingIDs[id] = true
		}
	}

	// Find the next available number
	for i := 1; ; i++ {
		candidateID := fmt.Sprintf("%s_%d", prefix, i)
		if !existingIDs[candidateID] {
			return candidateID
		}
	}
}

// CreateObjectWithAutoID creates an ObjectData with default values and an auto-generated unique ID.
// This should be used when placing new objects that need IDs (doors, platforms, checkpoints).
func CreateObjectWithAutoID(typ world.ObjectType, x, y float64, existingObjects []world.ObjectData) world.ObjectData {
	obj := CreateDefaultObject(typ, x, y)

	// Auto-generate ID for types that need it
	if NeedsAutoID(typ) {
		if obj.Props == nil {
			obj.Props = make(map[string]any)
		}
		obj.Props["id"] = GenerateUniqueID(typ, existingObjects)
	}

	return obj
}

// GetAllExistingIDs returns all existing ID property values from objects.
// Useful for validation and ID conflict detection.
func GetAllExistingIDs(objects []world.ObjectData) []string {
	var ids []string
	for _, obj := range objects {
		if id, ok := obj.Props["id"].(string); ok && id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

// ValidateUniqueID checks if the given ID is unique among existing objects.
// Returns true if the ID is unique, false if it conflicts.
func ValidateUniqueID(id string, objects []world.ObjectData, excludeIndex int) bool {
	for i, obj := range objects {
		if i == excludeIndex {
			continue
		}
		if existingID, ok := obj.Props["id"].(string); ok && existingID == id {
			return false
		}
	}
	return true
}

// GetIDPrefixForType returns the recommended ID prefix for an object type.
// This can be used for generating semantic IDs.
func GetIDPrefixForType(typ world.ObjectType) string {
	switch typ {
	case world.ObjectTypeDoor:
		return "door"
	case world.ObjectTypePlatform:
		return "platform"
	case world.ObjectTypeCheckpoint:
		return "cp"
	case world.ObjectTypeSwitch:
		return "switch"
	default:
		return strings.ToLower(string(typ))
	}
}
