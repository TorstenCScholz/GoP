// Package time provides fixed timestep utilities for deterministic physics.
package time

import (
	"time"
)

const (
	// TargetFPS is the target updates per second for physics.
	TargetFPS = 60

	// FixedTick is the duration of each fixed physics step.
	FixedTick = time.Second / TargetFPS

	// MaxFrameTime prevents the spiral of death by clamping frame time.
	// If a frame takes longer than this, extra time is discarded.
	MaxFrameTime = 250 * time.Millisecond
)

// Timestep implements an accumulator pattern for fixed timestep updates.
// This ensures physics runs at a consistent rate regardless of frame rate.
type Timestep struct {
	// accumulator stores accumulated time waiting to be consumed by fixed updates.
	accumulator time.Duration

	// tick is the fixed duration of each physics step.
	tick time.Duration

	// maxFrameTime is the maximum time that can be added per frame.
	maxFrameTime time.Duration

	// stepsThisFrame tracks how many physics steps occurred in the current frame.
	stepsThisFrame int

	// totalTicks tracks total physics steps for debugging.
	totalTicks int
}

// NewTimestep creates a new fixed timestep controller with default settings.
func NewTimestep() *Timestep {
	return &Timestep{
		accumulator:   0,
		tick:          FixedTick,
		maxFrameTime:  MaxFrameTime,
		stepsThisFrame: 0,
		totalTicks:    0,
	}
}

// AddFrameTime adds elapsed time to the accumulator.
// Call this once per frame with the frame delta time.
// The time is clamped by maxFrameTime to prevent spiral of death.
func (t *Timestep) AddFrameTime(dt time.Duration) {
	// Reset step counter for this frame
	t.stepsThisFrame = 0

	// Clamp to max frame time to prevent spiral of death
	if dt > t.maxFrameTime {
		dt = t.maxFrameTime
	}
	t.accumulator += dt
}

// ShouldUpdate returns true if a fixed update should run.
// Call this in a loop: for t.ShouldUpdate() { /* physics step */ t.ConsumeTick() }
func (t *Timestep) ShouldUpdate() bool {
	return t.accumulator >= t.tick
}

// ConsumeTick consumes one fixed tick from the accumulator.
// Call this after each physics update.
func (t *Timestep) ConsumeTick() {
	t.accumulator -= t.tick
	t.totalTicks++
	t.stepsThisFrame++
}

// Alpha returns the interpolation factor for rendering.
// This is used to interpolate between the previous and current physics state.
// Value is in range [0.0, 1.0), where 0 = just stepped, ~1 = about to step.
func (t *Timestep) Alpha() float64 {
	return float64(t.accumulator) / float64(t.tick)
}

// ResetFrame resets the per-frame step counter.
// This is called automatically by AddFrameTime, but can be called manually if needed.
func (t *Timestep) ResetFrame() {
	t.stepsThisFrame = 0
}

// TickDuration returns the fixed tick duration.
func (t *Timestep) TickDuration() time.Duration {
	return t.tick
}

// TickDurationSeconds returns the fixed tick duration in seconds (as float64).
// This is useful for physics calculations that work in seconds.
func (t *Timestep) TickDurationSeconds() float64 {
	return float64(t.tick) / float64(time.Second)
}

// StepsThisFrame returns the number of physics steps taken this frame.
func (t *Timestep) StepsThisFrame() int {
	return t.stepsThisFrame
}

// TotalTicks returns the total number of physics steps since creation.
func (t *Timestep) TotalTicks() int {
	return t.totalTicks
}

// Accumulator returns the current accumulator value for debugging.
func (t *Timestep) Accumulator() time.Duration {
	return t.accumulator
}
