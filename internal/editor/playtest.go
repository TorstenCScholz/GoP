package editor

import (
	"fmt"
	"image/color"
	"log"
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

// Constants for playtest mode
const (
	playtestFrameDuration = 100 * time.Millisecond
	playtestPlayerSize    = 12
)

// Colors for playtest rendering
var (
	playtestBackgroundColor = color.RGBA{0x10, 0x10, 0x20, 0xff}
	playtestPlayerColor     = color.RGBA{0x00, 0xff, 0x00, 0xff}
)

// EditorSnapshot captures the editor state for restoration after playtest.
type EditorSnapshot struct {
	CameraX       float64
	CameraY       float64
	Zoom          float64
	SelectedTool  Tool
	SelectedTile  int
	SelectedLayer string
}

// PlaytestController manages the playtest mode for the level editor.
// It handles entering/exiting playtest mode and manages the game scene.
type PlaytestController struct {
	editor     *App
	savedState *EditorSnapshot

	// Playtest scene components
	inp          *input.Input
	tileMap      *world.Map
	renderer     *world.MapRenderer
	collisionMap *world.CollisionMap
	entityWorld  *entities.EntityWorld
	camera       *camera.Camera
	playerBody   *physics.Body
	playerCtrl   *physics.Controller
	resolver     *physics.CollisionResolver
	state        *gameplay.StateMachine
	tuning       game.Tuning
	timestep     *timestep.Timestep
	sprite       *gfx.Sprite
	animator     *gfx.Animator

	// State
	isActive      bool
	width         int
	height        int
	initialSpawnX float64
	initialSpawnY float64
}

// NewPlaytestController creates a new playtest controller.
func NewPlaytestController(editor *App) *PlaytestController {
	return &PlaytestController{
		editor:   editor,
		inp:      input.NewInput(),
		tuning:   game.DefaultTuning(),
		timestep: timestep.NewTimestep(),
		state:    gameplay.NewStateMachine(),
	}
}

// IsActive returns true if playtest mode is currently active.
func (p *PlaytestController) IsActive() bool {
	return p.isActive
}

// StartPlaytest initializes and starts playtest mode.
func (p *PlaytestController) StartPlaytest() error {
	if p.isActive {
		return nil
	}

	log.Println("Entering playtest mode...")

	// 1. Save current editor state
	p.savedState = p.CreateSnapshot()

	// 2. Build game map from editor data
	if err := p.buildGameScene(); err != nil {
		return fmt.Errorf("failed to build game scene: %w", err)
	}

	// 3. Transition to playtest mode
	p.isActive = true

	log.Println("Playtest mode active. Press Escape to return to editor, R to restart.")
	return nil
}

// EndPlaytest exits playtest mode and restores editor state.
func (p *PlaytestController) EndPlaytest() {
	if !p.isActive {
		return
	}

	log.Println("Exiting playtest mode...")

	// 1. Clean up game scene resources
	p.cleanupGameScene()

	// 2. Restore editor state from snapshot
	p.RestoreSnapshot(p.savedState)

	// 3. Clear saved state
	p.savedState = nil

	// 4. Return to edit mode
	p.isActive = false

	log.Println("Returned to editor mode.")
}

// RestartPlaytest restarts the playtest from the beginning.
func (p *PlaytestController) RestartPlaytest() {
	if !p.isActive {
		return
	}

	log.Println("Restarting playtest...")

	// Reset player to initial spawn point
	p.playerBody.PosX = p.initialSpawnX
	p.playerBody.PosY = p.initialSpawnY
	p.playerBody.VelX = 0
	p.playerBody.VelY = 0

	// Reset game state
	p.state = gameplay.NewStateMachine()
	p.state.SetRespawnPoint(p.initialSpawnX, p.initialSpawnY)

	// Reset entities
	p.rebuildEntities()
}

// CreateSnapshot captures the current editor state.
func (p *PlaytestController) CreateSnapshot() *EditorSnapshot {
	state := p.editor.State()
	return &EditorSnapshot{
		CameraX:       state.CameraX,
		CameraY:       state.CameraY,
		Zoom:          state.Zoom,
		SelectedTool:  state.CurrentTool,
		SelectedTile:  state.SelectedTile,
		SelectedLayer: state.CurrentLayer,
	}
}

// RestoreSnapshot restores editor state from a snapshot.
func (p *PlaytestController) RestoreSnapshot(snapshot *EditorSnapshot) {
	if snapshot == nil {
		return
	}

	state := p.editor.State()
	state.CameraX = snapshot.CameraX
	state.CameraY = snapshot.CameraY
	state.Zoom = snapshot.Zoom
	state.CurrentTool = snapshot.SelectedTool
	state.SelectedTile = snapshot.SelectedTile
	state.CurrentLayer = snapshot.SelectedLayer

	// Restore camera position
	cam := p.editor.Camera()
	cam.X = snapshot.CameraX
	cam.Y = snapshot.CameraY
	cam.Zoom = snapshot.Zoom
}

// Update handles playtest mode updates.
func (p *PlaytestController) Update() error {
	if !p.isActive {
		return nil
	}

	// Handle playtest controls
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		p.EndPlaytest()
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		p.RestartPlaytest()
		return nil
	}

	// Add frame time to timestep accumulator
	p.timestep.AddFrameTime(time.Second / 60)

	// Run fixed timestep physics
	for p.timestep.ShouldUpdate() {
		p.FixedUpdate()
		p.timestep.ConsumeTick()
	}

	// Update game state
	p.state.Update(1.0 / 60.0)

	// Handle respawn
	if p.state.IsRespawning() {
		p.respawnPlayer()
	}

	// Camera follows player
	playerCenterX := p.playerBody.PosX + p.playerBody.W/2
	playerCenterY := p.playerBody.PosY + p.playerBody.H/2
	p.camera.Follow(playerCenterX, playerCenterY, p.playerBody.W, p.playerBody.H)
	p.camera.Update(1.0 / 60.0)

	// Update entities
	p.entityWorld.Update(1.0 / 60.0)

	// Update animator
	if p.animator != nil {
		p.animator.Update(playtestFrameDuration)
	}

	// Update input state
	p.inp.Update()

	return nil
}

// FixedUpdate handles physics updates at fixed rate.
func (p *PlaytestController) FixedUpdate() {
	if !p.state.IsRunning() {
		return
	}

	dt := p.timestep.TickDuration()

	// Step 1: Update kinematic entities FIRST
	p.entityWorld.UpdateKinematics(p.collisionMap, dt.Seconds())

	// Step 2: Clear previous platform reference and check for carry
	p.playerCtrl.ClearPlatformCarry()
	for _, k := range p.entityWorld.GetKinematics() {
		if k.IsActive() {
			p.playerCtrl.ApplyPlatformCarry(k, dt)
		}
	}

	// Step 3: Create collision function
	collisionFunc := func(aabb physics.AABB) []physics.Collision {
		return p.resolveCollisions(aabb)
	}

	// Step 4: Update player physics
	p.playerCtrl.FixedUpdate(dt, p.inp, collisionFunc)

	// Step 5: Resolve solid entity collisions
	p.resolveSolidEntityCollisions()

	// Step 6: Check triggers
	p.entityWorld.CheckTriggers(p.playerBody)
}

// Draw renders the playtest mode.
func (p *PlaytestController) Draw(screen *ebiten.Image) {
	if !p.isActive {
		return
	}

	// Fill background
	screen.Fill(playtestBackgroundColor)

	// Create render context
	ctx := world.NewRenderContext(p.camera, screen, 1.0/60.0)

	// Draw map
	p.renderer.DrawWithContext(screen, ctx)

	// Draw entities
	p.entityWorld.DrawWithContext(screen, ctx)

	// Draw player
	p.drawPlayer(screen)

	// Draw state overlay
	if p.state.IsDead() {
		p.drawDeathOverlay(screen)
	} else if p.state.IsCompleted() {
		p.drawCompleteOverlay(screen)
	}

	// Draw playtest indicator
	p.drawPlaytestIndicator(screen)
}

// Layout handles screen size changes during playtest.
func (p *PlaytestController) Layout(outsideW, outsideH int) (int, int) {
	p.width = outsideW
	p.height = outsideH
	if p.camera != nil {
		p.camera.ViewportW = outsideW
		p.camera.ViewportH = outsideH
		// Note: Deadzone is set once in buildGameScene() and should not be modified here
		// as it can cause camera position changes during rendering
	}
	return outsideW, outsideH
}

// buildGameScene creates the game scene from editor data.
func (p *PlaytestController) buildGameScene() error {
	state := p.editor.State()

	// Get tileset from editor
	editorTileset := p.editor.Tileset()
	if editorTileset == nil || !editorTileset.IsLoaded() {
		return fmt.Errorf("no tileset loaded")
	}

	// Get map data from editor
	if state.MapData == nil {
		return fmt.Errorf("no map data available")
	}

	// Get the world tileset from editor tileset
	worldTileset := editorTileset.Tileset()
	if worldTileset == nil {
		return fmt.Errorf("failed to get world tileset")
	}

	// Create the game map
	p.tileMap = world.NewMap(state.MapData, worldTileset)

	// Create renderer
	p.renderer = world.NewMapRenderer(p.tileMap)

	// Get screen dimensions from editor if not already set
	if p.width == 0 || p.height == 0 {
		editorWidth, editorHeight := p.editor.ScreenSize()
		p.width = editorWidth
		p.height = editorHeight
	}

	// Create camera with proper viewport dimensions
	p.camera = camera.NewCamera(p.width, p.height)
	p.camera.SetDeadzoneCentered(0.25, 0.4)
	p.camera.SetLevelBounds(float64(p.tileMap.PixelWidth()), float64(p.tileMap.PixelHeight()))
	p.camera.PixelPerfect = true

	// Create collision map from "Collision" layer
	p.collisionMap = world.NewCollisionMapFromMap(p.tileMap, "Collision")

	// Create entity world
	p.entityWorld = entities.NewEntityWorld()

	// Create player
	p.playerBody = &physics.Body{
		PosX: 100,
		PosY: 100,
		W:    playtestPlayerSize,
		H:    playtestPlayerSize,
	}
	p.playerCtrl = physics.NewController(p.playerBody, p.tuning)
	p.resolver = physics.NewCollisionResolver(16, 16)

	// Reset game state
	p.state = gameplay.NewStateMachine()
	p.timestep = timestep.NewTimestep()

	// Load entities from editor objects
	p.loadEntitiesFromEditor(state.Objects)

	// Initialize camera position to center on player spawn point
	// This prevents flickering on the first frames
	p.camera.X = p.playerBody.PosX - float64(p.camera.ViewportW)/2
	p.camera.Y = p.playerBody.PosY - float64(p.camera.ViewportH)/2

	// Initialize sprite
	if err := p.initSprite(); err != nil {
		log.Printf("Warning: failed to load sprite: %v", err)
	}

	return nil
}

// loadEntitiesFromEditor spawns entities from editor object data.
func (p *PlaytestController) loadEntitiesFromEditor(objects []world.ObjectData) {
	// Find spawn point
	if spawnX, spawnY, found := world.FindSpawnPoint(objects); found {
		p.playerBody.PosX = spawnX
		p.playerBody.PosY = spawnY
		p.initialSpawnX = spawnX
		p.initialSpawnY = spawnY
	} else {
		p.initialSpawnX = 100
		p.initialSpawnY = 100
	}
	p.state.SetRespawnPoint(p.initialSpawnX, p.initialSpawnY)

	// Create spawn context
	ctx := gameplay.SpawnContext{
		OnDeath: func() {
			p.state.TriggerDeath()
		},
		OnCheckpoint: func(id string, x, y float64) {
			p.state.SetRespawnPoint(x, y)
			log.Printf("Checkpoint '%s' activated at (%.0f, %.0f)", id, x, y)
		},
		OnGoalReached: func() {
			p.state.TriggerComplete()
			log.Println("Level Complete!")
		},
		Registry: p.entityWorld.TargetRegistry,
	}

	// Spawn entities
	_, triggers, solidEnts, kinematics := gameplay.SpawnEntities(objects, ctx)

	// Add entities to world
	for _, t := range triggers {
		p.entityWorld.AddTrigger(t)
	}
	for _, e := range solidEnts {
		p.entityWorld.AddSolidEntity(e)
	}
	for _, k := range kinematics {
		p.entityWorld.AddKinematic(k)
	}
}

// rebuildEntities recreates entities from editor data (for restart).
func (p *PlaytestController) rebuildEntities() {
	state := p.editor.State()

	// Clear existing entities
	p.entityWorld = entities.NewEntityWorld()

	// Recreate spawn context
	ctx := gameplay.SpawnContext{
		OnDeath: func() {
			p.state.TriggerDeath()
		},
		OnCheckpoint: func(id string, x, y float64) {
			p.state.SetRespawnPoint(x, y)
		},
		OnGoalReached: func() {
			p.state.TriggerComplete()
		},
		Registry: p.entityWorld.TargetRegistry,
	}

	// Spawn entities
	_, triggers, solidEnts, kinematics := gameplay.SpawnEntities(state.Objects, ctx)

	for _, t := range triggers {
		p.entityWorld.AddTrigger(t)
	}
	for _, e := range solidEnts {
		p.entityWorld.AddSolidEntity(e)
	}
	for _, k := range kinematics {
		p.entityWorld.AddKinematic(k)
	}
}

// cleanupGameScene releases game scene resources.
func (p *PlaytestController) cleanupGameScene() {
	p.tileMap = nil
	p.renderer = nil
	p.collisionMap = nil
	p.entityWorld = nil
	p.camera = nil
	p.playerBody = nil
	p.playerCtrl = nil
	p.resolver = nil
	p.state = nil
	p.sprite = nil
	p.animator = nil
}

// initSprite loads the player sprite.
func (p *PlaytestController) initSprite() error {
	// Load spritesheet from assets
	sheetImg, err := assets.LoadSpriteSheet()
	if err != nil {
		return fmt.Errorf("failed to load spritesheet: %w", err)
	}

	// Create sheet with 32x32 frames
	sheet := assets.NewSheet(sheetImg, 32, 32)

	// Create animation
	anim := gfx.NewAnimationFromSheet(sheet, playtestFrameDuration)
	anim.Loop = true

	// Create animator
	p.animator = gfx.NewAnimator(anim)
	p.animator.Play()

	// Create sprite
	p.sprite = gfx.NewSprite(nil)
	p.sprite.SetScale(0.5, 0.5)
	p.sprite.SetOrigin(0.5, 0.5)

	return nil
}

// drawPlayer renders the player.
func (p *PlaytestController) drawPlayer(screen *ebiten.Image) {
	screenX := p.playerBody.PosX + p.playerBody.W/2 - p.camera.X
	screenY := p.playerBody.PosY + p.playerBody.H/2 - p.camera.Y

	if p.sprite != nil && p.animator != nil {
		p.sprite.Image = p.animator.CurrentFrame()
		p.sprite.SetPosition(screenX, screenY)
		p.sprite.Draw(screen)
	} else {
		// Fallback rectangle
		drawX := p.playerBody.PosX - p.camera.X
		drawY := p.playerBody.PosY - p.camera.Y
		ebitenutil.DrawRect(screen, drawX, drawY, p.playerBody.W, p.playerBody.H, playtestPlayerColor)
	}
}

// drawPlaytestIndicator shows the playtest mode indicator.
func (p *PlaytestController) drawPlaytestIndicator(screen *ebiten.Image) {
	text := "PLAYTEST MODE | ESC: Exit | R: Restart"
	ebitenutil.DebugPrintAt(screen, text, 10, 10)
}

// drawDeathOverlay shows death message.
func (p *PlaytestController) drawDeathOverlay(screen *ebiten.Image) {
	text := "YOU DIED"
	x := p.width/2 - 30
	y := p.height/2 - 10
	ebitenutil.DebugPrintAt(screen, text, x, y)
}

// drawCompleteOverlay shows level complete message.
func (p *PlaytestController) drawCompleteOverlay(screen *ebiten.Image) {
	text := "LEVEL COMPLETE!"
	x := p.width/2 - 50
	y := p.height/2 - 10
	ebitenutil.DebugPrintAt(screen, text, x, y)
}

// respawnPlayer resets player to respawn point.
func (p *PlaytestController) respawnPlayer() {
	p.playerBody.PosX = p.state.RespawnX
	p.playerBody.PosY = p.state.RespawnY
	p.playerBody.VelX = 0
	p.playerBody.VelY = 0
	p.state.FinishRespawn()
}

// resolveCollisions checks for tile collisions.
func (p *PlaytestController) resolveCollisions(aabb physics.AABB) []physics.Collision {
	var collisions []physics.Collision

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
	if endTX >= p.tileMap.Width() {
		endTX = p.tileMap.Width() - 1
	}
	if endTY >= p.tileMap.Height() {
		endTY = p.tileMap.Height() - 1
	}

	// Check each tile
	for ty := startTY; ty <= endTY; ty++ {
		for tx := startTX; tx <= endTX; tx++ {
			if p.collisionMap.IsSolidAtTile(tx, ty) {
				tileLeft := float64(tx * tileSize)
				tileRight := float64((tx + 1) * tileSize)
				tileTop := float64(ty * tileSize)
				tileBottom := float64((ty + 1) * tileSize)

				overlapLeft := (aabb.X + aabb.W) - tileLeft
				overlapRight := tileRight - aabb.X
				overlapTop := (aabb.Y + aabb.H) - tileTop
				overlapBottom := tileBottom - aabb.Y

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

	// Check solid entities
	for _, e := range p.entityWorld.SolidEntities() {
		entityBounds := e.Bounds()
		if entityBounds.W == 0 || entityBounds.H == 0 {
			continue
		}

		if aabb.Intersects(entityBounds) {
			overlapLeft := (aabb.X + aabb.W) - entityBounds.X
			overlapRight := (entityBounds.X + entityBounds.W) - aabb.X
			overlapTop := (aabb.Y + aabb.H) - entityBounds.Y
			overlapBottom := (entityBounds.Y + entityBounds.H) - aabb.Y

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

			var col physics.Collision
			col.TileX = -1
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
func (p *PlaytestController) resolveSolidEntityCollisions() {
	var solidAABBs []physics.AABB

	for _, e := range p.entityWorld.SolidEntities() {
		if !e.IsActive() {
			continue
		}
		bounds := e.Bounds()
		if bounds.W > 0 && bounds.H > 0 {
			solidAABBs = append(solidAABBs, bounds)
		}
	}

	for _, k := range p.entityWorld.GetKinematics() {
		if !k.IsActive() {
			continue
		}
		body := k.GetBody()
		if body != nil && body.W > 0 && body.H > 0 {
			solidAABBs = append(solidAABBs, body.AABB())
		}
	}

	if len(solidAABBs) > 0 {
		physics.ResolveSolids(p.playerBody, solidAABBs)
	}
}
