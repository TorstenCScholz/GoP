package app

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/torsten/GoP/internal/input"
)

// Scene represents a game scene that can be active in the app.
type Scene interface {
	// Update updates the scene's logic.
	Update(inp *input.Input) error
	// Draw renders the scene to the screen.
	Draw(screen *ebiten.Image)
	// Layout returns the logical screen size.
	Layout(outsideW, outsideH int) (int, int)
	// DebugInfo returns debug information to display in the overlay.
	DebugInfo() string
}

// SceneDebugger is an optional interface that scenes can implement to draw debug visuals.
type SceneDebugger interface {
	// DrawDebug draws debug visuals on top of the scene.
	DrawDebug(screen *ebiten.Image)
}

// App is the main application struct that implements ebiten.Game.
type App struct {
	scene       Scene
	input       *input.Input
	config      *Config
	debugActive bool
}

// New creates a new App with the given configuration.
func New(cfg *Config) *App {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return &App{
		input:  input.NewInput(),
		config: cfg,
	}
}

// SetScene switches the current scene.
func (a *App) SetScene(scene Scene) {
	a.scene = scene
}

// Update implements ebiten.Game.Update.
func (a *App) Update() error {
	// Handle debug toggle
	if a.input.JustPressed(input.ActionDebugToggle) {
		a.debugActive = !a.debugActive
	}

	// Handle quit action
	if a.input.Pressed(input.ActionQuit) {
		return fmt.Errorf("quit requested")
	}

	// Delegate to current scene
	if a.scene != nil {
		if err := a.scene.Update(a.input); err != nil {
			return err
		}
	}

	// Update input state at the end of each frame to save previous key states
	a.input.Update()

	return nil
}

// Draw implements ebiten.Game.Draw.
func (a *App) Draw(screen *ebiten.Image) {
	// Delegate to current scene
	if a.scene != nil {
		a.scene.Draw(screen)
	}

	// Draw debug overlay on top
	if a.debugActive {
		a.drawDebugOverlay(screen)
	}
}

// Layout implements ebiten.Game.Layout.
func (a *App) Layout(outsideW, outsideH int) (int, int) {
	if a.scene != nil {
		return a.scene.Layout(outsideW, outsideH)
	}
	return a.config.WindowWidth, a.config.WindowHeight
}

// Run starts the game loop.
func (a *App) Run() error {
	ebiten.SetWindowSize(a.config.WindowWidth, a.config.WindowHeight)
	ebiten.SetWindowTitle(a.config.WindowTitle)
	return ebiten.RunGame(a)
}

// drawDebugOverlay renders the debug information overlay.
func (a *App) drawDebugOverlay(screen *ebiten.Image) {
	// Draw scene debug visuals if available
	if a.scene != nil {
		if debugger, ok := a.scene.(SceneDebugger); ok {
			debugger.DrawDebug(screen)
		}
	}

	fps := ebiten.CurrentFPS()
	tps := ebiten.CurrentTPS()
	w, h := screen.Size()

	debugText := fmt.Sprintf("FPS: %.1f\nTPS: %.1f\nWindow: %dx%d", fps, tps, w, h)

	// Add scene debug info if available
	if a.scene != nil {
		sceneInfo := a.scene.DebugInfo()
		if sceneInfo != "" {
			debugText += "\n" + sceneInfo
		}
	}

	ebitenutil.DebugPrint(screen, debugText)
}
