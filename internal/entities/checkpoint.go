package entities

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/torsten/GoP/internal/physics"
	"github.com/torsten/GoP/internal/world"
)

// Checkpoint updates the respawn point when touched.
type Checkpoint struct {
	bounds    physics.AABB
	id        string
	state     TriggerState
	triggered bool // Has been activated at least once

	// Callback when checkpoint is activated
	OnActivate func(id string, x, y float64)
}

// NewCheckpoint creates a new checkpoint at the given position.
func NewCheckpoint(x, y, w, h float64, id string) *Checkpoint {
	return &Checkpoint{
		bounds:    physics.AABB{X: x, Y: y, W: w, H: h},
		id:        id,
		state:     NewTriggerState(),
		triggered: false,
	}
}

// Update implements Entity.
func (c *Checkpoint) Update(dt float64) {
	// Checkpoints don't need per-frame updates
}

// Draw implements Entity.
// Deprecated: Use DrawWithContext for new implementations.
func (c *Checkpoint) Draw(screen *ebiten.Image, camX, camY float64) {
	x := c.bounds.X - camX
	y := c.bounds.Y - camY

	// Draw checkpoint indicator
	var col color.RGBA
	if c.triggered {
		// Green when activated
		col = color.RGBA{0, 255, 0, 128}
	} else {
		// Yellow when not yet activated
		col = color.RGBA{255, 255, 0, 128}
	}
	ebitenutil.DrawRect(screen, x, y, c.bounds.W, c.bounds.H, col)
}

// DrawWithContext implements Entity.
func (c *Checkpoint) DrawWithContext(screen *ebiten.Image, ctx *world.RenderContext) {
	x, y := ctx.WorldToScreen(c.bounds.X, c.bounds.Y)

	// Draw checkpoint indicator
	var col color.RGBA
	if c.triggered {
		// Green when activated
		col = color.RGBA{0, 255, 0, 128}
	} else {
		// Yellow when not yet activated
		col = color.RGBA{255, 255, 0, 128}
	}
	ebitenutil.DrawRect(screen, x, y, c.bounds.W, c.bounds.H, col)
}

// Bounds implements Entity.
func (c *Checkpoint) Bounds() physics.AABB {
	return c.bounds
}

// OnEnter implements Trigger.
func (c *Checkpoint) OnEnter(player *physics.Body) {
	if !c.triggered {
		c.triggered = true
		if c.OnActivate != nil {
			c.OnActivate(c.id, c.bounds.X, c.bounds.Y)
		}
	}
}

// OnExit implements Trigger.
func (c *Checkpoint) OnExit(player *physics.Body) {
	// Nothing to do on exit
}

// IsActive implements Trigger.
func (c *Checkpoint) IsActive() bool {
	return c.state.IsActive()
}

// WasTriggered implements Trigger.
func (c *Checkpoint) WasTriggered() bool {
	return c.state.WasTriggered()
}

// SetTriggered implements Trigger.
func (c *Checkpoint) SetTriggered(triggered bool) {
	c.state.SetTriggered(triggered)
}

// ID returns the checkpoint identifier.
func (c *Checkpoint) ID() string {
	return c.id
}

// IsTriggered returns whether this checkpoint has been activated.
func (c *Checkpoint) IsTriggered() bool {
	return c.triggered
}
