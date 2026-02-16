package physics

import "github.com/torsten/GoP/internal/world"

// SolidEntity represents an entity with a physical body that can collide
type SolidEntity interface {
	GetBody() *Body
	IsActive() bool
}

// Kinematic represents a solid entity that moves itself and can carry the player
type Kinematic interface {
	SolidEntity
	Velocity() (vx, vy float64)
	MoveAndSlide(collisionMap *world.CollisionMap, dt float64) (dx, dy float64)
}

// Body represents a physics body with position, velocity, and size.
// Position (PosX, PosY) is the top-left corner of the bounding box.
type Body struct {
	PosX, PosY float64 // Position (top-left corner)
	VelX, VelY float64 // Velocity in pixels per second
	W, H       float64 // Size (width and height)
	OnGround   bool    // True if standing on solid ground
}

// AABB returns the axis-aligned bounding box for this body.
func (b *Body) AABB() AABB {
	return AABB{
		X: b.PosX,
		Y: b.PosY,
		W: b.W,
		H: b.H,
	}
}
