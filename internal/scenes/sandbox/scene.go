// Package sandbox provides a test scene for development and prototyping.
package sandbox

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/torsten/GoP/internal/assets"
	"github.com/torsten/GoP/internal/gfx"
	"github.com/torsten/GoP/internal/input"
	"github.com/torsten/GoP/internal/physics"
	"github.com/torsten/GoP/internal/world"
)

const (
	// Animation frame duration.
	frameDuration = 100 * time.Millisecond
	// Player size in pixels.
	playerSize = 12
)

// Colors for the scene.
var (
	backgroundColor = color.RGBA{0x10, 0x10, 0x20, 0xff}
	collisionColor  = color.RGBA{0xff, 0x00, 0x00, 0x80}
	playerColor     = color.RGBA{0x00, 0xff, 0x00, 0xff}
)

// Scene represents the sandbox test scene with tilemap and physics.
type Scene struct {
	// Input
	inp *input.Input

	// Map
	tileMap      *world.Map
	renderer     *world.MapRenderer
	camera       *world.Camera
	collisionMap *world.CollisionMap

	// Player
	playerBody       *physics.Body
	playerController *physics.PlayerController
	resolver         *physics.CollisionResolver

	// Player sprite (reuse existing ball animation)
	sprite   *gfx.Sprite
	animator *gfx.Animator

	// Screen dimensions
	width  int
	height int

	// Debug
	showCollision bool
	debugText     string
}

// New creates a new sandbox scene.
func New() *Scene {
	s := &Scene{
		inp:    input.NewInput(),
		width:  640,
		height: 360,
	}

	// Load tileset image
	tilesetImg, err := assets.LoadTileset()
	if err != nil {
		panic(fmt.Sprintf("failed to load tileset: %v", err))
	}
	tileset := world.NewTilesetFromImage(tilesetImg, 16, 16)

	// Load map data
	levelData, err := assets.LoadLevelJSON()
	if err != nil {
		panic(fmt.Sprintf("failed to load level: %v", err))
	}
	mapData, err := world.ParseTiledJSON(levelData)
	if err != nil {
		panic(fmt.Sprintf("failed to parse level: %v", err))
	}
	s.tileMap = world.NewMap(mapData, tileset)

	// Create renderer
	s.renderer = world.NewMapRenderer(s.tileMap)

	// Create camera
	s.camera = world.NewCamera(s.width, s.height)

	// Create collision map from "Collision" layer
	s.collisionMap = world.NewCollisionMapFromMap(s.tileMap, "Collision")

	// Create player
	s.playerBody = &physics.Body{
		PosX: 100,
		PosY: 100,
		W:    playerSize,
		H:    playerSize,
	}
	s.playerController = physics.NewPlayerController(s.playerBody)
	s.resolver = physics.NewCollisionResolver(16, 16)

	// Load player sprite (reuse existing ball sprite)
	if err := s.initSprite(); err != nil {
		fmt.Printf("Failed to load sprite: %v\n", err)
	}

	return s
}

// initSprite loads the spritesheet and creates the sprite animation.
func (s *Scene) initSprite() error {
	// Load spritesheet from assets
	sheetImg, err := assets.LoadSpriteSheet()
	if err != nil {
		return fmt.Errorf("failed to load spritesheet: %w", err)
	}

	// Create sheet with 32x32 frames
	sheet := assets.NewSheet(sheetImg, 32, 32)

	// Create animation from sheet
	anim := gfx.NewAnimationFromSheet(sheet, frameDuration)
	anim.Loop = true

	// Create animator and start playing
	s.animator = gfx.NewAnimator(anim)
	s.animator.Play()

	// Create sprite
	s.sprite = gfx.NewSprite(nil) // Image will be set in Draw
	s.sprite.SetScale(0.5, 0.5)   // Scale down to match player size (32*0.5 = 16)
	s.sprite.SetOrigin(0.5, 0.5)  // Center origin

	return nil
}

// Update implements app.Scene.Update.
func (s *Scene) Update(inp *input.Input) error {
	// Toggle collision debug with F2
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		s.showCollision = !s.showCollision
	}

	// Update player physics
	dt := 1.0 / 60.0 // Fixed timestep
	s.playerController.Update(s.inp, s.collisionMap, s.resolver, dt)

	// Update animator
	if s.animator != nil {
		s.animator.Update(time.Second / 60)
	}

	// Center camera on player
	playerCenterX := s.playerBody.PosX + s.playerBody.W/2
	playerCenterY := s.playerBody.PosY + s.playerBody.H/2
	s.camera.CenterOn(playerCenterX, playerCenterY, s.tileMap.PixelWidth(), s.tileMap.PixelHeight())

	// Update debug text
	s.debugText = fmt.Sprintf("pos: (%.1f, %.1f)\nvel: (%.1f, %.1f)\ngrounded: %v\nF2: toggle collision",
		s.playerBody.PosX, s.playerBody.PosY,
		s.playerBody.VelX, s.playerBody.VelY,
		s.playerBody.OnGround)

	// Update input state at end of frame so JustPressed works correctly
	s.inp.Update()

	return nil
}

// Draw implements app.Scene.Draw.
func (s *Scene) Draw(screen *ebiten.Image) {
	// Fill background
	screen.Fill(backgroundColor)

	// Draw map with camera offset
	s.renderer.Draw(screen, s.camera.X, s.camera.Y)

	// Draw player
	s.drawPlayer(screen)

	// Draw collision debug overlay
	if s.showCollision {
		s.drawCollisionDebug(screen)
	}

	// Draw debug text
	ebitenutil.DebugPrint(screen, s.debugText)
}

// drawPlayer renders the player sprite or a fallback rectangle.
func (s *Scene) drawPlayer(screen *ebiten.Image) {
	// Calculate screen position (center of player body)
	screenX := s.playerBody.PosX + s.playerBody.W/2 - s.camera.X
	screenY := s.playerBody.PosY + s.playerBody.H/2 - s.camera.Y

	if s.sprite != nil && s.animator != nil {
		// Update sprite image from current animation frame
		s.sprite.Image = s.animator.CurrentFrame()
		// Update position
		s.sprite.SetPosition(screenX, screenY)
		// Draw sprite
		s.sprite.Draw(screen)
	} else {
		// Fallback: draw a rectangle
		drawX := s.playerBody.PosX - s.camera.X
		drawY := s.playerBody.PosY - s.camera.Y
		ebitenutil.DrawRect(screen, drawX, drawY, s.playerBody.W, s.playerBody.H, playerColor)
	}
}

// drawCollisionDebug draws the collision overlay.
func (s *Scene) drawCollisionDebug(screen *ebiten.Image) {
	// Draw solid tiles as semi-transparent red rectangles
	tileSize := 16

	startX := int(s.camera.X) / tileSize
	startY := int(s.camera.Y) / tileSize
	endX := startX + s.width/tileSize + 2
	endY := startY + s.height/tileSize + 2

	// Clamp to map bounds
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	if endX > s.tileMap.Width() {
		endX = s.tileMap.Width()
	}
	if endY > s.tileMap.Height() {
		endY = s.tileMap.Height()
	}

	for ty := startY; ty < endY; ty++ {
		for tx := startX; tx < endX; tx++ {
			if s.collisionMap.IsSolidAtTile(tx, ty) {
				x := float64(tx*tileSize) - s.camera.X
				y := float64(ty*tileSize) - s.camera.Y
				ebitenutil.DrawRect(screen, x, y, float64(tileSize), float64(tileSize), collisionColor)
			}
		}
	}

	// Draw player AABB
	playerScreenX := s.playerBody.PosX - s.camera.X
	playerScreenY := s.playerBody.PosY - s.camera.Y
	borderColor := color.RGBA{255, 255, 255, 255}
	borderWidth := 1.0

	// Top border
	ebitenutil.DrawRect(screen, playerScreenX, playerScreenY, s.playerBody.W, borderWidth, borderColor)
	// Bottom border
	ebitenutil.DrawRect(screen, playerScreenX, playerScreenY+s.playerBody.H-borderWidth, s.playerBody.W, borderWidth, borderColor)
	// Left border
	ebitenutil.DrawRect(screen, playerScreenX, playerScreenY, borderWidth, s.playerBody.H, borderColor)
	// Right border
	ebitenutil.DrawRect(screen, playerScreenX+s.playerBody.W-borderWidth, playerScreenY, borderWidth, s.playerBody.H, borderColor)
}

// Layout implements app.Scene.Layout.
func (s *Scene) Layout(outsideW, outsideH int) (int, int) {
	s.width = outsideW
	s.height = outsideH
	if s.camera != nil {
		s.camera.ViewWidth = outsideW
		s.camera.ViewHeight = outsideH
	}
	return outsideW, outsideH
}

// DebugInfo implements app.Scene.DebugInfo.
func (s *Scene) DebugInfo() string {
	return s.debugText
}

// DrawDebug implements app.SceneDebugger.DrawDebug.
func (s *Scene) DrawDebug(screen *ebiten.Image) {
	// Just use the collision debug overlay
	s.drawCollisionDebug(screen)
}
