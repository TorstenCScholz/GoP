package editor

import (
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
