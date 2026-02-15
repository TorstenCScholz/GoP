package assets

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Sheet represents a spritesheet with sliced frames.
// Frames are extracted using SubImage, sharing memory with the original image.
type Sheet struct {
	image  *ebiten.Image
	frames []*ebiten.Image
}

// NewSheet creates a new spritesheet from an image with uniform frame size.
// Frames are extracted left-to-right, top-to-bottom.
func NewSheet(img *ebiten.Image, frameWidth, frameHeight int) *Sheet {
	bounds := img.Bounds()
	sheetW, sheetH := bounds.Dx(), bounds.Dy()

	cols := sheetW / frameWidth
	rows := sheetH / frameHeight

	frames := make([]*ebiten.Image, 0, cols*rows)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			x := col * frameWidth
			y := row * frameHeight
			// Use SubImage to slice frames - this shares memory with the original
			frame := img.SubImage(image.Rect(x, y, x+frameWidth, y+frameHeight)).(*ebiten.Image)
			frames = append(frames, frame)
		}
	}

	return &Sheet{
		image:  img,
		frames: frames,
	}
}

// Frame returns the frame at the given index.
// Returns nil if the index is out of bounds.
func (s *Sheet) Frame(index int) *ebiten.Image {
	if index < 0 || index >= len(s.frames) {
		return nil
	}
	return s.frames[index]
}

// Frames returns all frames as a slice.
func (s *Sheet) Frames() []*ebiten.Image {
	return s.frames
}

// FrameCount returns the total number of frames.
func (s *Sheet) FrameCount() int {
	return len(s.frames)
}
