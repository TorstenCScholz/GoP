package physics

import (
	"github.com/torsten/GoP/internal/input"
	"github.com/torsten/GoP/internal/world"
)

// PlayerController handles player input and physics for a platformer character.
type PlayerController struct {
	Body      *Body
	Speed     float64 // Horizontal movement speed (pixels/second)
	JumpForce float64 // Jump velocity (negative = up)
	Gravity   float64 // Gravity acceleration (pixels/second²)
	MaxFall   float64 // Maximum fall speed (pixels/second)
}

// NewPlayerController creates a new player controller with default values.
// Default values: Speed=90, JumpForce=-300, Gravity=800, MaxFall=400
func NewPlayerController(body *Body) *PlayerController {
	return &PlayerController{
		Body:      body,
		Speed:     90.0,
		JumpForce: -300.0,
		Gravity:   800.0,
		MaxFall:   400.0,
	}
}

// Update processes input and moves the player.
// It reads left/right input, applies gravity, handles jumping, and resolves collisions.
func (c *PlayerController) Update(inp *input.Input, collisionMap *world.CollisionMap, resolver *CollisionResolver, dt float64) {
	// Horizontal movement - read left/right input and set horizontal velocity
	c.Body.VelX = 0
	
	// DEBUG: Log input state
	leftPressed := inp.Pressed(input.ActionMoveLeft)
	rightPressed := inp.Pressed(input.ActionMoveRight)
	jumpPressed := inp.JustPressed(input.ActionJump)
	
	if leftPressed || rightPressed || jumpPressed {
		println("[DEBUG] Input: left=", leftPressed, " right=", rightPressed, " jump=", jumpPressed)
	}
	
	if leftPressed {
		c.Body.VelX = -c.Speed
		println("[DEBUG] Set VelX to", -c.Speed, "(moving left)")
	}
	if rightPressed {
		c.Body.VelX = c.Speed
		println("[DEBUG] Set VelX to", c.Speed, "(moving right)")
	}

	// Jump - only when grounded
	if inp.JustPressed(input.ActionJump) && c.Body.OnGround {
		c.Body.VelY = c.JumpForce
		c.Body.OnGround = false
	}

	// Apply gravity to vertical velocity
	c.Body.VelY += c.Gravity * dt

	// Clamp fall speed to maximum
	if c.Body.VelY > c.MaxFall {
		c.Body.VelY = c.MaxFall
	}

	// Calculate movement delta based on velocity and delta time
	dx := c.Body.VelX * dt
	dy := c.Body.VelY * dt
	
	if dx != 0 || dy != 0 {
		println("[DEBUG] Before resolve: PosX=", c.Body.PosX, " PosY=", c.Body.PosY, " dx=", dx, " dy=", dy)
	}

	// Resolve collisions and move the body
	resolver.Resolve(c.Body, collisionMap, dx, dy)
	
	if dx != 0 || dy != 0 {
		println("[DEBUG] After resolve: PosX=", c.Body.PosX, " PosY=", c.Body.PosY)
	}
}
