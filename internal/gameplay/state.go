// Package gameplay provides game state management for levels.
package gameplay

import (
	"github.com/torsten/GoP/internal/physics"
)

// State represents the current gameplay state.
type State int

const (
	// StateRunning is the normal gameplay state.
	StateRunning State = iota
	// StateDead is the state after player dies, before respawn.
	StateDead
	// StateRespawning is the state during respawn transition.
	StateRespawning
	// StateCompleted is the state after reaching the goal.
	StateCompleted
)

// String returns the state name for debugging.
func (s State) String() string {
	switch s {
	case StateRunning:
		return "Running"
	case StateDead:
		return "Dead"
	case StateRespawning:
		return "Respawning"
	case StateCompleted:
		return "Completed"
	default:
		return "Unknown"
	}
}

// StateMachine manages gameplay state transitions.
type StateMachine struct {
	// Current state
	Current State

	// Respawn point (top-left corner of player)
	RespawnX float64
	RespawnY float64

	// Death animation timing
	DeathTimer   float64 // Time since death
	RespawnDelay float64 // Delay before respawn (seconds)

	// Level completion callback
	OnComplete func()
}

// NewStateMachine creates a new state machine with default settings.
func NewStateMachine() *StateMachine {
	return &StateMachine{
		Current:      StateRunning,
		RespawnDelay: 1.0, // 1 second delay before respawn
	}
}

// Update processes state transitions.
func (sm *StateMachine) Update(dt float64) {
	switch sm.Current {
	case StateDead:
		sm.DeathTimer += dt
		if sm.DeathTimer >= sm.RespawnDelay {
			sm.Current = StateRespawning
		}
	case StateRespawning:
		// Respawn is handled by the scene
		// Scene will call FinishRespawn() after positioning player
	}
}

// TriggerDeath initiates death state.
func (sm *StateMachine) TriggerDeath() {
	if sm.Current == StateRunning {
		sm.Current = StateDead
		sm.DeathTimer = 0
	}
}

// TriggerComplete initiates level completion.
func (sm *StateMachine) TriggerComplete() {
	if sm.Current == StateRunning {
		sm.Current = StateCompleted
		if sm.OnComplete != nil {
			sm.OnComplete()
		}
	}
}

// FinishRespawn completes the respawn and returns to running state.
func (sm *StateMachine) FinishRespawn() {
	sm.Current = StateRunning
	sm.DeathTimer = 0
}

// SetRespawnPoint updates the respawn location.
func (sm *StateMachine) SetRespawnPoint(x, y float64) {
	sm.RespawnX = x
	sm.RespawnY = y
}

// IsRunning returns true if in normal gameplay state.
func (sm *StateMachine) IsRunning() bool {
	return sm.Current == StateRunning
}

// IsDead returns true if in death state.
func (sm *StateMachine) IsDead() bool {
	return sm.Current == StateDead
}

// IsRespawning returns true if in respawning state.
func (sm *StateMachine) IsRespawning() bool {
	return sm.Current == StateRespawning
}

// IsCompleted returns true if level is completed.
func (sm *StateMachine) IsCompleted() bool {
	return sm.Current == StateCompleted
}

// GetRespawnAABB returns the respawn point as an AABB.
func (sm *StateMachine) GetRespawnAABB(w, h float64) physics.AABB {
	return physics.AABB{
		X: sm.RespawnX,
		Y: sm.RespawnY,
		W: w,
		H: h,
	}
}
