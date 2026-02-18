# GoP

GoP is a 2D platformer built in Go with [Ebiten v2](https://ebitengine.org/). The repository includes both:

- a playable game runtime
- a full level editor with in-editor playtesting

Levels are stored in Tiled-compatible JSON and loaded from embedded assets.

## Requirements

- Go `1.24+`
- A desktop environment that supports Ebiten

## Quick Start

### Run the game

```bash
go run ./cmd/game
```

### Run the level editor

```bash
go run ./cmd/editor
```

## Build

```bash
# Game
make build
# or
go build -o bin/game ./cmd/game

# Editor
make build-editor
# or
go build -o bin/editor ./cmd/editor
```

## Common Development Commands

```bash
# Tests
go test ./...
# or
make test

# Format
gofmt -w .
# or
make fmt

# Dependency cleanup
go mod tidy
# or
make tidy

# Static checks
go vet ./...
```

## Utility Tools

```bash
# Generate test sprite sheet
go run ./cmd/gensheet

# Generate tileset
go run ./cmd/gentiles

# Verify level object parsing
go run ./cmd/verify_objects
```

## Project Structure

```text
cmd/               # Entrypoints (game, editor, tooling)
internal/
  app/             # App loop and scene lifecycle
  scenes/sandbox/  # Main game scene
  entities/        # Gameplay entities (platforms, switches, doors, etc.)
  physics/         # Collision and movement logic
  world/           # Tilemap loading/rendering and object parsing
  camera/          # Camera behavior (deadzone/smoothing)
  gfx/             # Sprite and animation helpers
  gameplay/        # Spawn/state orchestration
  input/           # Input abstractions
  assets/          # Embedded asset access
  editor/          # Level editor implementation
  game/            # Game tuning parameters
  time/            # Fixed timestep utilities
assets/            # Source art and level JSON
docs/              # Design and architecture docs
```

## Editor Notes

The editor supports painting/erasing/fill/select/object placement, undo/redo, validation, and playtest mode.

- Press `P` in the editor to enter playtest.
- Press `Escape` to return from playtest to editing.
- Press `R` during playtest to restart.

## Architecture Highlights

- **Scene-based runtime**: scenes implement update/fixed-update/draw/layout lifecycle.
- **Fixed timestep physics**: simulation runs at a stable update rate independent of rendering.
- **ID-based entity linking**: interactions (for example switch-to-door) are resolved by IDs via a registry.
- **RenderContext drawing model**: rendering passes shared camera/debug/screen context instead of raw offsets.

## Assets and Levels

Assets are embedded in the binary via Go `embed` and loaded through `internal/assets`.

- `assets/tiles/tiles.png`
- `assets/sprites/test_sheet.png`
- `assets/levels/level_01.json`

## Documentation

See `docs/` for detailed subsystem designs:

- `docs/architecture-overview.md`
- `docs/level-editor-design.md`
- `docs/tilemap-collision-design.md`
- `docs/moving-platforms-design.md`
- `docs/asset-rendering-design.md`
- `docs/feel-camera-design.md`
- `docs/interactive-world-design.md`

## CI

CI runs:

- `go vet ./...`
- `go test ./...`
