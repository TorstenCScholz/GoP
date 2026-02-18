// Package game provides game feel tuning parameters.
package game

import (
	"time"
)

// Tuning contains all tweakable parameters for player movement feel.
// All velocity values are in pixels/second, acceleration in pixels/second².
type Tuning struct {
	// Horizontal movement parameters
	Horizontal HorizontalTuning

	// Jump parameters
	Jump JumpTuning

	// Gravity parameters
	Gravity GravityTuning
}

// HorizontalTuning controls left/right movement feel.
type HorizontalTuning struct {
	// Acceleration when starting to move (pixels/second²).
	// Higher = snappier start, Lower = slippery start.
	Acceleration float64

	// Deceleration when releasing input (pixels/second²).
	// Higher = quick stop, Lower = sliding stop.
	Deceleration float64

	// Maximum horizontal speed (pixels/second).
	MaxSpeed float64

	// Friction coefficient when grounded (0-1).
	// Applied each frame: velX *= (1 - friction).
	// Higher = more friction, 0 = no friction.
	Friction float64

	// Air control multiplier (0-1).
	// Applied to acceleration when in air.
	// 1 = full air control, 0 = no air control.
	AirControl float64
}

// JumpTuning controls jump behavior.
type JumpTuning struct {
	// Initial jump velocity (pixels/second, negative = up).
	// Higher absolute value = higher jump.
	Velocity float64

	// CoyoteTime is the grace period after leaving ground where jump still works.
	CoyoteTime time.Duration

	// BufferTime is the pre-input window before landing.
	// Jump input is remembered for this duration before landing.
	BufferTime time.Duration

	// VariableHeight enables variable jump height based on button hold duration.
	VariableHeight bool

	// EarlyReleaseMult is the gravity multiplier when jump button is released early.
	// Higher = faster fall when button released early.
	EarlyReleaseMult float64
}

// GravityTuning controls falling behavior.
type GravityTuning struct {
	// Base gravity acceleration (pixels/second²).
	// Higher = faster fall.
	Base float64

	// FallMult is the multiplier applied when falling (velocity > 0).
	// > 1 = faster fall, creates arc feel.
	// 1 = symmetric jump arc.
	FallMult float64

	// MaxFall is the maximum fall speed (pixels/second).
	MaxFall float64
}

// DefaultTuning returns tuning parameters with good default feel.
// These values are based on common platformer conventions and can be tweaked.
func DefaultTuning() Tuning {
	return Tuning{
		Horizontal: HorizontalTuning{
			Acceleration: 1200.0, // Fast acceleration
			Deceleration: 800.0,  // Quick stop
			MaxSpeed:     150.0,  // Reasonable max speed
			Friction:     0.15,   // Some ground friction
			AirControl:   0.6,    // Reduced air control
		},
		Jump: JumpTuning{
			Velocity:         -280.0,            // Good jump height (negative = up)
			CoyoteTime:       100 * time.Millisecond, // 100ms coyote time
			BufferTime:       100 * time.Millisecond, // 100ms jump buffer
			VariableHeight:   true,              // Enable variable jump height
			EarlyReleaseMult: 2.5,               // Fall faster when released
		},
		Gravity: GravityTuning{
			Base:    900.0, // Moderate gravity
			FallMult: 1.5,  // Faster falling
			MaxFall: 400.0, // Terminal velocity
		},
	}
}
