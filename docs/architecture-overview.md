# GoP Architecture Overview

This document provides a comprehensive overview of the GoP (Go Platformer) codebase architecture, including core classes, their relationships, coordinate systems, abstraction layers, and event processing.

## Table of Contents

1. [Overview](#overview)
2. [Class Hierarchy and Relationships](#class-hierarchy-and-relationships)
3. [Abstraction Layers](#abstraction-layers)
4. [Coordinate Systems and Transformations](#coordinate-systems-and-transformations)
5. [Event Processing System](#event-processing-system)
6. [Data Flow Diagrams](#data-flow-diagrams)

---

## Overview

GoP is a 2D platformer game built with [Ebiten](https://ebiten.org/) in Go. The architecture follows a layered design with clear separation of concerns:

- **Application Layer**: Game loop, scene management, fixed timestep physics
- **Scene Layer**: Coordinates all subsystems for a specific game state
- **Entity Layer**: Game objects with update, draw, and collision behaviors
- **Physics Layer**: Collision detection, resolution, and player controller
- **World Layer**: Tilemap rendering, collision data, camera system
- **Graphics Layer**: Sprites, animations, and rendering utilities

```
+------------------+
|     cmd/         |  Entry points (main packages)
+------------------+
         |
         v
+------------------+
|   internal/app/  |  Application core (game loop, scene management)
+------------------+
         |
         v
+------------------+
| internal/scenes/ |  Scene implementations
+------------------+
         |
    +----+----+----+----+----+
    |    |    |    |    |    |
    v    v    v    v    v    v
  entities physics world camera gfx gameplay
```

---

## Class Hierarchy and Relationships

### Application Layer

Located in [`internal/app/`](internal/app/).

```
+---------------------------+
|          App              |
+---------------------------+
| - scene: Scene            |
| - input: *input.Input     |
| - config: *Config         |
| - debugActive: bool       |
| - timestep: *timestep.Timestep |
| - lastUpdate: time.Time   |
+---------------------------+
| + New(cfg *Config) *App   |
| + SetScene(scene Scene)   |
| + Update() error          |
| + Draw(screen *Image)     |
| + Layout(w,h int) (int,int)|
| + Run() error             |
+---------------------------+
             |
             | uses
             v
+---------------------------+
|     Scene interface       |
+---------------------------+
| + Update(inp *Input) error|
| + FixedUpdate() error     |
| + Draw(screen *Image)     |
| + Layout(w,h int)(int,int)|
| + DebugInfo() string      |
+---------------------------+
             ^
             |
+---------------------------+
|  SceneDebugger interface  |  (optional)
+---------------------------+
| + DrawDebug(screen *Image)|
+---------------------------+
```

**Key Files:**
- [`App struct`](internal/app/app.go:35) - Main application, implements `ebiten.Game`
- [`Scene interface`](internal/app/app.go:14) - Contract for game scenes
- [`SceneDebugger interface`](internal/app/app.go:29) - Optional debug rendering
- [`Config struct`](internal/app/config.go) - Window configuration

### Entity System

Located in [`internal/entities/`](internal/entities/).

```
+---------------------------+
|     Entity interface      |
+---------------------------+
| + Update(dt float64)      |
| + Draw(screen, camX, camY)|
| + Bounds() physics.AABB   |
+---------------------------+
             ^
             |
    +--------+--------+
    |                 |
+-----------+  +---------------+
|  Trigger  |  |  SolidEntity  |
+-----------+  +---------------+
| Entity    |  | Entity        |
+-----------+  +---------------+
| + OnEnter(player)        |
| + OnExit(player)         |
| + IsActive() bool        |
| + WasTriggered() bool    |
| + SetTriggered(bool)     |
+---------------------------+
             ^
             |
+---------------------------+
|      TriggerState         |
+---------------------------+
| Active: bool              |
| Triggered: bool           |
+---------------------------+
```

**Entity Implementations:**

```
+------------------+
|    Checkpoint    |
+------------------+
| bounds: AABB     |
| id: string       |
| state: TriggerState |
| triggered: bool  |
| OnActivate: func |
+------------------+

+------------------+
|      Door        |
+------------------+
| body: *Body      |
| id: string       |
| isOpen: bool     |
| closedW/H: float64 |
+------------------+

+------------------+
|      Goal        |
+------------------+
| bounds: AABB     |
| state: TriggerState |
| OnComplete: func |
+------------------+

+------------------+
|     Hazard       |
+------------------+
| bounds: AABB     |
| state: TriggerState |
| OnDeath: func    |
+------------------+

+------------------+
|     Switch       |
+------------------+
| bounds: AABB     |
| state: TriggerState |
| targetID: string |
| toggleMode: bool |
| once: bool       |
| used: bool       |
| targetDoor: *Door |
+------------------+
```

**Key Files:**
- [`Entity interface`](internal/entities/entity.go:11) - Base entity contract
- [`Trigger interface`](internal/entities/entity.go:24) - Overlap-based triggers
- [`SolidEntity interface`](internal/entities/entity.go:45) - Entities with physics bodies
- [`TriggerState struct`](internal/entities/entity.go:53) - Shared trigger state
- [`EntityWorld struct`](internal/entities/world.go:23) - Entity container and manager
- [`Checkpoint`](internal/entities/checkpoint.go), [`Door`](internal/entities/door.go), [`Goal`](internal/entities/goal.go), [`Hazard`](internal/entities/hazard.go), [`Switch`](internal/entities/switch.go)

### World/Map System

Located in [`internal/world/`](internal/world/).

```
+------------------+
|     Tileset      |
+------------------+
| image: *Image    |
| tileWidth: int   |
| tileHeight: int  |
| tiles: []*Image  |
| columns: int     |
+------------------+

+------------------+
|    TileLayer     |
+------------------+
| name: string     |
| width: int       |
| height: int      |
| data: []int      |
+------------------+
| + TileAt(x,y) int |
+------------------+

+------------------+
|      Map         |
+------------------+
| width: int       |
| height: int      |
| tileWidth: int   |
| tileHeight: int  |
| layers: []*TileLayer |
| tileset: *Tileset |
| layerIndex: map  |
+------------------+
| + Layer(name) *TileLayer |
| + Width(), Height() int  |
| + PixelWidth/Height() int |
+------------------+

+------------------+
|    SolidGrid     |
+------------------+
| width: int       |
| height: int      |
| data: []bool     |
+------------------+
| + IsSolid(tx,ty) bool |
| + SetSolid(tx,ty,bool)|
+------------------+

+------------------+
|   CollisionMap   |
+------------------+
| grid: *SolidGrid |
| tileW: int       |
| tileH: int       |
+------------------+
| + IsSolidAtTile() |
| + IsSolidAtWorld()|
| + OverlapsSolid() |
+------------------+
```

**Key Files:**
- [`Tileset`](internal/world/map.go) - Tile image collection
- [`TileLayer`](internal/world/map.go) - 2D tile ID storage
- [`Map`](internal/world/map.go) - Complete tilemap with layers
- [`SolidGrid`](internal/world/collision.go:9) - Boolean collision grid
- [`CollisionMap`](internal/world/collision.go:74) - Collision query interface
- [`Camera`](internal/world/render.go:8) - Viewport offset for scrolling
- [`MapRenderer`](internal/world/render.go:67) - Tilemap rendering with camera

### Physics System

Located in [`internal/physics/`](internal/physics/).

```
+------------------+
|       AABB       |
+------------------+
| X: float64       |  Top-left position
| Y: float64       |
| W: float64       |  Width
| H: float64       |  Height
+------------------+
| + Left() float64 |
| + Right() float64|
| + Top() float64  |
| + Bottom() float64|
| + Intersects(AABB) bool |
+------------------+

+------------------+
|       Body       |
+------------------+
| PosX, PosY: float64 |
| VelX, VelY: float64 |
| W, H: float64    |
| OnGround: bool   |
+------------------+
| + AABB() AABB    |
+------------------+

+------------------+
|    Collision     |
+------------------+
| TileX: int       |
| TileY: int       |
| NormalX: float64 |
| NormalY: float64 |
+------------------+

+------------------+
|   PlayerState    |
+------------------+
| TimeSinceGrounded: Duration |
| JumpBufferTime: Duration    |
| JumpHeldTime: Duration      |
| WasGrounded: bool           |
| JumpBuffered: bool          |
| IsJumping: bool             |
| JumpReleased: bool          |
+------------------+

+------------------+
|    Controller    |
+------------------+
| Body: *Body      |
| Tuning: game.Tuning |
| State: PlayerState |
+------------------+
| + Update(dt, input, collisionMap) |
| + ApplyGravity(dt)  |
| + HandleJump(input) |
| + ResolveCollisions()|
+------------------+

+----------------------+
| CollisionResolver    |
+----------------------+
| TileW: int           |
| TileH: int           |
+----------------------+
| + Resolve(body, collisionMap) |
| + resolveAxis()      |
+----------------------+
```

**Key Files:**
- [`AABB struct`](internal/physics/aabb.go:5) - Axis-aligned bounding box
- [`Body struct`](internal/physics/body.go:5) - Physics body with position/velocity
- [`Controller struct`](internal/physics/controller.go) - Player input-driven physics
- [`CollisionResolver`](internal/physics/resolve.go) - Axis-separated collision resolution

### Camera System

Located in [`internal/camera/`](internal/camera/).

```
+------------------+
|     Camera       |
+------------------+
| X, Y: float64    |  Position (top-left of viewport)
| ViewportW: int   |
| ViewportH: int   |
| DeadzoneX: float64 |
| DeadzoneY: float64 |
| DeadzoneW: float64 |
| DeadzoneH: float64 |
| LevelW: int      |
| LevelH: int      |
| Smoothing: float64 |
| targetX: float64 |
| targetY: float64 |
| PixelPerfect: bool |
+------------------+
| + CenterOn(x, y) |
| + Update()       |
| + WorldToScreen()|
| + ScreenToWorld()|
+------------------+
```

**Key Files:**
- [`Camera struct`](internal/camera/camera.go) - Advanced camera with deadzone and smoothing

### Graphics System

Located in [`internal/gfx/`](internal/gfx/).

```
+------------------+
|    Animation     |
+------------------+
| Frames: []*Image |
| FrameDuration: Duration |
| Loop: bool       |
+------------------+

+------------------+
|    Animator      |
+------------------+
| Animation: *Animation |
| elapsed: Duration |
| currentFrame: int |
| playing: bool    |
+------------------+
| + Update(dt)     |
| + Frame() *Image |
| + Play(), Stop() |
+------------------+

+------------------+
|     Sprite       |
+------------------+
| Image: *Image    |
| X, Y: float64    |
| ScaleX, ScaleY: float64 |
| Rotation: float64 |
| OriginX, OriginY: float64 |
+------------------+
| + Draw(screen, options) |
+------------------+
```

**Key Files:**
- [`Animation struct`](internal/gfx/animation.go) - Frame sequence definition
- [`Animator struct`](internal/gfx/animator.go) - Animation playback controller
- [`Sprite struct`](internal/gfx/sprite.go) - Positioned image with transform

### Gameplay System

Located in [`internal/gameplay/`](internal/gameplay/).

```
State (enum):
+------------------+
| StateRunning     |  Normal gameplay
| StateDead        |  Death animation
| StateRespawning  |  Respawn transition
| StateCompleted   |  Level complete
+------------------+

+------------------+
|   StateMachine   |
+------------------+
| Current: State   |
| RespawnX: float64 |
| RespawnY: float64 |
| DeathTimer: float64 |
| RespawnDelay: float64 |
| OnComplete: func() |
+------------------+
| + Update(dt)     |
| + TriggerDeath() |
| + TriggerComplete() |
| + FinishRespawn() |
| + IsRunning/Dead/Respawning/Completed() |
+------------------+

+------------------+
|   SpawnContext   |
+------------------+
| OnDeath: func()  |
| OnCheckpoint: func(id, x, y) |
| OnGoalReached: func() |
| GetDoor: func(id) *Door |
+------------------+
```

**Key Files:**
- [`State enum`](internal/gameplay/state.go:9) - Gameplay state enumeration
- [`StateMachine struct`](internal/gameplay/state.go:39) - State transition manager
- [`SpawnContext struct`](internal/gameplay/spawner.go) - Callback injection for entity spawning

---

## Abstraction Layers

The codebase is organized into distinct abstraction layers, each with specific responsibilities:

```
+============================================================================+
|                              LAYER DIAGRAM                                  |
+============================================================================+
|                                                                            |
|  Layer 1: SCENE LAYER                                                      |
|  +--------------------------------------------------------------------+   |
|  | internal/scenes/sandbox/scene.go                                   |   |
|  | - Coordinates all subsystems                                      |   |
|  | - Manages game state machine                                       |   |
|  | - Fixed timestep orchestration                                     |   |
|  +--------------------------------------------------------------------+   |
|                                    |                                       |
|  Layer 2: INPUT LAYER              |                                       |
|  +--------------------------------------------------------------------+   |
|  | internal/input/input.go                                            |   |
|  | - Action mapping (Jump, Left, Right, etc.)                        |   |
|  | - Key state tracking                                               |   |
|  | - JustPressed detection                                            |   |
|  +--------------------------------------------------------------------+   |
|                                    |                                       |
|  Layer 3: ENTITY LAYER             |                                       |
|  +--------------------------------------------------------------------+   |
|  | internal/entities/                                                 |   |
|  | - Entity interface (Update, Draw, Bounds)                         |   |
|  | - Trigger system (OnEnter, OnExit)                                |   |
|  | - SolidEntity (physics collision)                                 |   |
|  +--------------------------------------------------------------------+   |
|                                    |                                       |
|  Layer 4: RENDERING LAYER          |                                       |
|  +--------------------------------------------------------------------+   |
|  | internal/gfx/         | internal/world/render.go                  |   |
|  | - Animation/Animator  | - MapRenderer                              |   |
|  | - Sprite              | - Camera offset                            |   |
|  +--------------------------------------------------------------------+   |
|                                    |                                       |
|  Layer 5: PHYSICS LAYER           |                                       |
|  +--------------------------------------------------------------------+   |
|  | internal/physics/                                                  |   |
|  | - Body (position, velocity, size)                                 |   |
|  | - AABB (collision bounds)                                          |   |
|  | - Controller (input-driven physics)                               |   |
|  | - CollisionResolver (axis-separated resolution)                   |   |
|  +--------------------------------------------------------------------+   |
|                                    |                                       |
|  Layer 6: WORLD/MAP LAYER         |                                       |
|  +--------------------------------------------------------------------+   |
|  | internal/world/                                                    |   |
|  | - Map (tilemap with layers and tileset)                           |   |
|  | - TileLayer (2D tile ID storage)                                  |   |
|  | - CollisionMap (solid tile grid)                                  |   |
|  | - SolidGrid (boolean grid for collision queries)                  |   |
|  +--------------------------------------------------------------------+   |
|                                                                            |
+============================================================================+
```

### Layer Responsibilities

| Layer | Package | Responsibility |
|-------|---------|----------------|
| Scene | `internal/scenes/` | Game state coordination, entity spawning, level logic |
| Input | `internal/input/` | Input abstraction, action mapping, state tracking |
| Entity | `internal/entities/` | Game objects, triggers, collision responses |
| Rendering | `internal/gfx/`, `internal/world/render.go` | Visual output, camera transformation |
| Physics | `internal/physics/` | Movement, collision detection/resolution |
| World | `internal/world/` | Level data, tilemap, collision grid |

---

## Coordinate Systems and Transformations

### Coordinate System Types

```
+============================================================================+
|                        COORDINATE SYSTEMS                                   |
+============================================================================+

1. SCREEN/WINDOW COORDINATES
   +-------------------+
   | (0,0)      |      |
   |-----------+       |
   |           |       |
   |           +-------+
   |                (w,h)
   +-------------------+
   - Origin: Top-left of viewport
   - Units: Pixels (integers)
   - Purpose: Final rendering, mouse input

2. WORLD COORDINATES
   +-------------------+
   | (0,0)             |
   |                   |
   |    player @       |
   |    (wx,wy)        |
   |                   |
   |                   |
   +-------------------+
   - Origin: Top-left of game world
   - Units: Pixels (floating-point)
   - Purpose: Entity positions, physics, camera

3. TILE COORDINATES
   +---+---+---+---+
   |(0,0)|(1,0)|(2,0)|...
   +---+---+---+---+
   |(0,1)|(1,1)|...
   +---+---+---+
   |(0,2)|...
   +---+
   - Origin: Top-left tile
   - Units: Tile indices (integers)
   - Purpose: Map data, collision grid
```

### Coordinate Transformation Formulas

```
+============================================================================+
|                     TRANSFORMATION FORMULAS                                |
+============================================================================+

WORLD TO SCREEN:
  screenX = worldX - camera.X
  screenY = worldY - camera.Y
  
  Example: Player at world (500, 300), camera at (400, 200)
           screenX = 500 - 400 = 100
           screenY = 300 - 200 = 100
           Player renders at screen (100, 100)

SCREEN TO WORLD:
  worldX = screenX + camera.X
  worldY = screenY + camera.Y
  
  Example: Mouse click at screen (150, 100), camera at (400, 200)
           worldX = 150 + 400 = 550
           worldY = 100 + 200 = 300
           Click is at world position (550, 300)

WORLD TO TILE:
  tileX = int(worldX) / tileWidth
  tileY = int(worldY) / tileHeight
  
  Example: Position (150, 75) with 16x16 tiles
           tileX = 150 / 16 = 9
           tileY = 75 / 16 = 4
           Position is in tile (9, 4)

TILE TO WORLD:
  worldX = tileX * tileWidth
  worldY = tileY * tileHeight
  
  Example: Tile (5, 3) with 16x16 tiles
           worldX = 5 * 16 = 80
           worldY = 3 * 16 = 48
           Tile top-left corner at world (80, 48)
```

### AABB Bounds Calculations

The [`AABB`](internal/physics/aabb.go:5) struct uses top-left positioning:

```
AABB Bounds:
  
  (X,Y) -----> (X+W, Y)
    |              |
    |   AABB       |
    |              |
    v              v
  (X, Y+H) ---> (X+W, Y+H)
  
  Left()   = X
  Right()  = X + W
  Top()    = Y
  Bottom() = Y + H
```

### Implementation References

- [`Camera.WorldToScreen()`](internal/world/render.go:57) - World to screen conversion
- [`Camera.ScreenToWorld()`](internal/world/render.go:62) - Screen to world conversion
- [`WorldToTile()`](internal/world/collision.go:184) - World to tile conversion
- [`TileToWorld()`](internal/world/collision.go:189) - Tile to world conversion
- [`AABB.Intersects()`](internal/physics/aabb.go:31) - Overlap detection

---

## Event Processing System

### Event Types

The system supports several event types through the trigger mechanism:

| Event Type | Source | Callback | Effect |
|------------|--------|----------|--------|
| Hazard | [`Hazard.OnEnter()`](internal/entities/hazard.go) | `OnDeath` | Triggers player death |
| Goal | [`Goal.OnEnter()`](internal/entities/goal.go) | `OnComplete` | Level completion |
| Checkpoint | [`Checkpoint.OnEnter()`](internal/entities/checkpoint.go) | `OnActivate` | Updates respawn point |
| Switch | [`Switch.OnEnter()`](internal/entities/switch.go) | Direct door control | Opens/closes linked door |

### Game State Transitions

```
+============================================================================+
|                      GAME STATE MACHINE                                     |
+============================================================================+

                    +---------------+
                    | StateRunning  |<-----------------+
                    +---------------+                  |
                          |  |                         |
        Player dies       |  | Level complete          |
        (Hazard)          |  | (Goal)                  |
                          v  v                         |
                    +---------------+            +---------------+
                    |   StateDead   |            |StateCompleted |
                    +---------------+            +---------------+
                          |                            |
        Death timer       |                            |
        expires           |                            |
                          v                            |
                    +---------------+                  |
                    |StateRespawning|                  |
                    +---------------+                  |
                          |                            |
        Respawn           |                            |
        positioned        |                            |
                          +----------------------------+
```

### Event Detection Mechanism

Events are detected through AABB overlap tests in [`EntityWorld.CheckTriggers()`](internal/entities/world.go:86):

```
+============================================================================+
|                    TRIGGER DETECTION FLOW                                   |
+============================================================================+

Frame N:
  +-----------------+
  | Player AABB     |
  | (100,100,16,24) |
  +-----------------+
         |
         | No overlap with Hazard AABB
         v
  +-----------------+
  | Hazard AABB     |
  | (200,100,32,32) |
  +-----------------+
  wasTriggered = false

Frame N+1:
  +-----------------+
  | Player AABB     |
  | (180,100,16,24) |
  +-----------------+
         |
         | Overlap detected!
         | intersects = true
         | wasTriggered = false
         v
  +-----------------+
  | Hazard AABB     |
  | (200,100,32,32) |
  +-----------------+
  
  Result: OnEnter() called
          wasTriggered set to true

Frame N+2:
  +-----------------+
  | Player AABB     |
  | (250,100,16,24) |
  +-----------------+
         |
         | No overlap
         | intersects = false
         | wasTriggered = true
         v
  +-----------------+
  | Hazard AABB     |
  | (200,100,32,32) |
  +-----------------+
  
  Result: OnExit() called
          wasTriggered set to false
```

### Event Handling Patterns

#### Pattern 1: Callback Injection

The scene injects callbacks via [`SpawnContext`](internal/gameplay/spawner.go):

```go
// Scene creates SpawnContext with callbacks
spawnCtx := gameplay.SpawnContext{
    OnDeath: func() { stateMachine.TriggerDeath() },
    OnCheckpoint: func(id string, x, y float64) {
        stateMachine.SetRespawnPoint(x, y)
    },
    OnGoalReached: func() { stateMachine.TriggerComplete() },
    GetDoor: func(id string) *entities.Door { return doors[id] },
}
```

#### Pattern 2: State Machine

Centralized game state in [`StateMachine`](internal/gameplay/state.go:39):

```go
// State transitions
func (sm *StateMachine) TriggerDeath() {
    if sm.Current == StateRunning {
        sm.Current = StateDead
        sm.DeathTimer = 0
    }
}
```

#### Pattern 3: Direct Entity Coupling

Switch holds direct reference to Door:

```go
// Switch.OnEnter directly manipulates door
func (s *Switch) OnEnter(player *physics.Body) {
    if s.targetDoor != nil {
        s.targetDoor.Open()
    }
}
```

---

## Data Flow Diagrams

### Main Game Loop

```
+============================================================================+
|                         MAIN GAME LOOP                                      |
+============================================================================+

                    +-------------------+
                    |   ebiten.RunGame  |
                    +-------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                              App.Update()                                  |
|                         [internal/app/app.go:66]                          |
+---------------------------------------------------------------------------+
|                                                                           |
|  1. Handle debug toggle                                                   |
|  2. Handle quit action                                                    |
|  3. Fixed timestep physics loop:                                         |
|     +------------------+                                                  |
|     | timestep.AddFrameTime()                                             |
|     | while timestep.ShouldUpdate():                                      |
|     |     timestep.ConsumeTick()                                          |
|     |     scene.FixedUpdate()  <-- Physics at fixed rate                  |
|     +------------------+                                                  |
|  4. scene.Update(input)  <-- Non-physics update                          |
|  5. input.Update()  <-- Save previous key states                         |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                              App.Draw()                                    |
|                         [internal/app/app.go:104]                         |
+---------------------------------------------------------------------------+
|                                                                           |
|  1. scene.Draw(screen)                                                    |
|  2. if debugActive: drawDebugOverlay()                                   |
|                                                                           |
+---------------------------------------------------------------------------+
```

### Scene Update Flow

```
+============================================================================+
|                    SCENE FIXED UPDATE FLOW                                  |
+============================================================================+

                    +-------------------+
                    | scene.FixedUpdate |
                    +-------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Check Game State                                        |
+---------------------------------------------------------------------------+
|                                                                           |
|  if stateMachine.IsDead(): return                                        |
|  if stateMachine.IsRespawning(): handle respawn                         |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Process Input                                           |
+---------------------------------------------------------------------------+
|                                                                           |
|  controller.HandleInput(input)                                            |
|  - Horizontal movement                                                    |
|  - Jump input with buffering                                             |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Update Physics                                          |
+---------------------------------------------------------------------------+
|                                                                           |
|  controller.Update(dt, collisionMap)                                      |
|  1. Apply gravity                                                         |
|  2. Apply velocity to position                                           |
|  3. Resolve collisions                                                    |
|  4. Update ground state                                                   |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Check Triggers                                          |
+---------------------------------------------------------------------------+
|                                                                           |
|  entityWorld.CheckTriggers(player.Body)                                   |
|  - Test AABB overlap with all triggers                                   |
|  - Call OnEnter/OnExit based on state change                             |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Update Entities                                         |
+---------------------------------------------------------------------------+
|                                                                           |
|  entityWorld.Update(dt)                                                   |
|  - Update all entities                                                    |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Update Camera                                           |
+---------------------------------------------------------------------------+
|                                                                           |
|  camera.CenterOn(player.Body.PosX, player.Body.PosY)                     |
|  - Smooth follow with deadzone                                           |
|  - Constrain to level bounds                                             |
|                                                                           |
+---------------------------------------------------------------------------+
```

### Scene Draw Flow

```
+============================================================================+
|                      SCENE DRAW FLOW                                        |
+============================================================================+

                    +-------------------+
                    |   scene.Draw()    |
                    +-------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Clear Screen                                            |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Draw Background Layers                                  |
+---------------------------------------------------------------------------+
|                                                                           |
|  mapRenderer.DrawLayer(screen, "background", camX, camY)                 |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Draw Entities                                           |
+---------------------------------------------------------------------------+
|                                                                           |
|  entityWorld.Draw(screen, camX, camY)                                     |
|  - Draw all entities with camera offset                                   |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Draw Foreground Layers                                  |
+---------------------------------------------------------------------------+
|                                                                           |
|  mapRenderer.DrawLayer(screen, "foreground", camX, camY)                 |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Draw Player                                             |
+---------------------------------------------------------------------------+
|                                                                           |
|  screenX = player.Body.PosX - camX                                        |
|  screenY = player.Body.PosY - camY                                        |
|  playerSprite.Draw(screen, screenX, screenY)                              |
|                                                                           |
+---------------------------------------------------------------------------+
```

### Collision Resolution Flow

```
+============================================================================+
|                  COLLISION RESOLUTION FLOW                                  |
+============================================================================+

                    +-------------------+
                    | Resolve(body, map)|
                    +-------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Resolve X Axis                                          |
+---------------------------------------------------------------------------+
|                                                                           |
|  1. Apply velocity: body.PosX += body.VelX * dt                          |
|  2. Get overlapping tiles                                                 |
|  3. For each solid tile:                                                  |
|     - Calculate penetration depth                                        |
|     - Push body out (minimum penetration)                                |
|     - Set body.VelX = 0 if collision                                     |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Resolve Y Axis                                          |
+---------------------------------------------------------------------------+
|                                                                           |
|  1. Apply velocity: body.PosY += body.VelY * dt                          |
|  2. Get overlapping tiles                                                 |
|  3. For each solid tile:                                                  |
|     - Calculate penetration depth                                        |
|     - Push body out (minimum penetration)                                |
|     - Set body.VelY = 0 if collision                                     |
|     - Set body.OnGround = true if landing                                |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Update Ground State                                     |
+---------------------------------------------------------------------------+
|                                                                           |
|  - Check if standing on solid ground                                     |
|  - Update OnGround flag                                                   |
|                                                                           |
+---------------------------------------------------------------------------+
```

### Death/Respawn Event Flow

```
+============================================================================+
|                  DEATH/RESPAWN EVENT FLOW                                   |
+============================================================================+

                    +-------------------+
                    | Player overlaps   |
                    | Hazard trigger    |
                    +-------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Hazard.OnEnter()                                        |
|                         [hazard.go]                                        |
+---------------------------------------------------------------------------+
|                                                                           |
|  if OnDeath != nil:                                                       |
|      OnDeath()  --> stateMachine.TriggerDeath()                           |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    StateMachine.TriggerDeath()                             |
|                         [state.go:78]                                     |
+---------------------------------------------------------------------------+
|                                                                           |
|  Current = StateDead                                                      |
|  DeathTimer = 0                                                           |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    StateMachine.Update(dt)                                 |
|                         [state.go:64]                                     |
+---------------------------------------------------------------------------+
|                                                                           |
|  while Current == StateDead:                                              |
|      DeathTimer += dt                                                     |
|      if DeathTimer >= RespawnDelay:                                       |
|          Current = StateRespawning                                        |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    Scene handles StateRespawning                           |
+---------------------------------------------------------------------------+
|                                                                           |
|  player.Body.PosX = stateMachine.RespawnX                                 |
|  player.Body.PosY = stateMachine.RespawnY                                 |
|  player.Body.VelX = 0                                                     |
|  player.Body.VelY = 0                                                     |
|  stateMachine.FinishRespawn()                                             |
|                                                                           |
+---------------------------------------------------------------------------+
                            |
                            v
+---------------------------------------------------------------------------+
|                    StateMachine.FinishRespawn()                            |
|                         [state.go:96]                                     |
+---------------------------------------------------------------------------+
|                                                                           |
|  Current = StateRunning                                                   |
|  DeathTimer = 0                                                           |
|                                                                           |
+---------------------------------------------------------------------------+
```

---

## Summary

The GoP architecture demonstrates several key design principles:

1. **Separation of Concerns**: Each layer has a distinct responsibility
2. **Interface-based Design**: Entity, Trigger, and Scene use interfaces for flexibility
3. **Composition over Inheritance**: TriggerState, StateMachine use composition
4. **Fixed Timestep Physics**: Consistent physics regardless of frame rate
5. **Event-driven Triggers**: Clean separation between detection and response
6. **Coordinate System Clarity**: Explicit transformations between screen, world, and tile coordinates

This architecture supports a maintainable, extensible 2D platformer codebase with clear data flow and well-defined component interactions.