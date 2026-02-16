// Package entities provides game entity interfaces and implementations.
package entities

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/torsten/GoP/internal/physics"
	"github.com/torsten/GoP/internal/world"
)

// Entity is the base interface for all game entities.
// Entities are objects that exist in the world and can be updated/drawn.
type Entity interface {
	// Update called each frame with delta time.
	Update(dt float64)

	// Draw renders the entity to screen with camera offset.
	// Deprecated: Use DrawWithContext for new implementations.
	Draw(screen *ebiten.Image, camX, camY float64)

	// DrawWithContext renders the entity using a RenderContext.
	// This is the preferred method for rendering entities.
	DrawWithContext(screen *ebiten.Image, ctx *world.RenderContext)

	// Bounds returns the entity's AABB for overlap queries.
	Bounds() physics.AABB
}

// Trigger is an entity that responds to player overlap.
// Triggers do not have solid collision - they use AABB overlap tests.
type Trigger interface {
	Entity

	// OnEnter called when player enters the trigger zone.
	OnEnter(player *physics.Body)

	// OnExit called when player exits the trigger zone.
	OnExit(player *physics.Body)

	// IsActive returns whether the trigger is currently active.
	IsActive() bool

	// WasTriggered returns true if player was inside last frame.
	WasTriggered() bool

	// SetTriggered sets the triggered state for frame tracking.
	SetTriggered(triggered bool)
}

// SolidEntity is an entity with physics collision.
// These entities participate in tile collision resolution.
type SolidEntity interface {
	Entity

	// Body returns the physics body for collision resolution.
	Body() *physics.Body
}

// TriggerState provides common state for trigger implementations.
type TriggerState struct {
	Active    bool
	Triggered bool // True if player was inside last frame
}

// NewTriggerState creates a new trigger state (active by default).
func NewTriggerState() TriggerState {
	return TriggerState{Active: true, Triggered: false}
}

// IsActive returns whether the trigger is active.
func (s *TriggerState) IsActive() bool {
	return s.Active
}

// WasTriggered returns whether player was inside last frame.
func (s *TriggerState) WasTriggered() bool {
	return s.Triggered
}

// SetTriggered sets the triggered state.
func (s *TriggerState) SetTriggered(triggered bool) {
	s.Triggered = triggered
}

// SetActive sets the active state.
func (s *TriggerState) SetActive(active bool) {
	s.Active = active
}
