package gfx

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/torsten/GoP/internal/assets"
)

// Animation represents a sequence of frames with timing information.
// It stores the frame images and duration for playback.
type Animation struct {
	Frames        []*ebiten.Image
	FrameDuration time.Duration
	Loop          bool
}

// NewAnimation creates a new animation with the given frames and frame duration.
func NewAnimation(frames []*ebiten.Image, frameDuration time.Duration) *Animation {
	return &Animation{
		Frames:        frames,
		FrameDuration: frameDuration,
		Loop:          true,
	}
}

// NewAnimationFromSheet creates a new animation from a spritesheet.
// This is a convenience constructor that extracts frames from the sheet.
func NewAnimationFromSheet(sheet *assets.Sheet, frameDuration time.Duration) *Animation {
	return &Animation{
		Frames:        sheet.Frames(),
		FrameDuration: frameDuration,
		Loop:          true,
	}
}

// Length returns the total duration of the animation.
// Returns 0 if there are no frames.
func (a *Animation) Length() time.Duration {
	if len(a.Frames) == 0 {
		return 0
	}
	return a.FrameDuration * time.Duration(len(a.Frames))
}

// FrameCount returns the number of frames in the animation.
func (a *Animation) FrameCount() int {
	return len(a.Frames)
}
