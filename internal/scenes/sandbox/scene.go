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
	"github.com/torsten/GoP/internal/camera"
	"github.com/torsten/GoP/internal/game"
	"github.com/torsten/GoP/internal/gfx"
	"github.com/torsten/GoP/internal/input"
	"github.com/torsten/GoP/internal/physics"
	timestep "github.com/torsten/GoP/internal/time"
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
	deadzoneColor   = color.RGBA{0xff, 0xff, 0x00, 0x60}
)

// Scene represents the sandbox test scene with tilemap and physics.
type Scene struct {
	// Input
	inp *input.Input

	// Map
	tileMap      *world.Map
	renderer     *world.MapRenderer
	collisionMap *world.CollisionMap

	// Camera (enhanced with deadzone)
	camera *camera.Camera

	// Player
	playerBody       *physics.Body
	playerController *physics.Controller
	resolver         *physics.CollisionResolver

	// Tuning parameters
	tuning game.Tuning

	// Fixed timestep
	timestep *timestep.Timestep

	// Player sprite (reuse existing ball animation)
	sprite   *gfx.Sprite
	animator *gfx.Animator

	// Screen dimensions
	width  int
	height int

	// Debug toggles
	showDebugCollision bool
	showDebugDeadzone  bool
	showDebugState     bool
	showDebugSteps     bool

	// Debug text
	debugText string
}

// New creates a new sandbox scene.
func New() *Scene {
	s := &Scene{
		inp:      input.NewInput(),
		width:    640,
		height:   360,
		tuning:   game.DefaultTuning(),
		timestep: timestep.NewTimestep(),
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

	// Create enhanced camera with deadzone
	// Use smaller viewport (half screen size) for deadzone testing
	viewportW := s.width / 2
	viewportH := s.height / 2
	s.camera = camera.NewCamera(viewportW, viewportH)
	s.camera.SetDeadzoneCentered(0.25, 0.4) // 25% width, 40% height deadzone
	s.camera.SetLevelBounds(float64(s.tileMap.PixelWidth()), float64(s.tileMap.PixelHeight()))
	s.camera.PixelPerfect = true

	// Create collision map from "Collision" layer
	s.collisionMap = world.NewCollisionMapFromMap(s.tileMap, "Collision")

	// Create player
	s.playerBody = &physics.Body{
		PosX: 100,
		PosY: 100,
		W:    playerSize,
		H:    playerSize,
	}
	s.playerController = physics.NewController(s.playerBody, s.tuning)
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

// FixedUpdate handles physics updates at fixed rate.
// This is called multiple times per frame if needed.
func (s *Scene) FixedUpdate() error {
	// Create collision function for the controller
	collisionFunc := func(aabb physics.AABB) []physics.Collision {
		return s.resolveCollisions(aabb)
	}

	// Update player physics with fixed timestep
	s.playerController.FixedUpdate(s.timestep.TickDuration(), s.inp, collisionFunc)

	return nil
}

// resolveCollisions checks for collisions at the given AABB and returns collision info.
func (s *Scene) resolveCollisions(aabb physics.AABB) []physics.Collision {
	var collisions []physics.Collision

	// Get tile range to check
	tileSize := 16
	startTX := int(aabb.X) / tileSize
	startTY := int(aabb.Y) / tileSize
	endTX := int(aabb.X+aabb.W) / tileSize
	endTY := int(aabb.Y+aabb.H) / tileSize

	// Clamp to map bounds
	if startTX < 0 {
		startTX = 0
	}
	if startTY < 0 {
		startTY = 0
	}
	if endTX >= s.tileMap.Width() {
		endTX = s.tileMap.Width() - 1
	}
	if endTY >= s.tileMap.Height() {
		endTY = s.tileMap.Height() - 1
	}

	// Check each tile
	for ty := startTY; ty <= endTY; ty++ {
		for tx := startTX; tx <= endTX; tx++ {
			if s.collisionMap.IsSolidAtTile(tx, ty) {
				// Calculate collision normal
				tileLeft := float64(tx * tileSize)
				tileRight := float64((tx + 1) * tileSize)
				tileTop := float64(ty * tileSize)
				tileBottom := float64((ty + 1) * tileSize)

				// Calculate overlap on each axis
				overlapLeft := (aabb.X + aabb.W) - tileLeft
				overlapRight := tileRight - aabb.X
				overlapTop := (aabb.Y + aabb.H) - tileTop
				overlapBottom := tileBottom - aabb.Y

				// Find minimum overlap axis
				minOverlapX := overlapLeft
				normalX := -1.0
				if overlapRight < overlapLeft {
					minOverlapX = overlapRight
					normalX = 1.0
				}

				minOverlapY := overlapTop
				normalY := -1.0
				if overlapBottom < overlapTop {
					minOverlapY = overlapBottom
					normalY = 1.0
				}

				// Use the axis with minimum overlap
				var col physics.Collision
				col.TileX = tx
				col.TileY = ty
				if minOverlapX < minOverlapY {
					col.NormalX = normalX
					col.NormalY = 0
				} else {
					col.NormalX = 0
					col.NormalY = normalY
				}

				collisions = append(collisions, col)
			}
		}
	}

	return collisions
}

// Update implements app.Scene.Update.
// This handles non-physics updates and input.
func (s *Scene) Update(inp *input.Input) error {
	// Handle debug toggles
	s.handleDebugToggles()

	// Camera follows player
	playerCenterX := s.playerBody.PosX + s.playerBody.W/2
	playerCenterY := s.playerBody.PosY + s.playerBody.H/2
	s.camera.Follow(playerCenterX, playerCenterY, s.playerBody.W, s.playerBody.H)
	s.camera.Update(1.0 / 60.0)

	// Update animator (non-physics)
	if s.animator != nil {
		s.animator.Update(time.Second / 60)
	}

	// Update debug text
	s.updateDebugText()

	// Update input state at end of frame so JustPressed works correctly
	s.inp.Update()

	return nil
}

// handleDebugToggles processes debug key bindings.
func (s *Scene) handleDebugToggles() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		s.showDebugCollision = !s.showDebugCollision
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		s.showDebugDeadzone = !s.showDebugDeadzone
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF4) {
		s.showDebugState = !s.showDebugState
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		s.showDebugSteps = !s.showDebugSteps
	}
}

// updateDebugText updates the debug text string.
func (s *Scene) updateDebugText() {
	s.debugText = fmt.Sprintf("pos: (%.1f, %.1f)\nvel: (%.1f, %.1f)\ngrounded: %v\nF2: collision | F3: deadzone | F4: state | F5: steps",
		s.playerBody.PosX, s.playerBody.PosY,
		s.playerBody.VelX, s.playerBody.VelY,
		s.playerBody.OnGround)
}

// Draw implements app.Scene.Draw.
func (s *Scene) Draw(screen *ebiten.Image) {
	// Fill background
	screen.Fill(backgroundColor)

	// Draw map with camera offset
	s.renderer.Draw(screen, s.camera.X, s.camera.Y)

	// Draw player
	s.drawPlayer(screen)

	// Draw debug overlays
	if s.showDebugCollision {
		s.drawCollisionDebug(screen)
	}
	if s.showDebugDeadzone {
		s.drawDeadzone(screen)
	}
	if s.showDebugState {
		s.drawPlayerState(screen)
	}
	if s.showDebugSteps {
		s.drawStepCounter(screen)
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

// drawDeadzone visualizes the camera deadzone.
func (s *Scene) drawDeadzone(screen *ebiten.Image) {
	// Draw deadzone rectangle in screen space
	ebitenutil.DrawRect(screen,
		s.camera.DeadzoneX, s.camera.DeadzoneY,
		s.camera.DeadzoneW, s.camera.DeadzoneH,
		deadzoneColor)

	// Draw camera position info
	camInfo := fmt.Sprintf("Camera: (%.0f, %.0f)\nTarget: (%.0f, %.0f)",
		s.camera.X, s.camera.Y,
		s.camera.TargetX(), s.camera.TargetY())
	ebitenutil.DebugPrintAt(screen, camInfo, 10, s.height-50)
}

// drawPlayerState shows player state information.
func (s *Scene) drawPlayerState(screen *ebiten.Image) {
	state := s.playerController.State
	info := fmt.Sprintf(
		"Vel: (%.1f, %.1f)\nGrounded: %v\nCoyote: %.0fms\nBuffer: %.0fms\nJumping: %v",
		s.playerBody.VelX, s.playerBody.VelY,
		s.playerBody.OnGround,
		state.TimeSinceGrounded.Seconds()*1000,
		state.JumpBufferTime.Seconds()*1000,
		state.IsJumping,
	)
	ebitenutil.DebugPrintAt(screen, info, s.width-120, 10)
}

// drawStepCounter shows physics step statistics.
func (s *Scene) drawStepCounter(screen *ebiten.Image) {
	info := fmt.Sprintf("Steps: %d/frame\nTotal: %d",
		s.timestep.StepsThisFrame(),
		s.timestep.TotalTicks())
	ebitenutil.DebugPrintAt(screen, info, s.width-100, s.height-40)
}

// Layout implements app.Scene.Layout.
func (s *Scene) Layout(outsideW, outsideH int) (int, int) {
	s.width = outsideW
	s.height = outsideH
	if s.camera != nil {
		// Keep viewport at half screen size for deadzone testing
		s.camera.ViewportW = outsideW / 2
		s.camera.ViewportH = outsideH / 2
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
