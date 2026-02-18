// Package editor provides the level editor functionality.
package editor

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/torsten/GoP/internal/world"
)

// TiledJSON represents the full Tiled JSON format for serialization.
// This ensures 100% read/write compatibility with the existing level format.
type TiledJSON struct {
	CompressionLevel int            `json:"compressionlevel"`
	Height           int            `json:"height"`
	Infinite         bool           `json:"infinite"`
	Layers           []TiledLayer   `json:"layers"`
	NextLayerID      int            `json:"nextlayerid"`
	NextObjectID     int            `json:"nextobjectid"`
	Orientation      string         `json:"orientation"`
	RenderOrder      string         `json:"renderorder"`
	TiledVersion     string         `json:"tiledversion"`
	TileHeight       int            `json:"tileheight"`
	Tilesets         []TiledTileset `json:"tilesets"`
	TileWidth        int            `json:"tilewidth"`
	Type             string         `json:"type"`
	Version          string         `json:"version"`
	Width            int            `json:"width"`
}

// TiledLayer represents a layer in the Tiled JSON format.
type TiledLayer struct {
	Data       []int           `json:"data,omitempty"` // For tile layers
	Height     int             `json:"height"`
	ID         int             `json:"id"`
	Name       string          `json:"name"`
	Opacity    float64         `json:"opacity"`
	Type       string          `json:"type"` // "tilelayer" or "objectgroup"
	Visible    bool            `json:"visible"`
	Width      int             `json:"width"`
	X          int             `json:"x"`
	Y          int             `json:"y"`
	Objects    []TiledObject   `json:"objects,omitempty"`    // For object layers
	Properties []TiledProperty `json:"properties,omitempty"` // For object layers
}

// TiledObject represents an object in the Tiled JSON format.
type TiledObject struct {
	Height     float64         `json:"height"`
	ID         int             `json:"id"`
	Name       string          `json:"name"`
	Properties []TiledProperty `json:"properties,omitempty"`
	Type       string          `json:"type"`
	Visible    bool            `json:"visible,omitempty"`
	Width      float64         `json:"width"`
	X          float64         `json:"x"`
	Y          float64         `json:"y"`
}

// TiledProperty represents a custom property in the Tiled JSON format.
type TiledProperty struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// TiledTileset represents a tileset reference in the Tiled JSON format.
type TiledTileset struct {
	Columns     int    `json:"columns"`
	FirstGID    int    `json:"firstgid"`
	Image       string `json:"image"`
	ImageHeight int    `json:"imageheight"`
	ImageWidth  int    `json:"imagewidth"`
	Margin      int    `json:"margin"`
	Name        string `json:"name"`
	Spacing     int    `json:"spacing"`
	TileCount   int    `json:"tilecount"`
	TileHeight  int    `json:"tileheight"`
	TileWidth   int    `json:"tilewidth"`
}

// DefaultLevelPath is the hardcoded path for OpenLevel (will be replaced with dialogs later).
const DefaultLevelPath = "assets/levels/level_01.json"

// DefaultTilesetPath is the path to the default tileset.
const DefaultTilesetPath = "../tiles/tiles.png"

// Default level dimensions
const (
	DefaultLevelWidth  = 80
	DefaultLevelHeight = 25
	DefaultTileSize    = 16
)

// NewLevel creates a new empty level with the specified dimensions.
// The level is initialized with empty Tiles and Collision layers.
func NewLevel(width, height int) *EditorState {
	// Create empty tile data
	tileData := make([]int, width*height)
	collisionData := make([]int, width*height)

	// Create the TiledJSON structure
	tiledJSON := &TiledJSON{
		CompressionLevel: -1,
		Height:           height,
		Infinite:         false,
		Layers: []TiledLayer{
			{
				Data:    tileData,
				Height:  height,
				ID:      1,
				Name:    "Tiles",
				Opacity: 1.0,
				Type:    "tilelayer",
				Visible: true,
				Width:   width,
				X:       0,
				Y:       0,
			},
			{
				Data:    collisionData,
				Height:  height,
				ID:      2,
				Name:    "Collision",
				Opacity: 1.0,
				Type:    "tilelayer",
				Visible: true,
				Width:   width,
				X:       0,
				Y:       0,
			},
			{
				Height:  0,
				ID:      3,
				Name:    "Objects",
				Opacity: 1.0,
				Type:    "objectgroup",
				Visible: true,
				Width:   0,
				X:       0,
				Y:       0,
			},
		},
		NextLayerID:  4,
		NextObjectID: 1,
		Orientation:  "orthogonal",
		RenderOrder:  "right-down",
		TiledVersion: "1.10.2",
		TileHeight:   DefaultTileSize,
		Tilesets: []TiledTileset{
			{
				Columns:     8,
				FirstGID:    1,
				Image:       DefaultTilesetPath,
				ImageHeight: 128,
				ImageWidth:  128,
				Margin:      0,
				Name:        "tiles",
				Spacing:     0,
				TileCount:   64,
				TileHeight:  DefaultTileSize,
				TileWidth:   DefaultTileSize,
			},
		},
		TileWidth: DefaultTileSize,
		Type:      "map",
		Version:   "1.10",
		Width:     width,
	}

	// Convert to EditorState
	state, err := tiledJSONToEditorState(tiledJSON, "")
	if err != nil {
		// This should never happen for a newly created level
		panic(fmt.Sprintf("failed to create new level: %v", err))
	}

	return state
}

// OpenLevel loads an existing Tiled JSON file and returns the editor state.
func OpenLevel(path string) (*EditorState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read level file: %w", err)
	}

	return ParseLevel(data, path)
}

// ParseLevel parses Tiled JSON data and returns the editor state.
func ParseLevel(data []byte, path string) (*EditorState, error) {
	// Parse the full Tiled JSON structure
	var tiledJSON TiledJSON
	if err := json.Unmarshal(data, &tiledJSON); err != nil {
		return nil, fmt.Errorf("failed to parse Tiled JSON: %w", err)
	}

	return tiledJSONToEditorState(&tiledJSON, path)
}

// tiledJSONToEditorState converts a TiledJSON structure to EditorState.
func tiledJSONToEditorState(tiledJSON *TiledJSON, path string) (*EditorState, error) {
	// Create MapData from the Tiled JSON
	mapData := &world.MapData{}
	// We need to use internal access - let's create a workaround
	// by re-parsing with the world package functions

	// Serialize back to JSON and use world.ParseTiledJSON
	jsonData, err := json.Marshal(tiledJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to re-serialize Tiled JSON: %w", err)
	}

	// Parse using the world package functions
	mapData, err = world.ParseTiledJSON(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse map data: %w", err)
	}

	// Parse objects
	objects, err := world.ParseObjects(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse objects: %w", err)
	}

	// Create editor state
	state := NewEditorState()
	state.FilePath = path
	state.MapData = mapData
	state.Objects = objects

	return state, nil
}

// SaveLevel saves the editor state to the current file path.
// Returns an error if no file path is set.
func SaveLevel(state *EditorState) error {
	if state.FilePath == "" {
		return fmt.Errorf("no file path set, use SaveLevelAs instead")
	}
	return SaveLevelAs(state, state.FilePath)
}

// SaveLevelAs saves the editor state to a new file path.
func SaveLevelAs(state *EditorState, path string) error {
	if state.MapData == nil {
		return fmt.Errorf("no level data to save")
	}

	// Convert EditorState to TiledJSON
	tiledJSON, err := editorStateToTiledJSON(state)
	if err != nil {
		return fmt.Errorf("failed to convert level data: %w", err)
	}

	// Serialize to JSON
	jsonData, err := json.MarshalIndent(tiledJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize level: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write level file: %w", err)
	}

	// Update state with new path and clear modified flag
	state.FilePath = path
	state.SetModified(false)

	return nil
}

// editorStateToTiledJSON converts EditorState to TiledJSON for serialization.
func editorStateToTiledJSON(state *EditorState) (*TiledJSON, error) {
	if state.MapData == nil {
		return nil, fmt.Errorf("no map data")
	}

	// Build layers
	layers := make([]TiledLayer, 0)

	// Track layer IDs
	layerID := 1
	nextObjectID := 1

	// Add tile layers
	for _, layer := range state.MapData.Layers() {
		tiledLayer := TiledLayer{
			Data:    layer.Data(),
			Height:  layer.Height(),
			ID:      layerID,
			Name:    layer.Name(),
			Opacity: 1.0,
			Type:    "tilelayer",
			Visible: true,
			Width:   layer.Width(),
			X:       0,
			Y:       0,
		}
		layers = append(layers, tiledLayer)
		layerID++
	}

	// Build objects from state
	objects := make([]TiledObject, 0)
	for _, obj := range state.Objects {
		// Track the highest object ID
		if obj.ID >= nextObjectID {
			nextObjectID = obj.ID + 1
		}

		// Convert properties
		props := make([]TiledProperty, 0)
		for key, value := range obj.Props {
			prop := TiledProperty{
				Name:  key,
				Value: value,
			}
			// Determine type based on value
			switch v := value.(type) {
			case string:
				prop.Type = "string"
			case float64:
				prop.Type = "float"
			case int:
				prop.Type = "int"
				prop.Value = float64(v)
			case bool:
				prop.Type = "bool"
			default:
				prop.Type = "string"
			}
			props = append(props, prop)
		}

		tiledObj := TiledObject{
			Height:     obj.H,
			ID:         obj.ID,
			Name:       obj.Name,
			Properties: props,
			Type:       string(obj.Type),
			Visible:    true,
			Width:      obj.W,
			X:          obj.X,
			Y:          obj.Y,
		}
		objects = append(objects, tiledObj)
	}

	// Add object layer
	objectLayer := TiledLayer{
		Height:  0,
		ID:      layerID,
		Name:    "Objects",
		Opacity: 1.0,
		Type:    "objectgroup",
		Visible: true,
		Width:   0,
		X:       0,
		Y:       0,
		Objects: objects,
	}
	layers = append(layers, objectLayer)

	// Build the full Tiled JSON
	tiledJSON := &TiledJSON{
		CompressionLevel: -1,
		Height:           state.MapData.Height(),
		Infinite:         false,
		Layers:           layers,
		NextLayerID:      layerID + 1,
		NextObjectID:     nextObjectID,
		Orientation:      "orthogonal",
		RenderOrder:      "right-down",
		TiledVersion:     "1.10.2",
		TileHeight:       state.MapData.TileHeight(),
		Tilesets: []TiledTileset{
			{
				Columns:     8,
				FirstGID:    1,
				Image:       DefaultTilesetPath,
				ImageHeight: 128,
				ImageWidth:  128,
				Margin:      0,
				Name:        "tiles",
				Spacing:     0,
				TileCount:   64,
				TileHeight:  DefaultTileSize,
				TileWidth:   DefaultTileSize,
			},
		},
		TileWidth: state.MapData.TileWidth(),
		Type:      "map",
		Version:   "1.10",
		Width:     state.MapData.Width(),
	}

	return tiledJSON, nil
}
