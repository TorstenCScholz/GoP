package entities

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/torsten/GoP/internal/physics"
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
	entities  []Entity
	triggers  []Trigger
	solidEnts []SolidEntity
}

// NewEntityWorld creates an empty entity world.
func NewEntityWorld() *EntityWorld {
	return &EntityWorld{
		entities:  make([]Entity, 0),
		triggers:  make([]Trigger, 0),
		solidEnts: make([]SolidEntity, 0),
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
func (w *EntityWorld) AddSolidEntity(e SolidEntity) {
	w.solidEnts = append(w.solidEnts, e)
	w.entities = append(w.entities, e) // Also add to general entities list
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

// Update updates all entities.
func (w *EntityWorld) Update(dt float64) {
	for _, e := range w.entities {
		e.Update(dt)
	}
}

// Draw renders all entities with camera offset.
func (w *EntityWorld) Draw(screen *ebiten.Image, camX, camY float64) {
	for _, e := range w.entities {
		e.Draw(screen, camX, camY)
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