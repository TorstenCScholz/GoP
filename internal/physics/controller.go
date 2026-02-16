package physics

import (
	"time"

	"github.com/torsten/GoP/internal/game"
	"github.com/torsten/GoP/internal/input"
	"github.com/torsten/GoP/internal/world"
)

// Collision represents collision information returned by the collision function.
type Collision struct {
	// Tile coordinates
	TileX, TileY int

	// Normal direction to resolve collision
	NormalX, NormalY float64
}

// PlayerState tracks extended player state for feel mechanics.
type PlayerState struct {
	// Coyote time tracking
	TimeSinceGrounded time.Duration // Time since last OnGround = true
	WasGrounded       bool          // Previous frame's ground state

	// Jump buffer tracking
	JumpBufferTime time.Duration // Time remaining in jump buffer
	JumpBuffered   bool          // True if jump is buffered

	// Variable jump tracking
	JumpHeldTime time.Duration // How long jump has been held
	IsJumping    bool          // True during active jump
	JumpReleased bool          // True if jump released early
}

// Controller handles player input and physics with feel mechanics.
type Controller struct {
	Body   *Body
	Tuning game.Tuning
	State  PlayerState
}

// NewController creates a controller with tuning parameters.
func NewController(body *Body, tuning game.Tuning) *Controller {
	return &Controller{
		Body:   body,
		Tuning: tuning,
	}
}

// PlayerController is an alias for Controller for backward compatibility.
// Deprecated: Use Controller instead.
type PlayerController = Controller

// NewPlayerController creates a controller with default tuning for backward compatibility.
// Deprecated: Use NewController instead.
func NewPlayerController(body *Body) *Controller {
	return &Controller{
		Body:   body,
		Tuning: game.DefaultTuning(),
	}
}

// Update processes input and updates physics with feel mechanics.
// This method is provided for backward compatibility with the existing codebase.
// It uses the CollisionResolver and CollisionMap directly instead of a collision function.
// Deprecated: Use FixedUpdate with a collision function for more flexibility.
func (c *Controller) Update(inp *input.Input, collisionMap *world.CollisionMap, resolver *CollisionResolver, dt float64) {
	dtDuration := time.Duration(dt * float64(time.Second))

	// Update state tracking
	c.updateStateTracking(dtDuration, inp)

	// Process horizontal movement with acceleration
	c.updateHorizontal(inp, dt)

	// Process jump mechanics (coyote time, buffer, variable height)
	c.updateJump(inp, dt)

	// Apply gravity
	c.applyGravity(dt)

	// Calculate movement delta based on velocity and delta time
	dx := c.Body.VelX * dt
	dy := c.Body.VelY * dt

	// Resolve collisions using the resolver
	resolver.Resolve(c.Body, collisionMap, dx, dy)

	// Update ground state tracking after collision
	c.State.WasGrounded = c.Body.OnGround
	if c.Body.OnGround {
		c.State.TimeSinceGrounded = 0
		c.State.IsJumping = false
		c.State.JumpReleased = false
	}
}

// FixedUpdate processes input and updates physics with feel mechanics.
// This method is designed to be called at a fixed timestep (e.g., 60Hz).
// The collisionFunc should check for collisions at the given AABB and return
// collision information for resolution.
func (c *Controller) FixedUpdate(dt time.Duration, inp *input.Input, collisionFunc func(AABB) []Collision) {
	dtSeconds := dt.Seconds()

	// Update state tracking
	c.updateStateTracking(dt, inp)

	// Process horizontal movement with acceleration
	c.updateHorizontal(inp, dtSeconds)

	// Process jump mechanics (coyote time, buffer, variable height)
	c.updateJump(inp, dtSeconds)

	// Apply gravity
	c.applyGravity(dtSeconds)

	// Calculate movement delta based on velocity and delta time
	dx := c.Body.VelX * dtSeconds
	dy := c.Body.VelY * dtSeconds

	// Resolve collisions using the provided collision function
	c.resolveCollisions(dx, dy, collisionFunc)

	// Update ground state tracking after collision
	c.State.WasGrounded = c.Body.OnGround
	if c.Body.OnGround {
		c.State.TimeSinceGrounded = 0
		c.State.IsJumping = false
		c.State.JumpReleased = false
	}
}

// updateStateTracking updates time-based state counters.
func (c *Controller) updateStateTracking(dt time.Duration, inp *input.Input) {
	// Track time since last grounded (for coyote time)
	if !c.Body.OnGround {
		c.State.TimeSinceGrounded += dt
	}

	// Update jump buffer countdown
	if c.State.JumpBufferTime > 0 {
		c.State.JumpBufferTime -= dt
		if c.State.JumpBufferTime <= 0 {
			c.State.JumpBuffered = false
		}
	}

	// Track jump hold time for variable jump
	if c.State.IsJumping && inp.Pressed(input.ActionJump) {
		c.State.JumpHeldTime += dt
	}
}

// updateHorizontal handles horizontal movement with acceleration and friction.
func (c *Controller) updateHorizontal(inp *input.Input, dt float64) {
	tuning := c.Tuning.Horizontal

	// Determine input direction
	var inputDir float64
	if inp.Pressed(input.ActionMoveLeft) {
		inputDir = -1
	} else if inp.Pressed(input.ActionMoveRight) {
		inputDir = 1
	}

	// Apply acceleration or deceleration
	if inputDir != 0 {
		// Calculate acceleration (reduced in air)
		accel := tuning.Acceleration
		if !c.Body.OnGround {
			accel *= tuning.AirControl
		}

		// Apply acceleration in input direction
		c.Body.VelX += inputDir * accel * dt

		// Clamp to max speed
		if c.Body.VelX > tuning.MaxSpeed {
			c.Body.VelX = tuning.MaxSpeed
		} else if c.Body.VelX < -tuning.MaxSpeed {
			c.Body.VelX = -tuning.MaxSpeed
		}
	} else {
		// No input - apply deceleration or friction
		if c.Body.OnGround {
			// Ground friction
			friction := tuning.Friction * 60 * dt // Normalize to per-frame
			if friction > 1 {
				friction = 1
			}
			c.Body.VelX *= (1 - friction)

			// Deceleration for more responsive stops
			if c.Body.VelX > 0 {
				c.Body.VelX -= tuning.Deceleration * dt
				if c.Body.VelX < 0 {
					c.Body.VelX = 0
				}
			} else if c.Body.VelX < 0 {
				c.Body.VelX += tuning.Deceleration * dt
				if c.Body.VelX > 0 {
					c.Body.VelX = 0
				}
			}
		} else {
			// Air deceleration (reduced)
			airDecel := tuning.Deceleration * tuning.AirControl
			if c.Body.VelX > 0 {
				c.Body.VelX -= airDecel * dt
				if c.Body.VelX < 0 {
					c.Body.VelX = 0
				}
			} else if c.Body.VelX < 0 {
				c.Body.VelX += airDecel * dt
				if c.Body.VelX > 0 {
					c.Body.VelX = 0
				}
			}
		}
	}
}

// updateJump handles jump mechanics including coyote time and buffering.
func (c *Controller) updateJump(inp *input.Input, dt float64) {
	jumpTuning := c.Tuning.Jump

	// Buffer jump input
	if inp.JustPressed(input.ActionJump) {
		c.State.JumpBuffered = true
		c.State.JumpBufferTime = jumpTuning.BufferTime
	}

	// Check if we can jump (grounded or within coyote time)
	canJump := c.Body.OnGround || c.State.TimeSinceGrounded <= jumpTuning.CoyoteTime

	// Execute buffered jump if possible
	if c.State.JumpBuffered && canJump {
		c.executeJump()
		c.State.JumpBuffered = false
		c.State.JumpBufferTime = 0
	}

	// Variable jump height - apply extra gravity if released early
	if jumpTuning.VariableHeight && c.State.IsJumping {
		if !inp.Pressed(input.ActionJump) {
			c.State.JumpReleased = true
		}
	}
}

// executeJump performs the actual jump.
func (c *Controller) executeJump() {
	c.Body.VelY = c.Tuning.Jump.Velocity
	c.Body.OnGround = false
	c.State.IsJumping = true
	c.State.JumpHeldTime = 0
	c.State.TimeSinceGrounded = c.Tuning.Jump.CoyoteTime + time.Millisecond // Prevent double-jump
}

// applyGravity applies gravity with variable jump height support.
func (c *Controller) applyGravity(dt float64) {
	gravityTuning := c.Tuning.Gravity
	jumpTuning := c.Tuning.Jump

	// Base gravity
	gravity := gravityTuning.Base

	// Apply falling multiplier when moving down
	if c.Body.VelY > 0 {
		gravity *= gravityTuning.FallMult
	}

	// Apply early release multiplier for variable jump
	if jumpTuning.VariableHeight && c.State.JumpReleased && c.Body.VelY < 0 {
		gravity *= jumpTuning.EarlyReleaseMult
	}

	c.Body.VelY += gravity * dt

	// Clamp to max fall speed
	if c.Body.VelY > gravityTuning.MaxFall {
		c.Body.VelY = gravityTuning.MaxFall
	}
}

// resolveCollisions handles collision resolution using axis-separated resolution.
// This maintains compatibility with the existing collision system.
func (c *Controller) resolveCollisions(dx, dy float64, collisionFunc func(AABB) []Collision) {
	// Resolve X-axis collision
	if dx != 0 {
		c.Body.PosX += dx

		// Check for collision after X movement
		collisions := collisionFunc(c.Body.AABB())
		if len(collisions) > 0 {
			// Find the primary collision direction
			for _, col := range collisions {
				if col.NormalX != 0 {
					if dx > 0 && col.NormalX < 0 {
						// Moving right, hit left side of tile
						c.Body.PosX = float64(col.TileX*16) - c.Body.W // Assuming 16px tiles
					} else if dx < 0 && col.NormalX > 0 {
						// Moving left, hit right side of tile
						c.Body.PosX = float64((col.TileX + 1) * 16)
					}
					c.Body.VelX = 0
					break
				}
			}
		}
	}

	// Resolve Y-axis collision
	if dy != 0 {
		c.Body.PosY += dy

		// Reset ground state before checking
		c.Body.OnGround = false

		// Check for collision after Y movement
		collisions := collisionFunc(c.Body.AABB())
		if len(collisions) > 0 {
			for _, col := range collisions {
				if col.NormalY != 0 {
					if dy > 0 && col.NormalY < 0 {
						// Moving down, hit top of tile (landing)
						c.Body.PosY = float64(col.TileY*16) - c.Body.H
						c.Body.OnGround = true
					} else if dy < 0 && col.NormalY > 0 {
						// Moving up, hit bottom of tile (ceiling)
						c.Body.PosY = float64((col.TileY + 1) * 16)
					}
					c.Body.VelY = 0
					break
				}
			}
		}
	}
}
