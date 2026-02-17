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
	"github.com/torsten/GoP/internal/entities"
	"github.com/torsten/GoP/internal/game"
	"github.com/torsten/GoP/internal/gameplay"
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

	// Map and entities
	tileMap      *world.Map
	renderer     *world.MapRenderer
	collisionMap *world.CollisionMap
	entityWorld  *entities.EntityWorld
	levelData    []byte // Store raw level data for object parsing

	// Camera (enhanced with deadzone)
	camera *camera.Camera

	// Player
	playerBody       *physics.Body
	playerController *physics.Controller
	resolver         *physics.CollisionResolver

	// Gameplay state
	state *gameplay.StateMachine

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
	showDebugEntities  bool

	// Debug renderer
	debugRenderer *entities.DebugRenderer

	// Debug text
	debugText string
}

// New creates a new sandbox scene.
func New() *Scene {
	s := &Scene{
		inp:           input.NewInput(),
		width:         640,
		height:        360,
		tuning:        game.DefaultTuning(),
		timestep:      timestep.NewTimestep(),
		state:         gameplay.NewStateMachine(),
		debugRenderer: entities.NewDebugRenderer(),
	}

	// Load tileset image
	tilesetImg, err := assets.LoadTileset()
	if err != nil {
		panic(fmt.Sprintf("failed to load tileset: %v", err))
	}
	tileset := world.NewTilesetFromImage(tilesetImg, 16, 16)

	// Load map data
	s.levelData, err = assets.LoadLevelJSON()
	if err != nil {
		panic(fmt.Sprintf("failed to load level: %v", err))
	}
	mapData, err := world.ParseTiledJSON(s.levelData)
	if err != nil {
		panic(fmt.Sprintf("failed to parse level: %v", err))
	}
	s.tileMap = world.NewMap(mapData, tileset)

	// Create renderer
	s.renderer = world.NewMapRenderer(s.tileMap)

	// Create enhanced camera with deadzone
	s.camera = camera.NewCamera(s.width, s.height)
	s.camera.SetDeadzoneCentered(0.25, 0.4) // 25% width, 40% height deadzone
	s.camera.SetLevelBounds(float64(s.tileMap.PixelWidth()), float64(s.tileMap.PixelHeight()))
	s.camera.PixelPerfect = true

	// Create collision map from "Collision" layer
	s.collisionMap = world.NewCollisionMapFromMap(s.tileMap, "Collision")

	// Create entity world
	s.entityWorld = entities.NewEntityWorld()

	// Create player
	s.playerBody = &physics.Body{
		PosX: 100,
		PosY: 100,
		W:    playerSize,
		H:    playerSize,
	}
	s.playerController = physics.NewController(s.playerBody, s.tuning)
	s.resolver = physics.NewCollisionResolver(16, 16)

	// Load entities from level
	s.loadEntities()

	// Load player sprite (reuse existing ball sprite)
	if err := s.initSprite(); err != nil {
		fmt.Printf("Failed to load sprite: %v\n", err)
	}

	return s
}

// loadEntities parses the level data and spawns entities.
func (s *Scene) loadEntities() {
	// Parse objects from level data
	objects, err := world.ParseObjects(s.levelData)
	if err != nil {
		fmt.Printf("Failed to parse objects: %v\n", err)
		return
	}

	// Find spawn point
	if spawnX, spawnY, found := world.FindSpawnPoint(objects); found {
		s.playerBody.PosX = spawnX
		s.playerBody.PosY = spawnY
		s.state.SetRespawnPoint(spawnX, spawnY)
	}

	// Create spawn context with callbacks
	ctx := gameplay.SpawnContext{
		OnDeath: func() {
			s.state.TriggerDeath()
		},
		OnCheckpoint: func(id string, x, y float64) {
			s.state.SetRespawnPoint(x, y)
			fmt.Printf("Checkpoint '%s' activated at (%.0f, %.0f)\n", id, x, y)
		},
		OnGoalReached: func() {
			s.state.TriggerComplete()
			fmt.Println("Level Complete!")
		},
		Registry: s.entityWorld.TargetRegistry,
	}

	// Spawn entities
	_, triggers, solidEnts, kinematics := gameplay.SpawnEntities(objects, ctx)

	// Add entities to world
	for _, t := range triggers {
		s.entityWorld.AddTrigger(t)
	}
	for _, e := range solidEnts {
		s.entityWorld.AddSolidEntity(e)
	}
	for _, k := range kinematics {
		s.entityWorld.AddKinematic(k)
	}
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
	// Skip physics during death/respawn/completed states
	if !s.state.IsRunning() {
		return nil
	}

	dt := s.timestep.TickDuration()

	// Step 1: Update kinematic entities FIRST (platforms move before player physics)
	s.entityWorld.UpdateKinematics(s.collisionMap, dt.Seconds())

	// Step 2: Clear previous platform reference and check for carry
	s.playerController.ClearPlatformCarry()
	for _, k := range s.entityWorld.GetKinematics() {
		if k.IsActive() {
			s.playerController.ApplyPlatformCarry(k, dt)
		}
	}

	// Step 3: Create collision function for the controller
	collisionFunc := func(aabb physics.AABB) []physics.Collision {
		return s.resolveCollisions(aabb)
	}

	// Step 4: Update player physics with fixed timestep
	s.playerController.FixedUpdate(dt, s.inp, collisionFunc)

	// Step 5: Resolve player collision against solid entities (including platforms)
	s.resolveSolidEntityCollisions()

	// Step 6: Check triggers after movement
	s.entityWorld.CheckTriggers(s.playerBody)

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

	// Check solid entities (doors, etc.)
	for _, e := range s.entityWorld.SolidEntities() {
		entityBounds := e.Bounds()
		if entityBounds.W == 0 || entityBounds.H == 0 {
			// Skip entities with no collision (e.g., open doors)
			continue
		}

		if aabb.Intersects(entityBounds) {
			// Calculate collision normal
			overlapLeft := (aabb.X + aabb.W) - entityBounds.X
			overlapRight := (entityBounds.X + entityBounds.W) - aabb.X
			overlapTop := (aabb.Y + aabb.H) - entityBounds.Y
			overlapBottom := (entityBounds.Y + entityBounds.H) - aabb.Y

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
			col.TileX = -1 // Mark as entity collision
			col.TileY = -1
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

	return collisions
}

// resolveSolidEntityCollisions resolves player collision against solid entities.
// This is called after tile collision to handle entity-specific collision.
func (s *Scene) resolveSolidEntityCollisions() {
	// Collect AABBs of all active solid entities
	var solidAABBs []physics.AABB

	// Add regular solid entities
	for _, e := range s.entityWorld.SolidEntities() {
		if !e.IsActive() {
			continue
		}
		bounds := e.Bounds()
		if bounds.W > 0 && bounds.H > 0 {
			solidAABBs = append(solidAABBs, bounds)
		}
	}

	// Add kinematic entities (platforms)
	for _, k := range s.entityWorld.GetKinematics() {
		if !k.IsActive() {
			continue
		}
		body := k.GetBody()
		if body != nil && body.W > 0 && body.H > 0 {
			solidAABBs = append(solidAABBs, body.AABB())
		}
	}

	// Resolve against all solids
	if len(solidAABBs) > 0 {
		physics.ResolveSolids(s.playerBody, solidAABBs)
	}
}

// Update implements app.Scene.Update.
// This handles non-physics updates and input.
func (s *Scene) Update(inp *input.Input) error {
	// Update state machine
	s.state.Update(1.0 / 60.0)

	// Handle respawn
	if s.state.IsRespawning() {
		s.respawnPlayer()
	}

	// Handle debug toggles
	s.handleDebugToggles()

	// Camera follows player
	playerCenterX := s.playerBody.PosX + s.playerBody.W/2
	playerCenterY := s.playerBody.PosY + s.playerBody.H/2
	s.camera.Follow(playerCenterX, playerCenterY, s.playerBody.W, s.playerBody.H)
	s.camera.Update(1.0 / 60.0)

	// Update entities
	s.entityWorld.Update(1.0 / 60.0)

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

// respawnPlayer resets player position to the respawn point.
func (s *Scene) respawnPlayer() {
	s.playerBody.PosX = s.state.RespawnX
	s.playerBody.PosY = s.state.RespawnY
	s.playerBody.VelX = 0
	s.playerBody.VelY = 0
	s.state.FinishRespawn()
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
	if inpututil.IsKeyJustPressed(ebiten.KeyF6) {
		s.showDebugEntities = !s.showDebugEntities
	}
	// Force respawn with R
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		s.state.TriggerDeath()
	}
}

// updateDebugText updates the debug text string.
func (s *Scene) updateDebugText() {
	// Build platform info
	platformInfo := "onPlatform: false"
	if platform := s.playerController.GetCurrentPlatform(); platform != nil {
		// Try to get platform ID if available
		if idGetter, ok := platform.(interface{ GetID() string }); ok {
			platformInfo = fmt.Sprintf("onPlatform: true\nplatform: %s", idGetter.GetID())
		} else {
			platformInfo = "onPlatform: true"
		}
	}

	s.debugText = fmt.Sprintf("pos: (%.1f, %.1f)\nvel: (%.1f, %.1f)\ngrounded: %v\n%s\nstate: %s\nF2: collision | F3: deadzone | F4: state | F5: steps | F6: entities | R: respawn",
		s.playerBody.PosX, s.playerBody.PosY,
		s.playerBody.VelX, s.playerBody.VelY,
		s.playerBody.OnGround,
		platformInfo,
		s.state.Current.String())
}

// Draw implements app.Scene.Draw.
func (s *Scene) Draw(screen *ebiten.Image) {
	// Fill background
	screen.Fill(backgroundColor)

	// Create render context
	ctx := world.NewRenderContext(s.camera, screen, 1.0/60.0)
	ctx.Debug = s.showDebugEntities

	// Draw map with camera offset
	s.renderer.DrawWithContext(screen, ctx)

	// Draw entities
	s.entityWorld.DrawWithContext(screen, ctx)

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
	if s.showDebugEntities {
		s.debugRenderer.ShowAll = true
		s.debugRenderer.DrawWithContext(screen, s.entityWorld, ctx)
		s.debugRenderer.DrawPlayerDebugWithContext(screen, s.playerBody, ctx)
		// Draw platform debug visualization (paths and bounds)
		s.entityWorld.DrawKinematicsDebug(screen, ctx)
	}

	// Draw state overlay
	if s.state.IsDead() {
		s.drawDeathOverlay(screen)
	} else if s.state.IsCompleted() {
		s.drawCompleteOverlay(screen)
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

// drawDeathOverlay shows a death message.
func (s *Scene) drawDeathOverlay(screen *ebiten.Image) {
	text := "YOU DIED"
	x := s.width/2 - 30
	y := s.height/2 - 10
	ebitenutil.DebugPrintAt(screen, text, x, y)
}

// drawCompleteOverlay shows a level complete message.
func (s *Scene) drawCompleteOverlay(screen *ebiten.Image) {
	text := "LEVEL COMPLETE!"
	x := s.width/2 - 50
	y := s.height/2 - 10
	ebitenutil.DebugPrintAt(screen, text, x, y)
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
		s.camera.ViewportW = outsideW
		s.camera.ViewportH = outsideH
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
