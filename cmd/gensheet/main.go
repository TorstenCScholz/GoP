// Command gensheet generates a test spritesheet PNG for animation testing.
// The spritesheet contains 4 frames of 32x32 pixels each (128x32 total).
package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	// Spritesheet dimensions: 4 frames of 32x32 each
	const (
		frameWidth  = 32
		frameHeight = 32
		numFrames   = 4
		totalWidth  = frameWidth * numFrames
		totalHeight = frameHeight
	)

	// Create a new RGBA image
	img := image.NewRGBA(image.Rect(0, 0, totalWidth, totalHeight))

	// Define a palette-friendly color scheme
	// Each frame will have a different colored circle to show animation
	colors := []color.RGBA{
		{255, 100, 100, 255}, // Frame 1: Red
		{255, 180, 100, 255}, // Frame 2: Orange
		{255, 255, 100, 255}, // Frame 3: Yellow
		{100, 255, 100, 255}, // Frame 4: Green
	}

	// Background is transparent (alpha = 0)
	// No need to fill - RGBA image starts transparent by default

	// Draw a circle in each frame that fills the entire 32x32 frame
	// Radius 15 gives diameter 30, leaving 1 pixel margin on each side
	const radius = 15

	for frame := 0; frame < numFrames; frame++ {
		centerX := frame*frameWidth + frameWidth/2
		centerY := frameHeight / 2 // Centered vertically (no bounce offset)
		frameColor := colors[frame]

		// Draw filled circle
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				// Check if point is inside circle
				if dx*dx+dy*dy <= radius*radius {
					x := centerX + dx
					y := centerY + dy
					if x >= 0 && x < totalWidth && y >= 0 && y < totalHeight {
						img.Set(x, y, frameColor)
					}
				}
			}
		}

		// Add a highlight to make it look more like a ball
		highlightColor := color.RGBA{255, 255, 255, 180}
		highlightRadius := radius / 3
		highlightX := centerX - radius/3
		highlightY := centerY - radius/3

		for dy := -highlightRadius; dy <= highlightRadius; dy++ {
			for dx := -highlightRadius; dx <= highlightRadius; dx++ {
				if dx*dx+dy*dy <= highlightRadius*highlightRadius {
					x := highlightX + dx
					y := highlightY + dy
					if x >= 0 && x < totalWidth && y >= 0 && y < totalHeight {
						img.Set(x, y, highlightColor)
					}
				}
			}
		}
	}

	// Ensure output directory exists
	outputPath := filepath.Join("assets", "sprites", "test_sheet.png")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		panic(err)
	}

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Encode as PNG
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}

	println("Generated spritesheet:", outputPath)
}
