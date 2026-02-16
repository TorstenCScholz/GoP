package entities

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/torsten/GoP/internal/physics"
	"github.com/torsten/GoP/internal/world"
)

// Hazard kills the player on touch.
type Hazard struct {
	bounds physics.AABB
	state  TriggerState

	// Callback to trigger death
	OnDeath func()
}

// NewHazard creates a new hazard at the given position.
func NewHazard(x, y, w, h float64) *Hazard {
	return &Hazard{
		bounds: physics.AABB{X: x, Y: y, W: w, H: h},
		state:  NewTriggerState(),
	}
}

// Update implements Entity.
func (h *Hazard) Update(dt float64) {
	// Hazards don't need per-frame updates
}

// Draw implements Entity.
// Deprecated: Use DrawWithContext for new implementations.
func (h *Hazard) Draw(screen *ebiten.Image, camX, camY float64) {
	// Draw hazard indicator (red semi-transparent)
	if h.state.Active {
		x := h.bounds.X - camX
		y := h.bounds.Y - camY
		hazardColor := color.RGBA{255, 0, 0, 128}
		ebitenutil.DrawRect(screen, x, y, h.bounds.W, h.bounds.H, hazardColor)
	}
}

// DrawWithContext implements Entity.
func (h *Hazard) DrawWithContext(screen *ebiten.Image, ctx *world.RenderContext) {
	// Draw hazard indicator (red semi-transparent)
	if h.state.Active {
		x, y := ctx.WorldToScreen(h.bounds.X, h.bounds.Y)
		hazardColor := color.RGBA{255, 0, 0, 128}
		ebitenutil.DrawRect(screen, x, y, h.bounds.W, h.bounds.H, hazardColor)
	}
}

// Bounds implements Entity.
func (h *Hazard) Bounds() physics.AABB {
	return h.bounds
}

// OnEnter implements Trigger.
func (h *Hazard) OnEnter(player *physics.Body) {
	if h.OnDeath != nil {
		h.OnDeath()
	}
}

// OnExit implements Trigger.
func (h *Hazard) OnExit(player *physics.Body) {
	// Nothing to do on exit
}

// IsActive implements Trigger.
func (h *Hazard) IsActive() bool {
	return h.state.IsActive()
}

// WasTriggered implements Trigger.
func (h *Hazard) WasTriggered() bool {
	return h.state.WasTriggered()
}

// SetTriggered implements Trigger.
func (h *Hazard) SetTriggered(triggered bool) {
	h.state.SetTriggered(triggered)
}

// SetActive sets whether the hazard is active.
func (h *Hazard) SetActive(active bool) {
	h.state.SetActive(active)
}
