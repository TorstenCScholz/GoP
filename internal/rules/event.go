// Package rules provides a data-driven rule system for entity interactions.
package rules

// EventType defines the type of game event.
type EventType string

// MVP event types.
const (
	// EventEnterRegion is emitted when a player enters a trigger region
	EventEnterRegion EventType = "enter_region"
	// EventExitRegion is emitted when a player exits a trigger region
	EventExitRegion EventType = "exit_region"
)

// Event represents a game event that can trigger rules.
type Event struct {
	// Type is the event type
	Type EventType
	// RegionID is the region/trigger ID (e.g., "switch_A")
	RegionID string
	// ActorType is the actor type: "player", "enemy", etc. (MVP: only "player")
	ActorType string
}

// NewEvent creates a new event with the given parameters.
func NewEvent(typ EventType, regionID, actorType string) Event {
	return Event{
		Type:      typ,
		RegionID:  regionID,
		ActorType: actorType,
	}
}
