package entities

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/torsten/GoP/internal/physics"
	"github.com/torsten/GoP/internal/world"
)

// DebugRenderer draws entity bounds and state info for debugging.
type DebugRenderer struct {
	ShowTriggers bool
	ShowBounds   bool
	ShowState    bool
	ShowAll      bool
}

// NewDebugRenderer creates a new debug renderer.
func NewDebugRenderer() *DebugRenderer {
	return &DebugRenderer{}
}

// TriggerChecker provides an interface for checking trigger state.
// This avoids importing the world package.
type TriggerChecker interface {
	Triggers() []Trigger
	Entities() []Entity
}

// Draw renders debug information for the entity world.
// Deprecated: Use DrawWithContext for new implementations.
func (d *DebugRenderer) Draw(screen *ebiten.Image, checker TriggerChecker, camX, camY float64) {
	if d.ShowTriggers || d.ShowAll {
		d.drawTriggers(screen, checker, camX, camY)
	}
	if d.ShowBounds || d.ShowAll {
		d.drawBounds(screen, checker, camX, camY)
	}
}

// DrawWithContext renders debug information using a RenderContext.
func (d *DebugRenderer) DrawWithContext(screen *ebiten.Image, checker TriggerChecker, ctx *world.RenderContext) {
	if d.ShowTriggers || d.ShowAll {
		d.drawTriggersWithContext(screen, checker, ctx)
	}
	if d.ShowBounds || d.ShowAll {
		d.drawBoundsWithContext(screen, checker, ctx)
	}
}

// drawTriggers draws all trigger zones.
// Deprecated: Use drawTriggersWithContext for new implementations.
func (d *DebugRenderer) drawTriggers(screen *ebiten.Image, checker TriggerChecker, camX, camY float64) {
	triggerColor := color.RGBA{255, 255, 0, 100}    // Yellow semi-transparent
	activeColor := color.RGBA{0, 255, 0, 100}       // Green for active triggers
	inactiveColor := color.RGBA{100, 100, 100, 100} // Gray for inactive

	for _, t := range checker.Triggers() {
		bounds := t.Bounds()
		x := bounds.X - camX
		y := bounds.Y - camY

		var col color.RGBA
		if t.IsActive() {
			if t.WasTriggered() {
				col = activeColor
			} else {
				col = triggerColor
			}
		} else {
			col = inactiveColor
		}

		ebitenutil.DrawRect(screen, x, y, bounds.W, bounds.H, col)
	}
}

// drawTriggersWithContext draws all trigger zones using RenderContext.
func (d *DebugRenderer) drawTriggersWithContext(screen *ebiten.Image, checker TriggerChecker, ctx *world.RenderContext) {
	triggerColor := color.RGBA{255, 255, 0, 100}    // Yellow semi-transparent
	activeColor := color.RGBA{0, 255, 0, 100}       // Green for active triggers
	inactiveColor := color.RGBA{100, 100, 100, 100} // Gray for inactive

	for _, t := range checker.Triggers() {
		bounds := t.Bounds()
		x, y := ctx.WorldToScreen(bounds.X, bounds.Y)

		var col color.RGBA
		if t.IsActive() {
			if t.WasTriggered() {
				col = activeColor
			} else {
				col = triggerColor
			}
		} else {
			col = inactiveColor
		}

		ebitenutil.DrawRect(screen, x, y, bounds.W, bounds.H, col)
	}
}

// drawBounds draws entity bounding boxes.
// Deprecated: Use drawBoundsWithContext for new implementations.
func (d *DebugRenderer) drawBounds(screen *ebiten.Image, checker TriggerChecker, camX, camY float64) {
	entityColor := color.RGBA{0, 255, 255, 200} // Cyan for entities
	solidColor := color.RGBA{255, 0, 255, 200}  // Magenta for solid entities
	borderWidth := 1.0

	for _, e := range checker.Entities() {
		bounds := e.Bounds()
		x := bounds.X - camX
		y := bounds.Y - camY

		// Determine color based on entity type
		col := entityColor
		if _, isSolid := e.(SolidEntity); isSolid {
			col = solidColor
		}

		// Draw border only
		ebitenutil.DrawRect(screen, x, y, bounds.W, borderWidth, col)
		ebitenutil.DrawRect(screen, x, y+bounds.H-borderWidth, bounds.W, borderWidth, col)
		ebitenutil.DrawRect(screen, x, y, borderWidth, bounds.H, col)
		ebitenutil.DrawRect(screen, x+bounds.W-borderWidth, y, borderWidth, bounds.H, col)
	}
}

// drawBoundsWithContext draws entity bounding boxes using RenderContext.
func (d *DebugRenderer) drawBoundsWithContext(screen *ebiten.Image, checker TriggerChecker, ctx *world.RenderContext) {
	entityColor := color.RGBA{0, 255, 255, 200} // Cyan for entities
	solidColor := color.RGBA{255, 0, 255, 200}  // Magenta for solid entities
	borderWidth := 1.0

	for _, e := range checker.Entities() {
		bounds := e.Bounds()
		x, y := ctx.WorldToScreen(bounds.X, bounds.Y)

		// Determine color based on entity type
		col := entityColor
		if _, isSolid := e.(SolidEntity); isSolid {
			col = solidColor
		}

		// Draw border only
		ebitenutil.DrawRect(screen, x, y, bounds.W, borderWidth, col)
		ebitenutil.DrawRect(screen, x, y+bounds.H-borderWidth, bounds.W, borderWidth, col)
		ebitenutil.DrawRect(screen, x, y, borderWidth, bounds.H, col)
		ebitenutil.DrawRect(screen, x+bounds.W-borderWidth, y, borderWidth, bounds.H, col)
	}
}

// DrawPlayerDebug draws debug info for the player body.
// Deprecated: Use DrawPlayerDebugWithContext for new implementations.
func (d *DebugRenderer) DrawPlayerDebug(screen *ebiten.Image, player *physics.Body, camX, camY float64) {
	if !d.ShowBounds && !d.ShowAll {
		return
	}

	x := player.PosX - camX
	y := player.PosY - camY
	col := color.RGBA{255, 255, 255, 255}
	borderWidth := 1.0

	// Draw border
	ebitenutil.DrawRect(screen, x, y, player.W, borderWidth, col)
	ebitenutil.DrawRect(screen, x, y+player.H-borderWidth, player.W, borderWidth, col)
	ebitenutil.DrawRect(screen, x, y, borderWidth, player.H, col)
	ebitenutil.DrawRect(screen, x+player.W-borderWidth, y, borderWidth, player.H, col)
}

// DrawPlayerDebugWithContext draws debug info for the player body using RenderContext.
func (d *DebugRenderer) DrawPlayerDebugWithContext(screen *ebiten.Image, player *physics.Body, ctx *world.RenderContext) {
	if !d.ShowBounds && !d.ShowAll {
		return
	}

	x, y := ctx.WorldToScreen(player.PosX, player.PosY)
	col := color.RGBA{255, 255, 255, 255}
	borderWidth := 1.0

	// Draw border
	ebitenutil.DrawRect(screen, x, y, player.W, borderWidth, col)
	ebitenutil.DrawRect(screen, x, y+player.H-borderWidth, player.W, borderWidth, col)
	ebitenutil.DrawRect(screen, x, y, borderWidth, player.H, col)
	ebitenutil.DrawRect(screen, x+player.W-borderWidth, y, borderWidth, player.H, col)
}

// ToggleAll toggles all debug displays.
func (d *DebugRenderer) ToggleAll() {
	d.ShowAll = !d.ShowAll
}
