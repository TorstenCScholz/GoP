package gfx

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Animator handles animation playback state and timing.
// It manages the current frame, elapsed time, and play state.
type Animator struct {
	Animation    *Animation
	elapsed      time.Duration
	currentFrame int
	playing      bool
}

// NewAnimator creates a new animator with the given animation.
// The animator starts in a stopped state at frame 0.
func NewAnimator(anim *Animation) *Animator {
	return &Animator{
		Animation:    anim,
		elapsed:      0,
		currentFrame: 0,
		playing:      false,
	}
}

// Play starts or resumes the animation from the current frame.
func (a *Animator) Play() {
	a.playing = true
}

// Stop stops the animation and resets to frame 0.
func (a *Animator) Stop() {
	a.playing = false
	a.currentFrame = 0
	a.elapsed = 0
}

// Pause pauses the animation at the current frame.
func (a *Animator) Pause() {
	a.playing = false
}

// Reset resets the animation to frame 0 without changing the play state.
func (a *Animator) Reset() {
	a.currentFrame = 0
	a.elapsed = 0
}

// SetAnimation changes the current animation and resets to frame 0.
func (a *Animator) SetAnimation(anim *Animation) {
	a.Animation = anim
	a.currentFrame = 0
	a.elapsed = 0
}

// Update advances the animation by the given delta time.
// Handles frame advancement and looping.
func (a *Animator) Update(dt time.Duration) {
	if !a.playing || a.Animation == nil {
		return
	}

	// Handle empty or single frame animations
	frameCount := a.Animation.FrameCount()
	if frameCount <= 1 {
		return
	}

	a.elapsed += dt

	// Advance frames based on elapsed time
	for a.elapsed >= a.Animation.FrameDuration {
		a.elapsed -= a.Animation.FrameDuration
		a.currentFrame++

		// Handle end of animation
		if a.currentFrame >= frameCount {
			if a.Animation.Loop {
				a.currentFrame = a.currentFrame % frameCount
			} else {
				a.currentFrame = frameCount - 1
				a.playing = false
				a.elapsed = 0
				break
			}
		}
	}
}

// CurrentFrame returns the current frame image, or nil if no animation or empty frames.
func (a *Animator) CurrentFrame() *ebiten.Image {
	if a.Animation == nil || len(a.Animation.Frames) == 0 {
		return nil
	}

	// Clamp to valid range
	if a.currentFrame < 0 {
		a.currentFrame = 0
	}
	if a.currentFrame >= len(a.Animation.Frames) {
		a.currentFrame = len(a.Animation.Frames) - 1
	}

	return a.Animation.Frames[a.currentFrame]
}

// CurrentFrameIndex returns the current frame index.
func (a *Animator) CurrentFrameIndex() int {
	return a.currentFrame
}

// IsPlaying returns whether the animation is currently playing.
func (a *Animator) IsPlaying() bool {
	return a.playing
}
