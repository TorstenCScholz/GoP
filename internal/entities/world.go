package entities

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/torsten/GoP/internal/physics"
	"github.com/torsten/GoP/internal/world"
)

// MapInterface provides the minimal map interface needed by EntityWorld.
// This avoids importing the world package.
type MapInterface interface {
	PixelWidth() int
	PixelHeight() int
}

// CollisionMapInterface provides the minimal collision interface needed by EntityWorld.
type CollisionMapInterface interface {
	IsSolidAtTile(tx, ty int) bool
	OverlapsSolid(x, y, w, h float64) bool
}

// EntityWorld holds entities and provides trigger checking.
// It does not hold the map or collision data directly - those are passed in.
type EntityWorld struct {
	entities   []Entity
	triggers   []Trigger
	solidEnts  []SolidEntity
	kinematics []physics.Kinematic

	// TargetRegistry manages ID-to-target lookups for switches, etc.
	TargetRegistry *TargetRegistry
}

// NewEntityWorld creates an empty entity world.
func NewEntityWorld() *EntityWorld {
	return &EntityWorld{
		entities:       make([]Entity, 0),
		triggers:       make([]Trigger, 0),
		solidEnts:      make([]SolidEntity, 0),
		TargetRegistry: NewTargetRegistry(),
	}
}

// AddEntity adds an entity to the world.
func (w *EntityWorld) AddEntity(e Entity) {
	w.entities = append(w.entities, e)
}

// AddTrigger adds a trigger to the world.
func (w *EntityWorld) AddTrigger(t Trigger) {
	w.triggers = append(w.triggers, t)
	w.entities = append(w.entities, t) // Also add to general entities list
}

// AddSolidEntity adds a solid entity to the world.
// If the entity implements Targetable, it is also registered with the TargetRegistry.
func (w *EntityWorld) AddSolidEntity(e SolidEntity) {
	w.solidEnts = append(w.solidEnts, e)
	w.entities = append(w.entities, e) // Also add to general entities list

	// Auto-register Targetable entities
	if t, ok := e.(Targetable); ok {
		w.TargetRegistry.Register(t)
	}
}

// RegisterTarget adds a Targetable entity to the registry.
func (w *EntityWorld) RegisterTarget(t Targetable) {
	w.TargetRegistry.Register(t)
}

// Entities returns all entities.
func (w *EntityWorld) Entities() []Entity {
	return w.entities
}

// Triggers returns all triggers.
func (w *EntityWorld) Triggers() []Trigger {
	return w.triggers
}

// SolidEntities returns all solid entities.
func (w *EntityWorld) SolidEntities() []SolidEntity {
	return w.solidEnts
}

// AddKinematic adds a kinematic entity to the world.
func (w *EntityWorld) AddKinematic(k physics.Kinematic) {
	w.kinematics = append(w.kinematics, k)
}

// GetKinematics returns all kinematic entities.
func (w *EntityWorld) GetKinematics() []physics.Kinematic {
	return w.kinematics
}

// ActiveSolidAABBs returns unique AABBs for all active solid bodies.
// Kinematics are included, but bodies already present in solid entities are deduplicated.
func (w *EntityWorld) ActiveSolidAABBs() []physics.AABB {
	var solids []physics.AABB
	seenBodies := make(map[*physics.Body]struct{})

	addBody := func(body *physics.Body) {
		if body == nil || body.W <= 0 || body.H <= 0 {
			return
		}
		if _, seen := seenBodies[body]; seen {
			return
		}
		seenBodies[body] = struct{}{}
		solids = append(solids, body.AABB())
	}

	for _, e := range w.solidEnts {
		if !e.IsActive() {
			continue
		}
		addBody(e.GetBody())
	}

	for _, k := range w.kinematics {
		if !k.IsActive() {
			continue
		}
		addBody(k.GetBody())
	}

	return solids
}

// UpdateKinematics updates all kinematic entities with collision detection.
func (w *EntityWorld) UpdateKinematics(collisionMap *world.CollisionMap, dt float64) {
	for _, k := range w.kinematics {
		if k.IsActive() {
			k.MoveAndSlide(collisionMap, dt)
		}
	}
}

// Update updates all entities.
func (w *EntityWorld) Update(dt float64) {
	for _, e := range w.entities {
		e.Update(dt)
	}
}

// Draw renders all entities with camera offset.
// Deprecated: Use DrawWithContext for new implementations.
func (w *EntityWorld) Draw(screen *ebiten.Image, camX, camY float64) {
	for _, e := range w.entities {
		e.Draw(screen, camX, camY)
	}
}

// DrawWithContext renders all entities using a RenderContext.
func (w *EntityWorld) DrawWithContext(screen *ebiten.Image, ctx *world.RenderContext) {
	for _, e := range w.entities {
		e.DrawWithContext(screen, ctx)
	}
}

// CheckTriggers tests player against all triggers.
// Returns true if any trigger was activated.
func (w *EntityWorld) CheckTriggers(player *physics.Body) bool {
	playerAABB := player.AABB()
	anyTriggered := false

	for _, t := range w.triggers {
		if !t.IsActive() {
			continue
		}

		intersects := playerAABB.Intersects(t.Bounds())
		wasTriggered := t.WasTriggered()

		if intersects && !wasTriggered {
			// Player just entered the trigger
			t.OnEnter(player)
			t.SetTriggered(true)
			anyTriggered = true
		} else if !intersects && wasTriggered {
			// Player just exited the trigger
			t.OnExit(player)
			t.SetTriggered(false)
		}
	}

	return anyTriggered
}

// FindDoorByID finds a solid entity that is a door with the given ID.
// Returns nil if not found.
func (w *EntityWorld) FindDoorByID(id string) SolidEntity {
	for _, e := range w.solidEnts {
		// Check if it's a door by checking if it has a GetID method
		if door, ok := e.(interface{ GetID() string }); ok {
			if door.GetID() == id {
				return e
			}
		}
	}
	return nil
}

// OverlapsSolidEntity checks if the given AABB overlaps any solid entity.
func (w *EntityWorld) OverlapsSolidEntity(aabb physics.AABB) bool {
	for _, e := range w.solidEnts {
		if aabb.Intersects(e.Bounds()) {
			return true
		}
	}
	return false
}

// DrawKinematicsDebug draws debug visualization for all kinematic entities.
// Entities must implement the DebugDrawable interface to be drawn.
func (w *EntityWorld) DrawKinematicsDebug(screen *ebiten.Image, ctx *world.RenderContext) {
	for _, k := range w.kinematics {
		// Check if the kinematic implements DebugDrawable
		if dd, ok := k.(DebugDrawable); ok {
			dd.DrawDebug(screen, ctx)
		}
	}
}

// DebugDrawable is an interface for entities that can draw debug visualization.
type DebugDrawable interface {
	DrawDebug(screen *ebiten.Image, ctx *world.RenderContext)
}
