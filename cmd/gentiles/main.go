// gentiles generates a sample tileset PNG for the tilemap system.
// The tileset is 128x128 pixels with 8x8 tiles of 16x16 pixels each.
package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

const (
	tileSize  = 16
	tilesX    = 8
	tilesY    = 8
	imageSize = tileSize * tilesX // 128x128
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, imageSize, imageSize))

	// Generate each tile
	for ty := 0; ty < tilesY; ty++ {
		for tx := 0; tx < tilesX; tx++ {
			tileID := ty*tilesX + tx
			drawTile(img, tx, ty, tileID)
		}
	}

	// Create output directory
	outDir := "assets/tiles"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		panic(err)
	}

	// Write PNG file
	outPath := filepath.Join(outDir, "tiles.png")
	f, err := os.Create(outPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		panic(err)
	}

	println("Generated tileset:", outPath)
}

func drawTile(img *image.RGBA, tx, ty, tileID int) {
	// Calculate pixel offset for this tile
	offsetX := tx * tileSize
	offsetY := ty * tileSize

	for py := 0; py < tileSize; py++ {
		for px := 0; px < tileSize; px++ {
			var c color.RGBA

			switch tileID {
			case 0:
				// Empty/transparent
				c = color.RGBA{0, 0, 0, 0}

			case 1:
				// Grass top - green on top, brown on bottom
				if py < 4 {
					// Green grass
					c = color.RGBA{34, 139, 34, 255} // Forest green
				} else {
					// Brown dirt
					c = color.RGBA{139, 90, 43, 255} // Saddle brown
				}

			case 2:
				// Dirt - solid brown
				c = color.RGBA{139, 90, 43, 255}
				// Add some variation
				if (px+py)%3 == 0 {
					c = color.RGBA{120, 80, 35, 255}
				}

			case 3:
				// Stone - gray with texture
				base := uint8(128)
				variation := int8(((px*3 + py*7) % 40) - 20)
				c = color.RGBA{
					R: uint8(int(base) + int(variation)),
					G: uint8(int(base) + int(variation)),
					B: uint8(int(base) + int(variation)),
					A: 255,
				}
				// Add cracks
				if px == 7 || py == 8 {
					c = color.RGBA{100, 100, 100, 255}
				}

			case 4:
				// Brick - red-brown brick pattern
				// Determine brick row
				brickRow := py / 4
				brickOffset := 0
				if brickRow%2 == 1 {
					brickOffset = 8
				}
				brickX := (px + brickOffset) % 16

				// Mortar lines
				if py%4 == 0 || brickX == 0 {
					c = color.RGBA{180, 160, 140, 255} // Mortar
				} else {
					c = color.RGBA{178, 102, 68, 255} // Brick
					// Add variation
					if (px+py)%5 == 0 {
						c = color.RGBA{160, 90, 55, 255}
					}
				}

			case 5:
				// Grass variant with flowers
				if py < 4 {
					c = color.RGBA{34, 139, 34, 255}
					// Add small flowers
					if (px == 4 || px == 11) && py == 2 {
						c = color.RGBA{255, 255, 0, 255} // Yellow flower
					}
				} else {
					c = color.RGBA{139, 90, 43, 255}
				}

			case 6:
				// Stone variant with moss
				base := uint8(128)
				variation := int8(((px*3 + py*7) % 40) - 20)
				c = color.RGBA{
					R: uint8(int(base) + int(variation) - 20),
					G: uint8(int(base) + int(variation) + 10),
					B: uint8(int(base) + int(variation) - 10),
					A: 255,
				}
				// Moss patches
				if (px+py)%7 < 3 && py > 8 {
					c = color.RGBA{60, 120, 60, 255}
				}

			case 7:
				// Dark brick
				brickRow := py / 4
				brickOffset := 0
				if brickRow%2 == 1 {
					brickOffset = 8
				}
				brickX := (px + brickOffset) % 16

				if py%4 == 0 || brickX == 0 {
					c = color.RGBA{100, 90, 80, 255}
				} else {
					c = color.RGBA{120, 60, 40, 255}
				}

			case 8:
				// Wood planks
				if px%4 == 0 {
					c = color.RGBA{100, 70, 40, 255} // Gap
				} else {
					c = color.RGBA{160, 120, 80, 255} // Wood
					// Wood grain
					if py%3 == 0 {
						c = color.RGBA{140, 100, 60, 255}
					}
				}

			case 9:
				// Sand
				c = color.RGBA{238, 214, 175, 255}
				if (px*2+py*3)%7 == 0 {
					c = color.RGBA{220, 200, 160, 255}
				}

			case 10:
				// Water (solid blue for now)
				c = color.RGBA{65, 105, 225, 255}
				// Wave pattern
				if (px+py)%5 < 2 {
					c = color.RGBA{100, 149, 237, 255}
				}

			case 11:
				// Lava
				c = color.RGBA{255, 80, 0, 255}
				if (px*2+py)%4 < 2 {
					c = color.RGBA{255, 160, 0, 255}
				}

			default:
				// Placeholder tiles - colored squares with number pattern
				baseR := uint8((tileID * 37) % 256)
				baseG := uint8((tileID * 73) % 256)
				baseB := uint8((tileID * 113) % 256)
				c = color.RGBA{baseR, baseG, baseB, 255}

				// Add border
				if px == 0 || py == 0 || px == 15 || py == 15 {
					c = color.RGBA{255, 255, 255, 255}
				}
			}

			img.SetRGBA(offsetX+px, offsetY+py, c)
		}
	}
}
