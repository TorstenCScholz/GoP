# Moving Platforms Design Document

## Overview

This document describes the architecture for adding moving platforms to the game. Moving platforms are solid entities that move along a defined path (A↔B), collide with the player, carry the player when standing on them, and optionally push the player sideways.

## Analysis of Existing Codebase

### Entity System

The entity system is defined in [`internal/entities/entity.go`](internal/entities/entity.go) with these key interfaces:

- **`Entity`**: Base interface with `Update(dt)`, `Draw()`, `DrawWithContext()`, and `Bounds()`
- **`SolidEntity`**: Entity with physics collision - extends Entity with `Body() *physics.Body`
- **`Trigger`**: Entity that responds to player overlap with `OnEnter()` and `OnExit()`

The [`EntityWorld`](internal/entities/world.go:24) manages three separate lists:
- `entities []Entity` - all entities
- `triggers []Trigger` - trigger entities for overlap checking
- `solidEnts []SolidEntity` - solid entities for collision

### Physics System

The physics system in [`internal/physics/`](internal/physics/) consists of:

- **`Body`** ([`body.go`](internal/physics/body.go)): Position, velocity, size, and ground state
- **`Controller`** ([`controller.go`](internal/physics/controller.go)): Player input handling with feel mechanics
- **`CollisionResolver`** ([`resolve.go`](internal/physics/resolve.go)): Axis-separated collision resolution

Key observation: The collision resolution uses axis-separated resolution (X then Y), which is important for platform carrying logic.

### Collision Detection

From [`internal/world/collision.go`](internal/world/collision.go) and [`internal/scenes/sandbox/scene.go`](internal/scenes/sandbox/scene.go):

- Tile collision uses `CollisionMap.IsSolidAtTile()` and `OverlapsSolid()`
- Entity collision is checked in the scene's `resolveCollisions()` method (lines 326-369)
- Solid entities are checked after tile collision

### Tiled Object Parsing

From [`internal/world/objects.go`](internal/world/objects.go):

- `ObjectData` struct holds parsed Tiled object properties
- Property accessors: `GetPropString()`, `GetPropInt()`, `GetPropBool()`, `GetPropFloat()`
- Object types are defined as constants (e.g., `ObjectTypeDoor`, `ObjectTypeSwitch`)

### Existing Entity Patterns

**Door** ([`internal/entities/door.go`](internal/entities/door.go)):
- Implements `SolidEntity` and `Targetable`
- Modifies `body.W` and `body.H` to enable/disable collision
- Shows pattern for solid entities with dynamic state

**Switch** ([`internal/entities/switch.go`](internal/entities/switch.go)):
- Implements `Trigger` interface
- Uses `TriggerState` for active/triggered tracking
- Links to targets via `TargetRegistry`

### Spawner Pattern

From [`internal/gameplay/spawner.go`](internal/gameplay/spawner.go):

- `SpawnContext` provides callbacks for entity spawning
- `SpawnEntities()` creates entities from `ObjectData`
- Two-pass spawning: create entities first, then link references

---

## Proposed Architecture

### Kinematic Interface

A new interface for moving solid entities that can carry the player:

```go
// Kinematic represents a solid entity that moves and can carry the player.
// Moving platforms, elevators, and similar entities implement this interface.
type Kinematic interface {
    SolidEntity

    // Velocity returns the current movement velocity in pixels per second.
    Velocity() (vx, vy float64)

    // MoveAndSlide moves the platform and returns the actual displacement.
    // The platform should handle its own collision with the world.
    MoveAndSlide(collisionMap *world.CollisionMap, dt float64) (dx, dy float64)
}
```

**Rationale**: 
- Extends `SolidEntity` to inherit collision behavior
- `Velocity()` exposes movement for player carry logic
- `MoveAndSlide()` allows platforms to handle their own collision

### MovingPlatform Struct

```go
// MovingPlatform is a solid entity that moves between two points.
type MovingPlatform struct {
    body      *physics.Body
    velocity  physics.Velocity  // Current velocity

    // Path configuration
    startX, startY float64  // Point A
    endX, endY     float64  // Point B
    speed          float64  // Movement speed in pixels/second

    // State
    movingToEnd bool  // true = moving to B, false = moving to A
    paused      bool
    pauseTimer  float64

    // Configuration
    waitTime    float64  // Time to wait at endpoints (seconds)
    pushPlayer  bool     // Whether to push player sideways
}
```

**Key Design Decisions**:
1. Uses `physics.Body` for consistency with existing entities
2. Simple A↔B path (no complex paths for v1)
3. Configurable wait time at endpoints
4. Optional player pushing

### Integration with Physics System

#### Platform Update Order

The platform must update **before** player physics:

```
1. Update platforms (move them)
2. Update player physics
3. Resolve player collision with tiles AND platforms
4. Apply platform velocity to player if standing on it
```

#### Player Carry Logic

When the player is standing on a platform, the platform's velocity should be added to the player's position:

```go
// In player physics update, after collision resolution:
for _, platform := range kinematics {
    if playerStandingOnPlatform(player, platform) {
        // Apply platform velocity to player
        pvx, pvy := platform.Velocity()
        player.PosX += pvx * dt
        // Y velocity already handled by collision
    }
}
```

#### Detecting "Standing On Platform"

A player is standing on a platform when:
1. Player's bottom edge aligns with platform's top edge (within tolerance)
2. Player's horizontal bounds overlap with platform's horizontal bounds
3. Player was moving down (dy > 0) in the last frame

```go
func playerStandingOnPlatform(player *physics.Body, platform Kinematic) bool {
    platformBounds := platform.Bounds()
    
    // Check vertical alignment (player bottom at platform top)
    tolerance := 2.0  // pixels
    playerBottom := player.PosY + player.H
    platformTop := platformBounds.Y
    if math.Abs(playerBottom - platformTop) > tolerance {
        return false
    }
    
    // Check horizontal overlap
    playerRight := player.PosX + player.W
    platformRight := platformBounds.X + platformBounds.W
    if player.PosX >= platformRight || playerRight <= platformBounds.X {
        return false
    }
    
    return true
}
```

### EntityWorld Integration

Add a new list for kinematic entities:

```go
type EntityWorld struct {
    entities  []Entity
    triggers  []Trigger
    solidEnts []SolidEntity
    kinematics []Kinematic  // NEW: Moving solid entities
    // ... rest unchanged
}

func (w *EntityWorld) AddKinematic(k Kinematic) {
    w.kinematics = append(w.kinematics, k)
    w.solidEnts = append(w.solidEnts, k)  // Also add to solid entities
    w.entities = append(w.entities, k)     // And general entities
}

func (w *EntityWorld) Kinematics() []Kinematic {
    return w.kinematics
}
```

### Tiled Object Properties

New object type: `ObjectTypePlatform = "platform"`

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `endX` | float | 0 | X offset from start position to endpoint B |
| `endY` | float | 0 | Y offset from start position to endpoint B |
| `speed` | float | 60 | Movement speed in pixels/second |
| `waitTime` | float | 0.5 | Time to wait at endpoints (seconds) |
| `pushPlayer` | bool | false | Whether to push player sideways |
| `startOpen` | bool | false | Start at endpoint B instead of A |

**Tiled Object Example**:
```
Type: platform
X: 100  (start position A)
Y: 200
Width: 64
Height: 16
Properties:
  - endX: 200     (endpoint B is at 300, 200)
  - endY: 0
  - speed: 80
  - waitTime: 1.0
  - pushPlayer: true
```

---

## Implementation Plan

### Phase 1: Core Infrastructure

1. **Add `Kinematic` interface** to [`internal/entities/entity.go`](internal/entities/entity.go)
   - Define interface with `Velocity()` and `MoveAndSlide()`

2. **Update `EntityWorld`** in [`internal/entities/world.go`](internal/entities/world.go)
   - Add `kinematics []Kinematic` field
   - Add `AddKinematic()` and `Kinematics()` methods

3. **Add `Velocity` type** to [`internal/physics/body.go`](internal/physics/body.go) (optional, or just use vx, vy fields)

### Phase 2: MovingPlatform Entity

4. **Create `internal/entities/platform.go`**
   - Define `MovingPlatform` struct
   - Implement `Kinematic` interface
   - Implement movement logic with pause at endpoints
   - Implement `MoveAndSlide()` for platform collision with tiles

5. **Add platform object type** to [`internal/world/objects.go`](internal/world/objects.go)
   - Add `ObjectTypePlatform` constant

### Phase 3: Physics Integration

6. **Update collision resolution** in [`internal/scenes/sandbox/scene.go`](internal/scenes/sandbox/scene.go)
   - Add platform collision to `resolveCollisions()`
   - Track which platform player is standing on

7. **Add carry logic** to player physics update
   - Apply platform velocity to player when standing on platform
   - Handle horizontal pushing if enabled

### Phase 4: Spawning

8. **Update spawner** in [`internal/gameplay/spawner.go`](internal/gameplay/spawner.go)
   - Add case for `ObjectTypePlatform`
   - Parse platform properties from Tiled object

### Phase 5: Testing

9. **Create test level** with moving platforms
   - Horizontal platform
   - Vertical platform (elevator)
   - Platform with player pushing

---

## Detailed Implementation Notes

### Platform Movement Logic

```go
func (p *MovingPlatform) Update(dt float64) {
    if p.paused {
        p.pauseTimer -= dt
        if p.pauseTimer <= 0 {
            p.paused = false
        }
        p.velocity.vx = 0
        p.velocity.vy = 0
        return
    }

    // Calculate direction to target
    targetX, targetY := p.endX, p.endY
    if !p.movingToEnd {
        targetX, targetY = p.startX, p.startY
    }

    dx := targetX - p.body.PosX
    dy := targetY - p.body.PosY
    dist := math.Sqrt(dx*dx + dy*dy)

    if dist < p.speed * dt {
        // Reached endpoint
        p.body.PosX = targetX
        p.body.PosY = targetY
        p.movingToEnd = !p.movingToEnd
        p.paused = true
        p.pauseTimer = p.waitTime
        p.velocity.vx = 0
        p.velocity.vy = 0
    } else {
        // Move towards target
        p.velocity.vx = (dx / dist) * p.speed
        p.velocity.vy = (dy / dist) * p.speed
        p.body.PosX += p.velocity.vx * dt
        p.body.PosY += p.velocity.vy * dt
    }
}
```

### Platform Collision with Tiles

Platforms should collide with solid tiles to prevent passing through walls:

```go
func (p *MovingPlatform) MoveAndSlide(collisionMap *world.CollisionMap, dt float64) (dx, dy float64) {
    // Try X movement
    p.body.PosX += p.velocity.vx * dt
    if collisionMap.OverlapsSolid(p.body.PosX, p.body.PosY, p.body.W, p.body.H) {
        // Revert and stop
        p.body.PosX -= p.velocity.vx * dt
        p.velocity.vx = 0
    }

    // Try Y movement
    p.body.PosY += p.velocity.vy * dt
    if collisionMap.OverlapsSolid(p.body.PosX, p.body.PosY, p.body.W, p.body.H) {
        p.body.PosY -= p.velocity.vy * dt
        p.velocity.vy = 0
    }

    return p.velocity.vx * dt, p.velocity.vy * dt
}
```

### Player Carry Logic Integration

In the scene's `FixedUpdate()`:

```go
func (s *Scene) FixedUpdate() error {
    // ... existing checks ...

    // 1. Update platforms first
    for _, k := range s.entityWorld.Kinematics() {
        k.MoveAndSlide(s.collisionMap, s.timestep.TickDuration().Seconds())
        k.Update(s.timestep.TickDuration().Seconds())
    }

    // 2. Track platform player was standing on before physics
    var standingPlatform Kinematic
    for _, k := range s.entityWorld.Kinematics() {
        if playerStandingOnPlatform(s.playerBody, k) {
            standingPlatform = k
            break
        }
    }

    // 3. Update player physics (existing code)
    s.playerController.FixedUpdate(...)

    // 4. Apply platform velocity if standing on one
    if standingPlatform != nil {
        pvx, _ := standingPlatform.Velocity()
        s.playerBody.PosX += pvx * s.timestep.TickDuration().Seconds()
    }

    return nil
}
```

---

## Future Considerations (Out of Scope for v1)

These features are documented for future reference but not part of the initial implementation:

1. **One-way platforms**: Player can jump through from below
2. **Sloped platforms**: Platforms at angles
3. **Crush detection**: Detect when player is crushed between platform and wall
4. **Complex paths**: Waypoint-based paths instead of simple A↔B
5. **Platform activation**: Switches that start/stop platforms
6. **Falling platforms**: Platforms that fall when stood on

---

## Summary

The MovingPlatform architecture integrates cleanly with the existing entity and physics systems by:

1. Adding a `Kinematic` interface that extends `SolidEntity`
2. Managing platforms in a separate list in `EntityWorld` for efficient iteration
3. Updating platforms before player physics to enable carry logic
4. Using the existing collision system for platform-tile collision
5. Following established patterns for Tiled object parsing and spawning

The design is minimal and focused on v1 requirements while providing clear extension points for future features.
