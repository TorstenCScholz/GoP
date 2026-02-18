# AGENTS.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

GoP is a 2D platformer game built with Ebiten (v2) in Go. The project includes both a game runtime and a level editor for designing levels. The game uses Tiled-compatible JSON for level data.

## Build & Run Commands

### Game
```bash
# Run the game directly
go run ./cmd/game

# Build the game binary
go build -o bin/game ./cmd/game
# Or use the Makefile
make build
```

### Level Editor
```bash
# Run the editor
go run ./cmd/editor

# Build the editor binary
go build -o editor ./cmd/editor
```

### Development Commands
```bash
# Run tests (note: currently no test files exist)
go test ./...
# Or use the Makefile
make test

# Format code
gofmt -w .
# Or use the Makefile
make fmt

# Tidy dependencies
go mod tidy
# Or use the Makefile
make tidy

# Run linting
go vet ./...
```

### Utility Commands
```bash
# Generate test spritesheet
go run ./cmd/gensheet

# Generate tileset
go run ./cmd/gentiles

# Verify level object parsing
go run ./cmd/verify_objects
```

## Architecture Overview

The codebase follows a layered architecture with clear separation of concerns:

```
cmd/               - Entry points (game, editor, utilities)
internal/
  app/             - Game loop, scene management, fixed timestep
  scenes/sandbox/  - Main game scene implementation
  entities/        - Game objects (triggers, platforms, hazards, etc.)
  physics/         - Collision detection, resolution, player controller
  world/           - Tilemap rendering, collision maps, object parsing
  camera/          - Advanced camera with deadzone and smoothing
  gfx/             - Sprites, animations, rendering utilities
  gameplay/        - State machine, spawn context, game state
  input/           - Input abstraction and action mapping
  assets/          - Asset loading from embedded filesystem
  editor/          - Level editor implementation (~7300 LOC)
  game/            - Game tuning parameters
  time/            - Fixed timestep implementation
```

### Key Architectural Patterns

**Scene System**: The game uses a scene-based architecture where `App` manages the current `Scene`. Scenes implement `Update()`, `FixedUpdate()`, `Draw()`, and `Layout()` methods.

**Fixed Timestep**: Physics updates run at a fixed rate (60Hz) independent of frame rate using the timestep accumulator pattern.

**Entity Component System**: Entities implement the `Entity` interface with `Update()`, `Draw()`, and `Bounds()` methods. Special interfaces like `Trigger` and `SolidEntity` add specific behaviors.

**Target Registry Pattern**: Instead of direct pointer references between entities (e.g., Switch → Door), the system uses ID-based resolution through `TargetRegistry`. This enables clean serialization and decoupling.

**RenderContext Pattern**: All draw methods receive a `RenderContext` that encapsulates camera, debug flags, screen buffer, and coordinate transformations. This replaced the old pattern of passing raw `camX, camY` coordinates.

**Coordinate Systems**:
- **World Coordinates**: Floating-point pixel positions for entities and physics
- **Tile Coordinates**: Integer grid indices for tilemap data
- **Screen Coordinates**: Final rendering positions after camera transformation

Transformations:
- `WorldToScreen`: `screenX = worldX - camera.X`
- `ScreenToWorld`: `worldX = screenX + camera.X`
- `WorldToTile`: `tileX = int(worldX) / tileWidth`

## Level Editor

The editor is a substantial component (~7300 LOC) that allows visual editing of levels with Tiled JSON compatibility.

### Editor Architecture
- **Ebiten-based**: Uses the same framework as the game for seamless playtest integration
- **Playtest Mode**: Press `P` to test levels in-game without leaving the editor
- **Tools**: Paint, Erase, Fill, Select, Place Object, Move, Resize
- **Layers**: Separate Tiles and Collision layers with visibility toggles
- **Entity System**: Full support for all entity types (spawn, platform, switch, door, hazard, checkpoint, goal)
- **Undo/Redo**: Complete history system for all edit operations
- **Property Editing**: Schema-driven property panel with type validation
- **Validation**: Real-time validation of IDs, references, and level requirements

### Entity Types & Properties
All entity schemas are defined in `internal/editor/schema.go`:
- **spawn**: Player spawn point (no properties)
- **platform**: Moving platforms with `id`, `endX`, `endY`, `speed`, `waitTime`, `pushPlayer`
- **switch**: Switches with `door_id`, `toggle`, `once`
- **door**: Doors with `id`, `startOpen`
- **hazard**: Deadly hazards (no properties)
- **checkpoint**: Save points with `id`
- **goal**: Level completion triggers (no properties)

### Editor Playtest Integration
The editor embeds the game's scene system for instant playtesting:
1. Creates snapshot of editor state
2. Converts editor data to game-compatible format
3. Launches full game scene with physics and entities
4. Press `Escape` to return to editing
5. Press `R` during playtest to restart

## Common Development Patterns

### Adding a New Entity Type
1. Define the entity type constant in `internal/world/objects.go`
2. Add schema to `internal/editor/schema.go` with properties and defaults
3. Create entity struct in `internal/entities/` implementing `Entity` interface
4. Add spawning logic to scene's entity creation
5. Update object parsing if new property types are needed

### Creating a New Scene
1. Create package under `internal/scenes/`
2. Implement the `Scene` interface from `internal/app/app.go`
3. Optionally implement `SceneDebugger` for debug rendering
4. Initialize scene in `cmd/game/main.go` or via scene transitions

### Working with Tilemaps
- Levels are stored as Tiled JSON in `assets/levels/`
- Use `world.ParseTiledJSON()` to load map structure
- Use `world.ParseObjects()` to extract entity placements
- Collision detection uses `CollisionMap` which wraps a boolean `SolidGrid`

### Physics Integration
- Create a `physics.Body` for movable entities
- Use `physics.Controller` for player input-driven movement
- Use `physics.CollisionResolver` for axis-separated collision resolution
- Collision checks happen in fixed timestep updates

## Known Issues

**Build Error**: There is currently a compilation error in `internal/entities/platform.go:236-237`:
```
conversion from int to string yields a string of one rune, not a string of digits
```
This is in the `GetDebugInfo()` method where integers are being converted to strings incorrectly. Use `strconv.Itoa()` or `fmt.Sprintf()` instead of direct `string(int)` conversion.

## Asset Pipeline

Assets are embedded in the binary using Go's `embed` directive:
- `assets/tiles/tiles.png` - 128x128 tileset (8x8 tiles of 16x16px each)
- `assets/sprites/test_sheet.png` - Animation spritesheet
- `assets/levels/level_01.json` - Level data in Tiled JSON format

Access via `internal/assets/assets.go` which provides `FS()` for embedded filesystem.

## Testing

Currently, the project has no test files. When adding tests:
- Place test files alongside the code they test (`*_test.go`)
- Use standard Go testing package
- Run with `go test ./...` or `make test`

## Design Documents

The `docs/` directory contains detailed design documents:
- `architecture-overview.md` - Comprehensive system architecture
- `level-editor-design.md` - Complete editor design and implementation plan
- `tilemap-collision-design.md` - Collision system design
- `moving-platforms-design.md` - Platform movement mechanics
- `asset-rendering-design.md` - Rendering pipeline
- `feel-camera-design.md` - Camera behavior
- `interactive-world-design.md` - Entity interaction system

Refer to these documents for deep dives into specific subsystems.

## Development Workflow

1. Make code changes in `internal/` packages
2. Test changes by running the game: `go run ./cmd/game`
3. For level editing, use the editor: `go run ./cmd/editor`
4. Format code: `make fmt` or `gofmt -w .`
5. Verify builds: `go vet ./...`
6. Build binaries: `make build` or `go build ./cmd/game`

The CI pipeline runs `go vet` and `go test` on push/PR.
