package physics

import (
	"math"

	"github.com/torsten/GoP/internal/world"
)

// CollisionResolver handles collision resolution against a tile grid.
// It uses axis-separated resolution for stable platformer physics.
type CollisionResolver struct {
	TileW, TileH int // tile dimensions for snapping
}

// NewCollisionResolver creates a new collision resolver with the given tile dimensions.
func NewCollisionResolver(tileW, tileH int) *CollisionResolver {
	return &CollisionResolver{
		TileW: tileW,
		TileH: tileH,
	}
}

// Resolve moves the body and resolves collisions. Returns the actual movement.
// The resolution algorithm uses axis-separated resolution:
// 1. Apply X movement first
// 2. Check if body overlaps solid tiles after X movement
// 3. If collision, snap to tile edge and zero X velocity
// 4. Apply Y movement
// 5. Check if body overlaps solid tiles after Y movement
// 6. If collision from above (moving down): snap to tile top, set OnGround=true, zero Y velocity
// 7. If collision from below (moving up): snap to tile bottom, zero Y velocity
// 8. If no Y collision: set OnGround=false
func (r *CollisionResolver) Resolve(body *Body, collisionMap *world.CollisionMap, dx, dy float64) (actualDx, actualDy float64) {
	actualDx = 0
	actualDy = 0

	// Step 1: Resolve X-axis collision
	if dx != 0 {
		body.PosX += dx
		actualDx = dx

		// Check for collision after X movement
		if collisionMap.OverlapsSolid(body.PosX, body.PosY, body.W, body.H) {
			println("[DEBUG RESOLVE] X collision detected! dx=", dx, " PosX before snap=", body.PosX)
			if dx > 0 {
				// Moving right - snap to left edge of colliding tile
				snapPos := r.snapLeft(body, collisionMap)
				body.PosX = snapPos - body.W
				println("[DEBUG RESOLVE] Moving right: snapLeft=", snapPos, " new PosX=", body.PosX)
			} else {
				// Moving left - snap to right edge of colliding tile
				snapPos := r.snapRight(body, collisionMap)
				body.PosX = snapPos
				println("[DEBUG RESOLVE] Moving left: snapRight=", snapPos, " new PosX=", body.PosX)
			}
			body.VelX = 0
			actualDx = 0
		}
	}

	// Step 2: Resolve Y-axis collision
	if dy != 0 {
		body.PosY += dy
		actualDy = dy

		// Reset ground state before checking
		wasOnGround := body.OnGround
		body.OnGround = false

		// Check for collision after Y movement
		if collisionMap.OverlapsSolid(body.PosX, body.PosY, body.W, body.H) {
			println("[DEBUG RESOLVE] Y collision detected! dy=", dy, " PosY before snap=", body.PosY)
			if dy > 0 {
				// Moving down - snap to top edge of colliding tile (landing)
				snapPos := r.snapTop(body, collisionMap)
				body.PosY = snapPos - body.H
				body.OnGround = true
				println("[DEBUG RESOLVE] Falling: snapTop=", snapPos, " body.H=", body.H, " new PosY=", body.PosY)
			} else {
				// Moving up - snap to bottom edge of colliding tile (ceiling)
				snapPos := r.snapBottom(body, collisionMap)
				body.PosY = snapPos
				println("[DEBUG RESOLVE] Jumping: snapBottom=", snapPos, " new PosY=", body.PosY)
			}
			body.VelY = 0
			actualDy = 0
		}
		// If no collision and we weren't on ground before, OnGround stays false
		_ = wasOnGround // unused, but kept for clarity
	}

	return actualDx, actualDy
}

// snapLeft returns the left edge X of the leftmost solid tile overlapping the body.
func (r *CollisionResolver) snapLeft(body *Body, collisionMap *world.CollisionMap) float64 {
	tiles := collisionMap.GetOverlappingTiles(body.PosX, body.PosY, body.W, body.H)
	minX := math.MaxFloat64

	for _, t := range tiles {
		if collisionMap.IsSolidAtTile(t.X, t.Y) {
			tileLeft := float64(t.X * r.TileW)
			if tileLeft < minX {
				minX = tileLeft
			}
		}
	}

	if minX == math.MaxFloat64 {
		return body.PosX
	}
	return minX
}

// snapRight returns the right edge X of the rightmost solid tile overlapping the body.
func (r *CollisionResolver) snapRight(body *Body, collisionMap *world.CollisionMap) float64 {
	tiles := collisionMap.GetOverlappingTiles(body.PosX, body.PosY, body.W, body.H)
	maxX := -math.MaxFloat64

	for _, t := range tiles {
		if collisionMap.IsSolidAtTile(t.X, t.Y) {
			tileRight := float64((t.X + 1) * r.TileW)
			if tileRight > maxX {
				maxX = tileRight
			}
		}
	}

	if maxX == -math.MaxFloat64 {
		return body.PosX
	}
	return maxX
}

// snapTop returns the top edge Y of the topmost solid tile overlapping the body.
func (r *CollisionResolver) snapTop(body *Body, collisionMap *world.CollisionMap) float64 {
	tiles := collisionMap.GetOverlappingTiles(body.PosX, body.PosY, body.W, body.H)
	minY := math.MaxFloat64

	println("[DEBUG snapTop] Checking", len(tiles), "tiles for body at (", body.PosX, ",", body.PosY, ") size (", body.W, "x", body.H, ")")
	for _, t := range tiles {
		solid := collisionMap.IsSolidAtTile(t.X, t.Y)
		tileTop := float64(t.Y * r.TileH)
		println("[DEBUG snapTop] Tile (", t.X, ",", t.Y, ") top=", tileTop, " solid=", solid)
		if solid {
			if tileTop < minY {
				minY = tileTop
			}
		}
	}

	if minY == math.MaxFloat64 {
		return body.PosY
	}
	return minY
}

// snapBottom returns the bottom edge Y of the bottommost solid tile overlapping the body.
func (r *CollisionResolver) snapBottom(body *Body, collisionMap *world.CollisionMap) float64 {
	tiles := collisionMap.GetOverlappingTiles(body.PosX, body.PosY, body.W, body.H)
	maxY := -math.MaxFloat64

	for _, t := range tiles {
		if collisionMap.IsSolidAtTile(t.X, t.Y) {
			tileBottom := float64((t.Y + 1) * r.TileH)
			if tileBottom > maxY {
				maxY = tileBottom
			}
		}
	}

	if maxY == -math.MaxFloat64 {
		return body.PosY
	}
	return maxY
}
