package world

import (
	"encoding/json"
	"fmt"
)

// ObjectType represents the type of entity to spawn.
type ObjectType string

const (
	ObjectTypeSpawn      ObjectType = "spawn"
	ObjectTypeHazard     ObjectType = "hazard"
	ObjectTypeCheckpoint ObjectType = "checkpoint"
	ObjectTypeSwitch     ObjectType = "switch"
	ObjectTypeDoor       ObjectType = "door"
	ObjectTypeGoal       ObjectType = "goal"
	ObjectTypePlatform   ObjectType = "platform"
)

// ObjectData represents a parsed Tiled object.
type ObjectData struct {
	ID    int
	Name  string
	Type  ObjectType
	X, Y  float64
	W, H  float64
	Props map[string]any
}

// GetPropString returns a string property or the default value.
func (o *ObjectData) GetPropString(key, def string) string {
	if v, ok := o.Props[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

// GetPropInt returns an int property or the default value.
func (o *ObjectData) GetPropInt(key string, def int) int {
	if v, ok := o.Props[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return def
}

// GetPropBool returns a bool property or the default value.
func (o *ObjectData) GetPropBool(key string, def bool) bool {
	if v, ok := o.Props[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

// GetPropFloat returns a float64 property or the default value.
func (o *ObjectData) GetPropFloat(key string, def float64) float64 {
	if v, ok := o.Props[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return def
}

// tiledObject represents the JSON structure of a Tiled object.
type tiledObject struct {
	ID         int             `json:"id"`
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	X          float64         `json:"x"`
	Y          float64         `json:"y"`
	Width      float64         `json:"width"`
	Height     float64         `json:"height"`
	Properties []tiledProperty `json:"properties"`
	// TODO: this seems like an anti-pattern. or at least, no one will understand this. we need to fix this properly.
	// Visible is a pointer to detect missing field (Tiled default is visible=true)
	Visible *bool `json:"visible"`
}

// tiledProperty represents a Tiled custom property.
type tiledProperty struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// tiledObjectLayer represents a Tiled object layer.
type tiledObjectLayer struct {
	Name    string        `json:"name"`
	Type    string        `json:"type"`
	Objects []tiledObject `json:"objects"`
}

// ParseObjects extracts objects from raw Tiled JSON data.
// This function parses the JSON and returns all objects from object layers.
func ParseObjects(data []byte) ([]ObjectData, error) {
	var tm struct {
		Layers []json.RawMessage `json:"layers"`
	}

	if err := json.Unmarshal(data, &tm); err != nil {
		return nil, fmt.Errorf("failed to parse Tiled JSON: %w", err)
	}

	var objects []ObjectData

	for _, layerData := range tm.Layers {
		// Try to parse as object layer
		var layer tiledObjectLayer
		if err := json.Unmarshal(layerData, &layer); err != nil {
			continue
		}

		if layer.Type != "objectgroup" {
			continue
		}

		for _, obj := range layer.Objects {
			// Default to visible=true if the property is missing (Tiled's default behavior)
			visible := true
			if obj.Visible != nil && !*obj.Visible {
				visible = false
			}

			// Skip invisible objects
			if !visible && obj.Type != "" {
				continue
			}

			// Parse properties
			props := make(map[string]any)
			for _, prop := range obj.Properties {
				props[prop.Name] = prop.Value
			}

			// Only include objects with a valid type
			if obj.Type == "" {
				continue
			}

			data := ObjectData{
				ID:    obj.ID,
				Name:  obj.Name,
				Type:  ObjectType(obj.Type),
				X:     obj.X,
				Y:     obj.Y,
				W:     obj.Width,
				H:     obj.Height,
				Props: props,
			}
			objects = append(objects, data)
		}
	}

	return objects, nil
}

// FilterObjectsByType returns objects matching the given type.
func FilterObjectsByType(objects []ObjectData, typ ObjectType) []ObjectData {
	var result []ObjectData
	for _, obj := range objects {
		if obj.Type == typ {
			result = append(result, obj)
		}
	}
	return result
}

// FindSpawnPoint returns the first spawn object, or a default position.
func FindSpawnPoint(objects []ObjectData) (x, y float64, found bool) {
	spawns := FilterObjectsByType(objects, ObjectTypeSpawn)
	if len(spawns) == 0 {
		return 0, 0, false
	}
	return spawns[0].X, spawns[0].Y, true
}
