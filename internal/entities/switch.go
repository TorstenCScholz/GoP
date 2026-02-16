package entities

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/torsten/GoP/internal/physics"
	"github.com/torsten/GoP/internal/world"
)

// Switch is a trigger that controls Targetable entities (doors, etc.).
// When touched, it can toggle or set the state of a target.
type Switch struct {
	bounds     physics.AABB
	state      TriggerState
	targetID   string
	toggleMode bool // true = toggle, false = one-shot open
	once       bool // true = deactivate after use
	used       bool // Has been used (for once mode)

	// Registry for resolving targets at runtime
	registry *TargetRegistry
}

// NewSwitch creates a new switch at the given position.
func NewSwitch(x, y, w, h float64, targetID string) *Switch {
	return &Switch{
		bounds:     physics.AABB{X: x, Y: y, W: w, H: h},
		state:      NewTriggerState(),
		targetID:   targetID,
		toggleMode: true, // Default to toggle mode
		once:       false,
	}
}

// SetToggleMode sets whether the switch toggles or one-shot opens.
func (s *Switch) SetToggleMode(toggle bool) {
	s.toggleMode = toggle
}

// SetOnce sets whether the switch deactivates after use.
func (s *Switch) SetOnce(once bool) {
	s.once = once
}

// SetRegistry sets the target registry for resolving targets at runtime.
func (s *Switch) SetRegistry(registry *TargetRegistry) {
	s.registry = registry
}

// GetTargetID returns the target ID for this switch.
func (s *Switch) GetTargetID() string {
	return s.targetID
}

// Update implements Entity.
func (s *Switch) Update(dt float64) {
	// Switches don't need per-frame updates
}

// Draw implements Entity.
// Deprecated: Use DrawWithContext for new implementations.
func (s *Switch) Draw(screen *ebiten.Image, camX, camY float64) {
	x := s.bounds.X - camX
	y := s.bounds.Y - camY

	var col color.RGBA
	if !s.state.Active {
		// Gray when deactivated
		col = color.RGBA{100, 100, 100, 255}
	} else if s.used {
		// Blue when used (if once mode)
		col = color.RGBA{100, 100, 255, 255}
	} else {
		// Yellow/orange when ready
		col = color.RGBA{255, 200, 0, 255}
	}

	// Draw switch body
	ebitenutil.DrawRect(screen, x, y, s.bounds.W, s.bounds.H, col)

	// Draw border
	borderColor := color.RGBA{50, 50, 50, 255}
	ebitenutil.DrawRect(screen, x, y, s.bounds.W, 2, borderColor)
	ebitenutil.DrawRect(screen, x, y+s.bounds.H-2, s.bounds.W, 2, borderColor)
	ebitenutil.DrawRect(screen, x, y, 2, s.bounds.H, borderColor)
	ebitenutil.DrawRect(screen, x+s.bounds.W-2, y, 2, s.bounds.H, borderColor)
}

// DrawWithContext implements Entity.
func (s *Switch) DrawWithContext(screen *ebiten.Image, ctx *world.RenderContext) {
	// Use WorldToScreen for coordinate conversion
	x, y := ctx.WorldToScreen(s.bounds.X, s.bounds.Y)

	var col color.RGBA
	if !s.state.Active {
		// Gray when deactivated
		col = color.RGBA{100, 100, 100, 255}
	} else if s.used {
		// Blue when used (if once mode)
		col = color.RGBA{100, 100, 255, 255}
	} else {
		// Yellow/orange when ready
		col = color.RGBA{255, 200, 0, 255}
	}

	// Draw switch body
	ebitenutil.DrawRect(screen, x, y, s.bounds.W, s.bounds.H, col)

	// Draw border
	borderColor := color.RGBA{50, 50, 50, 255}
	ebitenutil.DrawRect(screen, x, y, s.bounds.W, 2, borderColor)
	ebitenutil.DrawRect(screen, x, y+s.bounds.H-2, s.bounds.W, 2, borderColor)
	ebitenutil.DrawRect(screen, x, y, 2, s.bounds.H, borderColor)
	ebitenutil.DrawRect(screen, x+s.bounds.W-2, y, 2, s.bounds.H, borderColor)
}

// Bounds implements Entity.
func (s *Switch) Bounds() physics.AABB {
	return s.bounds
}

// OnEnter implements Trigger.
func (s *Switch) OnEnter(player *physics.Body) {
	if !s.state.Active {
		return
	}

	// Resolve target from registry
	if s.registry == nil {
		return
	}

	target := s.registry.Resolve(s.targetID)
	if target == nil {
		return
	}

	// Execute switch action
	if s.toggleMode {
		target.Toggle()
	} else {
		target.Activate()
	}

	// Mark as used if once mode
	if s.once {
		s.used = true
		s.state.Active = false
	}
}

// OnExit implements Trigger.
func (s *Switch) OnExit(player *physics.Body) {
	// Nothing to do on exit
}

// IsActive implements Trigger.
func (s *Switch) IsActive() bool {
	return s.state.IsActive()
}

// WasTriggered implements Trigger.
func (s *Switch) WasTriggered() bool {
	return s.state.WasTriggered()
}

// SetTriggered implements Trigger.
func (s *Switch) SetTriggered(triggered bool) {
	s.state.SetTriggered(triggered)
}
