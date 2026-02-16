// Package physics provides collision detection and resolution for platformer games.
package physics

// AABB represents an Axis-Aligned Bounding Box for collision detection.
type AABB struct {
	X, Y float64 // Top-left position
	W, H float64 // Width and height
}

// Left returns the left edge X coordinate.
func (a AABB) Left() float64 {
	return a.X
}

// Right returns the right edge X coordinate.
func (a AABB) Right() float64 {
	return a.X + a.W
}

// Top returns the top edge Y coordinate.
func (a AABB) Top() float64 {
	return a.Y
}

// Bottom returns the bottom edge Y coordinate.
func (a AABB) Bottom() float64 {
	return a.Y + a.H
}

// Intersects checks if two AABBs overlap.
func (a AABB) Intersects(other AABB) bool {
	// No intersection if one AABB is completely to the left/right or above/below the other
	if a.Right() <= other.Left() || other.Right() <= a.Left() {
		return false
	}
	if a.Bottom() <= other.Top() || other.Bottom() <= a.Top() {
		return false
	}
	return true
}
