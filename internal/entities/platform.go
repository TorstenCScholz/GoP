package entities

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/torsten/GoP/internal/physics"
	"github.com/torsten/GoP/internal/world"
)

// MovingPlatform is a solid entity that moves between two points (A and B).
// It implements the physics.Kinematic interface for integration with the physics system.
type MovingPlatform struct {
	id     string
	body   physics.Body
	active bool

	// Path definition
	startX, startY float64 // Initial position (point A)
	endX, endY     float64 // Target position (point B)

	// Movement state
	velocityX, velocityY float64
	speed                float64
	goingToEnd           bool    // true = A→B, false = B→A
	waitTimer            float64 // Time remaining before moving
	waitTime             float64 // Time to wait at endpoints

	// Options
	pushPlayer bool // Whether to push player sideways
}

// NewMovingPlatform creates a new moving platform.
// x, y is the initial position (point A).
// w, h is the platform size.
// endX, endY is the target position (point B) - this is an ABSOLUTE position, not relative.
// speed is movement speed in pixels/second.
func NewMovingPlatform(id string, x, y, w, h, endX, endY, speed float64) *MovingPlatform {
	return &MovingPlatform{
		id:     id,
		active: true,
		body: physics.Body{
			PosX: x,
			PosY: y,
			W:    w,
			H:    h,
		},
		startX:     x,
		startY:     y,
		endX:       endX,
		endY:       endY,
		speed:      speed,
		goingToEnd: true,
		waitTime:   0.5, // Default wait time at endpoints
		pushPlayer: false,
	}
}

// GetBody returns a pointer to the platform's physics body.
// Implements physics.SolidEntity interface.
func (p *MovingPlatform) GetBody() *physics.Body {
	return &p.body
}

// IsActive returns whether the platform is active.
// Platforms are always active for now.
// Implements physics.SolidEntity interface.
func (p *MovingPlatform) IsActive() bool {
	return p.active
}

// Velocity returns the current velocity in pixels per second.
// Implements physics.Kinematic interface.
func (p *MovingPlatform) Velocity() (vx, vy float64) {
	return p.velocityX, p.velocityY
}

// MoveAndSlide moves the platform and returns the actual displacement.
// Implements physics.Kinematic interface.
func (p *MovingPlatform) MoveAndSlide(collisionMap *world.CollisionMap, dt float64) (dx, dy float64) {
	// Step 1: If waiting at endpoint, decrement timer and return no movement
	if p.waitTimer > 0 {
		p.waitTimer -= dt
		p.velocityX = 0
		p.velocityY = 0
		return 0, 0
	}

	// Step 2: Calculate direction toward target
	targetX, targetY := p.endX, p.endY
	if !p.goingToEnd {
		targetX, targetY = p.startX, p.startY
	}

	// Direction vector from current position to target
	dirX := targetX - p.body.PosX
	dirY := targetY - p.body.PosY
	dist := math.Sqrt(dirX*dirX + dirY*dirY)

	// Avoid division by zero
	if dist < 0.001 {
		// Already at target, switch direction
		p.switchDirection()
		return 0, 0
	}

	// Normalize direction
	dirX /= dist
	dirY /= dist

	// Step 3: Calculate velocity = direction * speed
	p.velocityX = dirX * p.speed
	p.velocityY = dirY * p.speed

	// Step 4: Calculate potential movement
	potentialDx := p.velocityX * dt
	potentialDy := p.velocityY * dt

	// Step 5: Check if we'd overshoot target
	potentialDist := math.Sqrt(potentialDx*potentialDx + potentialDy*potentialDy)
	if potentialDist >= dist {
		// Snap to target and switch direction
		dx = targetX - p.body.PosX
		dy = targetY - p.body.PosY
		p.body.PosX = targetX
		p.body.PosY = targetY
		p.switchDirection()
		return dx, dy
	}

	// Step 6: Apply movement (no tile collision for platforms in v1)
	p.body.PosX += potentialDx
	p.body.PosY += potentialDy

	// Step 7: Return actual movement delta
	return potentialDx, potentialDy
}

// switchDirection reverses the platform's movement direction and starts wait timer.
func (p *MovingPlatform) switchDirection() {
	p.goingToEnd = !p.goingToEnd
	p.waitTimer = p.waitTime
	p.velocityX = 0
	p.velocityY = 0
}

// Update is called each frame for per-frame updates.
// Currently empty as movement is handled by MoveAndSlide.
func (p *MovingPlatform) Update(dt float64) {
	// Movement is handled by MoveAndSlide in the physics update loop
}

// Draw renders the platform with camera offset.
// Deprecated: Use DrawWithContext for new implementations.
func (p *MovingPlatform) Draw(screen *ebiten.Image, camX, camY float64) {
	x := p.body.PosX - camX
	y := p.body.PosY - camY

	// Draw platform with a distinct purple/blue color
	platformColor := color.RGBA{128, 64, 192, 255} // Purple
	ebitenutil.DrawRect(screen, x, y, p.body.W, p.body.H, platformColor)

	// Draw border for visibility
	borderColor := color.RGBA{80, 40, 140, 255}
	ebitenutil.DrawRect(screen, x, y, p.body.W, 2, borderColor)
	ebitenutil.DrawRect(screen, x, y+p.body.H-2, p.body.W, 2, borderColor)
	ebitenutil.DrawRect(screen, x, y, 2, p.body.H, borderColor)
	ebitenutil.DrawRect(screen, x+p.body.W-2, y, 2, p.body.H, borderColor)
}

// DrawWithContext renders the platform using a RenderContext.
func (p *MovingPlatform) DrawWithContext(screen *ebiten.Image, ctx *world.RenderContext) {
	// Convert world coordinates to screen coordinates
	x, y := ctx.WorldToScreen(p.body.PosX, p.body.PosY)

	// Draw platform with a distinct purple/blue color
	platformColor := color.RGBA{128, 64, 192, 255} // Purple
	ebitenutil.DrawRect(screen, x, y, p.body.W, p.body.H, platformColor)

	// Draw border for visibility
	borderColor := color.RGBA{80, 40, 140, 255}
	ebitenutil.DrawRect(screen, x, y, p.body.W, 2, borderColor)
	ebitenutil.DrawRect(screen, x, y+p.body.H-2, p.body.W, 2, borderColor)
	ebitenutil.DrawRect(screen, x, y, 2, p.body.H, borderColor)
	ebitenutil.DrawRect(screen, x+p.body.W-2, y, 2, p.body.H, borderColor)
}

// DrawDebug renders debug visualization for the platform.
// Draws the platform path as a line from start to end point,
// and draws the platform bounds with a distinct debug color.
func (p *MovingPlatform) DrawDebug(screen *ebiten.Image, ctx *world.RenderContext) {
	// Debug colors
	pathColor := color.RGBA{255, 255, 0, 255}   // Yellow for path line
	boundsColor := color.RGBA{0, 255, 255, 255} // Cyan for bounds

	// Convert world coordinates to screen coordinates
	startScreenX, startScreenY := ctx.WorldToScreen(p.startX, p.startY)
	endScreenX, endScreenY := ctx.WorldToScreen(p.endX, p.endY)
	platformScreenX, platformScreenY := ctx.WorldToScreen(p.body.PosX, p.body.PosY)

	// Draw path line from start to end (center of platform positions)
	startCenterX := startScreenX + p.body.W/2
	startCenterY := startScreenY + p.body.H/2
	endCenterX := endScreenX + p.body.W/2
	endCenterY := endScreenY + p.body.H/2
	ebitenutil.DrawLine(screen, startCenterX, startCenterY, endCenterX, endCenterY, pathColor)

	// Draw small markers at start and end points
	markerSize := 4.0
	ebitenutil.DrawRect(screen, startCenterX-markerSize/2, startCenterY-markerSize/2, markerSize, markerSize, pathColor)
	ebitenutil.DrawRect(screen, endCenterX-markerSize/2, endCenterY-markerSize/2, markerSize, markerSize, pathColor)

	// Draw platform bounds (border only)
	borderWidth := 2.0
	ebitenutil.DrawRect(screen, platformScreenX, platformScreenY, p.body.W, borderWidth, boundsColor)
	ebitenutil.DrawRect(screen, platformScreenX, platformScreenY+p.body.H-borderWidth, p.body.W, borderWidth, boundsColor)
	ebitenutil.DrawRect(screen, platformScreenX, platformScreenY, borderWidth, p.body.H, boundsColor)
	ebitenutil.DrawRect(screen, platformScreenX+p.body.W-borderWidth, platformScreenY, borderWidth, p.body.H, boundsColor)
}

// Bounds returns the AABB for the platform.
// Implements Entity interface.
func (p *MovingPlatform) Bounds() physics.AABB {
	return p.body.AABB()
}

// GetDebugInfo returns a string with debug information about the platform.
func (p *MovingPlatform) GetDebugInfo() string {
	direction := "A→B"
	if !p.goingToEnd {
		direction = "B→A"
	}
	return string("Platform[" + p.id + "] " + direction +
		" vel=(" + string(int(p.velocityX)) + "," + string(int(p.velocityY)) + ")" +
		" wait=" + string(int(p.waitTimer*1000)) + "ms")
}

// SetWaitTime sets the time to wait at endpoints.
func (p *MovingPlatform) SetWaitTime(seconds float64) {
	p.waitTime = seconds
}

// SetPushPlayer sets whether the platform should push the player sideways.
func (p *MovingPlatform) SetPushPlayer(push bool) {
	p.pushPlayer = push
}

// PushPlayer returns whether the platform pushes the player sideways.
func (p *MovingPlatform) PushPlayer() bool {
	return p.pushPlayer
}

// GetID returns the platform's identifier.
func (p *MovingPlatform) GetID() string {
	return p.id
}
