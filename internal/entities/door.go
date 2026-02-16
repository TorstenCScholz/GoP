package entities

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/torsten/GoP/internal/physics"
	"github.com/torsten/GoP/internal/world"
)

// Door is a SolidEntity that can open and close.
// When closed, it blocks player movement. When open, it has no collision.
type Door struct {
	body     *physics.Body
	id       string
	isOpen   bool
	closedW  float64 // Width when closed
	closedH  float64 // Height when closed
}

// NewDoor creates a new door at the given position.
// By default, doors start closed.
func NewDoor(x, y, w, h float64, id string) *Door {
	return &Door{
		body: &physics.Body{
			PosX: x,
			PosY: y,
			W:    w,
			H:    h,
		},
		id:      id,
		isOpen:  false,
		closedW: w,
		closedH: h,
	}
}

// Update implements Entity.
func (d *Door) Update(dt float64) {
	// Doors don't need per-frame updates
}

// Draw implements Entity.
// Deprecated: Use DrawWithContext for new implementations.
func (d *Door) Draw(screen *ebiten.Image, camX, camY float64) {
	if d.isOpen {
		// Draw open door (outline only)
		x := d.body.PosX - camX
		y := d.body.PosY - camY
		outlineColor := color.RGBA{100, 100, 100, 255}
		ebitenutil.DrawRect(screen, x, y, d.closedW, 2, outlineColor)
		ebitenutil.DrawRect(screen, x, y+d.closedH-2, d.closedW, 2, outlineColor)
		ebitenutil.DrawRect(screen, x, y, 2, d.closedH, outlineColor)
		ebitenutil.DrawRect(screen, x+d.closedW-2, y, 2, d.closedH, outlineColor)
	} else {
		// Draw closed door (solid)
		x := d.body.PosX - camX
		y := d.body.PosY - camY
		doorColor := color.RGBA{139, 90, 43, 255} // Brown
		ebitenutil.DrawRect(screen, x, y, d.body.W, d.body.H, doorColor)

		// Draw border
		borderColor := color.RGBA{80, 50, 20, 255}
		ebitenutil.DrawRect(screen, x, y, d.body.W, 2, borderColor)
		ebitenutil.DrawRect(screen, x, y+d.body.H-2, d.body.W, 2, borderColor)
		ebitenutil.DrawRect(screen, x, y, 2, d.body.H, borderColor)
		ebitenutil.DrawRect(screen, x+d.body.W-2, y, 2, d.body.H, borderColor)
	}
}

// DrawWithContext implements Entity.
func (d *Door) DrawWithContext(screen *ebiten.Image, ctx *world.RenderContext) {
	// Use WorldToScreen for coordinate conversion
	x, y := ctx.WorldToScreen(d.body.PosX, d.body.PosY)

	if d.isOpen {
		// Draw open door (outline only)
		outlineColor := color.RGBA{100, 100, 100, 255}
		ebitenutil.DrawRect(screen, x, y, d.closedW, 2, outlineColor)
		ebitenutil.DrawRect(screen, x, y+d.closedH-2, d.closedW, 2, outlineColor)
		ebitenutil.DrawRect(screen, x, y, 2, d.closedH, outlineColor)
		ebitenutil.DrawRect(screen, x+d.closedW-2, y, 2, d.closedH, outlineColor)
	} else {
		// Draw closed door (solid)
		doorColor := color.RGBA{139, 90, 43, 255} // Brown
		ebitenutil.DrawRect(screen, x, y, d.body.W, d.body.H, doorColor)

		// Draw border
		borderColor := color.RGBA{80, 50, 20, 255}
		ebitenutil.DrawRect(screen, x, y, d.body.W, 2, borderColor)
		ebitenutil.DrawRect(screen, x, y+d.body.H-2, d.body.W, 2, borderColor)
		ebitenutil.DrawRect(screen, x, y, 2, d.body.H, borderColor)
		ebitenutil.DrawRect(screen, x+d.body.W-2, y, 2, d.body.H, borderColor)
	}
}

// Bounds implements Entity.
func (d *Door) Bounds() physics.AABB {
	return d.body.AABB()
}

// GetBody implements SolidEntity.
func (d *Door) GetBody() *physics.Body {
	return d.body
}

// IsActive implements SolidEntity.
// A door is active when it is closed (blocking movement).
func (d *Door) IsActive() bool {
	return !d.isOpen
}

// GetID returns the door's identifier.
func (d *Door) GetID() string {
	return d.id
}

// IsOpen returns whether the door is open.
func (d *Door) IsOpen() bool {
	return d.isOpen
}

// Open opens the door (removes collision).
func (d *Door) Open() {
	d.isOpen = true
	d.body.W = 0
	d.body.H = 0
}

// Close closes the door (restores collision).
func (d *Door) Close() {
	d.isOpen = false
	d.body.W = d.closedW
	d.body.H = d.closedH
}

// Toggle switches the door state.
func (d *Door) Toggle() {
	if d.isOpen {
		d.Close()
	} else {
		d.Open()
	}
}

// Activate implements Targetable - opens the door.
func (d *Door) Activate() {
	d.Open()
}

// Deactivate implements Targetable - closes the door.
func (d *Door) Deactivate() {
	d.Close()
}

// TargetID implements Targetable - returns the door's unique identifier.
func (d *Door) TargetID() string {
	return d.id
}
