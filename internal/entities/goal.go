package entities

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/torsten/GoP/internal/physics"
	"github.com/torsten/GoP/internal/world"
)

// Goal triggers level completion when touched.
type Goal struct {
	bounds physics.AABB
	state  TriggerState

	// Callback when goal is reached
	OnComplete func()
}

// NewGoal creates a new goal at the given position.
func NewGoal(x, y, w, h float64) *Goal {
	return &Goal{
		bounds: physics.AABB{X: x, Y: y, W: w, H: h},
		state:  NewTriggerState(),
	}
}

// Update implements Entity.
func (g *Goal) Update(dt float64) {
	// Goals don't need per-frame updates
}

// Draw implements Entity.
// Deprecated: Use DrawWithContext for new implementations.
func (g *Goal) Draw(screen *ebiten.Image, camX, camY float64) {
	if !g.state.Active {
		return
	}

	x := g.bounds.X - camX
	y := g.bounds.Y - camY

	// Draw goal indicator (green/blue gradient effect)
	goalColor := color.RGBA{0, 200, 255, 128}
	ebitenutil.DrawRect(screen, x, y, g.bounds.W, g.bounds.H, goalColor)

	// Draw border
	borderColor := color.RGBA{255, 255, 255, 255}
	ebitenutil.DrawRect(screen, x, y, g.bounds.W, 2, borderColor)
	ebitenutil.DrawRect(screen, x, y+g.bounds.H-2, g.bounds.W, 2, borderColor)
	ebitenutil.DrawRect(screen, x, y, 2, g.bounds.H, borderColor)
	ebitenutil.DrawRect(screen, x+g.bounds.W-2, y, 2, g.bounds.H, borderColor)
}

// DrawWithContext implements Entity.
func (g *Goal) DrawWithContext(screen *ebiten.Image, ctx *world.RenderContext) {
	if !g.state.Active {
		return
	}

	x, y := ctx.WorldToScreen(g.bounds.X, g.bounds.Y)

	// Draw goal indicator (green/blue gradient effect)
	goalColor := color.RGBA{0, 200, 255, 128}
	ebitenutil.DrawRect(screen, x, y, g.bounds.W, g.bounds.H, goalColor)

	// Draw border
	borderColor := color.RGBA{255, 255, 255, 255}
	ebitenutil.DrawRect(screen, x, y, g.bounds.W, 2, borderColor)
	ebitenutil.DrawRect(screen, x, y+g.bounds.H-2, g.bounds.W, 2, borderColor)
	ebitenutil.DrawRect(screen, x, y, 2, g.bounds.H, borderColor)
	ebitenutil.DrawRect(screen, x+g.bounds.W-2, y, 2, g.bounds.H, borderColor)
}

// Bounds implements Entity.
func (g *Goal) Bounds() physics.AABB {
	return g.bounds
}

// OnEnter implements Trigger.
func (g *Goal) OnEnter(player *physics.Body) {
	if g.state.Active && g.OnComplete != nil {
		g.OnComplete()
		g.state.Active = false // Deactivate after triggering
	}
}

// OnExit implements Trigger.
func (g *Goal) OnExit(player *physics.Body) {
	// Nothing to do on exit
}

// IsActive implements Trigger.
func (g *Goal) IsActive() bool {
	return g.state.IsActive()
}

// WasTriggered implements Trigger.
func (g *Goal) WasTriggered() bool {
	return g.state.WasTriggered()
}

// SetTriggered implements Trigger.
func (g *Goal) SetTriggered(triggered bool) {
	g.state.SetTriggered(triggered)
}

// SetActive sets whether the goal is active.
func (g *Goal) SetActive(active bool) {
	g.state.SetActive(active)
}
