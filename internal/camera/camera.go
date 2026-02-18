// Package camera provides smooth camera following for platformers.
// It supports deadzone-based following, smoothing, level bounds clamping,
// and pixel-perfect snapping for clean pixel art rendering.
package camera

// Camera provides smooth following with deadzone and bounds constraints.
// The camera position (X, Y) represents the top-left corner of the viewport
// in world coordinates.
type Camera struct {
	// Position (top-left of viewport in world coordinates)
	X, Y float64

	// Viewport dimensions
	ViewportW, ViewportH int

	// Deadzone (camera stays still within this area)
	// The deadzone is defined relative to the viewport (screen coordinates).
	// When the target moves outside this rectangle, the camera follows.
	DeadzoneX, DeadzoneY, DeadzoneW, DeadzoneH float64

	// Level bounds for clamping
	LevelW, LevelH float64

	// Smoothing (0 = instant, 1 = very slow)
	// Applied as exponential smoothing: pos += (target - pos) * factor * dt
	Smoothing float64

	// Target position for smoothing (world coordinates)
	targetX, targetY float64

	// Pixel-perfect snapping
	// When true, rounds final position to integers to prevent sub-pixel
	// rendering issues for pixel art.
	PixelPerfect bool
}

// NewCamera creates a new camera with the given viewport dimensions.
// Default settings:
//   - Deadzone: 25% width, 40% height, centered in viewport
//   - Smoothing: 0 (instant follow)
//   - PixelPerfect: true
func NewCamera(viewportW, viewportH int) *Camera {
	c := &Camera{
		ViewportW:    viewportW,
		ViewportH:    viewportH,
		Smoothing:    0.0, // Instant by default
		PixelPerfect: true,
	}

	// Default deadzone: 25% width, 40% height, centered
	c.SetDeadzoneCentered(0.25, 0.4)

	return c
}

// SetLevelBounds sets the level dimensions for camera clamping.
// The camera will never show areas outside these bounds.
func (c *Camera) SetLevelBounds(w, h float64) {
	c.LevelW = w
	c.LevelH = h
}

// SetDeadzone sets the deadzone rectangle in screen coordinates (pixels).
// The deadzone defines an area within the viewport where the camera
// won't move even if the target moves.
func (c *Camera) SetDeadzone(x, y, w, h float64) {
	c.DeadzoneX = x
	c.DeadzoneY = y
	c.DeadzoneW = w
	c.DeadzoneH = h
}

// SetDeadzoneCentered sets the deadzone as a percentage of viewport size.
// The deadzone is centered within the viewport.
// widthPct and heightPct should be in range (0, 1].
// Example: 0.25 = 25% of viewport dimension.
func (c *Camera) SetDeadzoneCentered(widthPct, heightPct float64) {
	c.DeadzoneW = float64(c.ViewportW) * widthPct
	c.DeadzoneH = float64(c.ViewportH) * heightPct
	c.DeadzoneX = (float64(c.ViewportW) - c.DeadzoneW) / 2
	c.DeadzoneY = (float64(c.ViewportH) - c.DeadzoneH) / 2
}

// Follow sets the camera target to follow.
// The target is typically the player's center position.
// targetX, targetY: center point of the target in world coordinates
// targetW, targetH: size of the target (used for future enhancements)
func (c *Camera) Follow(targetX, targetY, targetW, targetH float64) {
	// Store target center
	c.targetX = targetX
	c.targetY = targetY
}

// Update updates the camera position based on the target and deadzone.
// dt is the delta time in seconds (used for smoothing).
// Call this once per frame after setting the target with Follow.
func (c *Camera) Update(dt float64) {
	// Calculate target position relative to current camera (screen coordinates)
	relX := c.targetX - c.X
	relY := c.targetY - c.Y

	// Calculate desired camera position (starts at current position)
	desiredX := c.X
	desiredY := c.Y

	// Deadzone boundaries in screen coordinates
	deadzoneLeft := c.DeadzoneX
	deadzoneRight := c.DeadzoneX + c.DeadzoneW
	deadzoneTop := c.DeadzoneY
	deadzoneBottom := c.DeadzoneY + c.DeadzoneH

	// Horizontal deadzone check
	if relX < deadzoneLeft {
		// Target is left of deadzone - move camera left
		// Camera should move so target is at deadzone left edge
		desiredX = c.targetX - deadzoneLeft
	} else if relX > deadzoneRight {
		// Target is right of deadzone - move camera right
		// Camera should move so target is at deadzone right edge
		desiredX = c.targetX - deadzoneRight
	}

	// Vertical deadzone check
	if relY < deadzoneTop {
		// Target is above deadzone - move camera up
		desiredY = c.targetY - deadzoneTop
	} else if relY > deadzoneBottom {
		// Target is below deadzone - move camera down
		desiredY = c.targetY - deadzoneBottom
	}

	// Apply smoothing (exponential smoothing)
	if c.Smoothing > 0 {
		// Calculate smoothing factor per frame
		// Higher smoothing = slower follow
		factor := 1.0 - c.Smoothing
		if factor < 0 {
			factor = 0
		}
		// Apply frame-rate independent smoothing
		// Using exponential decay: pos += (target - pos) * (1 - e^(-rate * dt))
		// Simplified: pos += (target - pos) * factor * dt * 60 (normalized to 60fps)
		smoothFactor := factor * dt * 60
		if smoothFactor > 1 {
			smoothFactor = 1
		}
		c.X += (desiredX - c.X) * smoothFactor
		c.Y += (desiredY - c.Y) * smoothFactor
	} else {
		// Instant follow
		c.X = desiredX
		c.Y = desiredY
	}

	// Clamp to level bounds
	if c.LevelW > 0 && c.LevelH > 0 {
		maxX := c.LevelW - float64(c.ViewportW)
		maxY := c.LevelH - float64(c.ViewportH)

		if maxX < 0 {
			// Level narrower than viewport: center horizontally
			c.X = maxX / 2
		} else {
			if c.X < 0 {
				c.X = 0
			} else if c.X > maxX {
				c.X = maxX
			}
		}

		if maxY < 0 {
			// Level shorter than viewport: center vertically
			c.Y = maxY / 2
		} else {
			if c.Y < 0 {
				c.Y = 0
			} else if c.Y > maxY {
				c.Y = maxY
			}
		}
	}

	// Pixel-perfect snapping
	if c.PixelPerfect {
		c.X = float64(int(c.X))
		c.Y = float64(int(c.Y))
	}
}

// WorldToScreen converts world coordinates to screen coordinates.
// Returns the screen position where a world point should be rendered.
func (c *Camera) WorldToScreen(x, y float64) (screenX, screenY int) {
	return int(x - c.X), int(y - c.Y)
}

// ScreenToWorld converts screen coordinates to world coordinates.
// Returns the world position at a given screen point.
func (c *Camera) ScreenToWorld(x, y int) (worldX, worldY float64) {
	return float64(x) + c.X, float64(y) + c.Y
}

// Bounds returns the visible area in world coordinates.
// Returns (x, y, w, h) where (x, y) is the top-left corner
// and (w, h) is the visible area size.
func (c *Camera) Bounds() (x, y, w, h float64) {
	return c.X, c.Y, float64(c.ViewportW), float64(c.ViewportH)
}

// VisibleTiles returns the visible tile range for rendering.
// tileSize is the size of each tile in pixels.
// Returns (tx1, ty1, tx2, ty2) where:
//   - (tx1, ty1) is the first visible tile (inclusive)
//   - (tx2, ty2) is the last visible tile + 1 (exclusive)
func (c *Camera) VisibleTiles(tileSize int) (tx1, ty1, tx2, ty2 int) {
	tx1 = int(c.X) / tileSize
	ty1 = int(c.Y) / tileSize
	tx2 = (int(c.X) + c.ViewportW + tileSize - 1) / tileSize
	ty2 = (int(c.Y) + c.ViewportH + tileSize - 1) / tileSize
	return
}

// TargetX returns the current target X position (world coordinates).
func (c *Camera) TargetX() float64 {
	return c.targetX
}

// TargetY returns the current target Y position (world coordinates).
func (c *Camera) TargetY() float64 {
	return c.targetY
}

// CenterX returns the center X of the viewport in world coordinates.
func (c *Camera) CenterX() float64 {
	return c.X + float64(c.ViewportW)/2
}

// CenterY returns the center Y of the viewport in world coordinates.
func (c *Camera) CenterY() float64 {
	return c.Y + float64(c.ViewportH)/2
}
