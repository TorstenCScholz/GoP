// Package rules provides a data-driven rule system for entity interactions.
package rules

// Event represents a game event that can trigger rules.
type Event struct {
	// Type is the event type, e.g., "EnterRegion", "ExitRegion"
	Type string
	// Source is the entity ID that triggered the event
	Source string
	// Data contains additional event data
	Data map[string]any
}

// Common event types.
const (
	// EventEnterRegion is emitted when a player enters a trigger region
	EventEnterRegion = "EnterRegion"
	// EventExitRegion is emitted when a player exits a trigger region
	EventExitRegion = "ExitRegion"
	// EventDeath is emitted when the player dies
	EventDeath = "Death"
	// EventCheckpoint is emitted when a checkpoint is activated
	EventCheckpoint = "Checkpoint"
	// EventGoalReached is emitted when the player reaches the goal
	EventGoalReached = "GoalReached"
	// EventTimerExpired is emitted when a timer expires
	EventTimerExpired = "TimerExpired"
)

// NewEvent creates a new event with optional data.
func NewEvent(eventType, source string, data map[string]any) Event {
	if data == nil {
		data = make(map[string]any)
	}
	return Event{
		Type:   eventType,
		Source: source,
		Data:   data,
	}
}
