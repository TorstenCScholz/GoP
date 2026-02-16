package physics

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
