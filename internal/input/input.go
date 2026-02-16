// Package input provides an action-based input abstraction layer.
package input

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Action represents a game action that can be triggered by user input.
type Action int

const (
	ActionMoveLeft Action = iota
	ActionMoveRight
	ActionMoveUp
	ActionMoveDown
	ActionJump
	ActionQuit
	ActionDebugToggle
)

// Input manages keyboard input with action mappings.
type Input struct {
	keyMap      map[Action][]ebiten.Key
	prevPressed map[ebiten.Key]bool
}

// NewInput creates a new Input manager with default key mappings.
func NewInput() *Input {
	i := &Input{
		keyMap:      make(map[Action][]ebiten.Key),
		prevPressed: make(map[ebiten.Key]bool),
	}

	// Default key mappings
	// Movement: Arrow keys and WASD
	i.keyMap[ActionMoveLeft] = []ebiten.Key{ebiten.KeyArrowLeft, ebiten.KeyA}
	i.keyMap[ActionMoveRight] = []ebiten.Key{ebiten.KeyArrowRight, ebiten.KeyD}
	i.keyMap[ActionMoveUp] = []ebiten.Key{ebiten.KeyArrowUp, ebiten.KeyW}
	i.keyMap[ActionMoveDown] = []ebiten.Key{ebiten.KeyArrowDown, ebiten.KeyS}
	i.keyMap[ActionJump] = []ebiten.Key{ebiten.KeySpace, ebiten.KeyZ}
	i.keyMap[ActionQuit] = []ebiten.Key{ebiten.KeyEscape}
	i.keyMap[ActionDebugToggle] = []ebiten.Key{ebiten.KeyF1}

	return i
}

// Pressed returns true if any key mapped to the action is currently pressed.
func (i *Input) Pressed(action Action) bool {
	keys, ok := i.keyMap[action]
	if !ok {
		println("[DEBUG INPUT] No keys mapped for action:", int(action))
		return false
	}

	// DEBUG: Check each key
	for _, key := range keys {
		pressed := ebiten.IsKeyPressed(key)
		if pressed {
			println("[DEBUG INPUT] Key", int(key), "is pressed for action", int(action))
			return true
		}
	}
	return false
}

// JustPressed returns true if any key mapped to the action was just pressed this frame.
func (i *Input) JustPressed(action Action) bool {
	keys, ok := i.keyMap[action]
	if !ok {
		return false
	}

	for _, key := range keys {
		if ebiten.IsKeyPressed(key) && !i.prevPressed[key] {
			return true
		}
	}
	return false
}

// Update updates the previous frame's key states.
// This should be called once per frame, typically at the start of the game loop.
func (i *Input) Update() {
	// Clear previous pressed state and update with current state
	for _, keys := range i.keyMap {
		for _, key := range keys {
			i.prevPressed[key] = ebiten.IsKeyPressed(key)
		}
	}
}
