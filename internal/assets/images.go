package assets

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png" // PNG decoder
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
)

// LoadImage loads an image from a filesystem and converts it to *ebiten.Image.
// The image is created with default settings suitable for pixel art rendering.
func LoadImage(fsys fs.FS, path string) (*ebiten.Image, error) {
	file, err := fsys.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image %s: %w", path, err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image %s: %w", path, err)
	}

	return ebiten.NewImageFromImage(img), nil
}

// LoadImageFromBytes loads an image from raw byte data and converts it to *ebiten.Image.
// The image is created with default settings suitable for pixel art rendering.
func LoadImageFromBytes(data []byte) (*ebiten.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image from bytes: %w", err)
	}

	return ebiten.NewImageFromImage(img), nil
}
