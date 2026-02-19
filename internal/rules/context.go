package rules

// Targetable is the interface for entities that can be activated/deactivated.
// This mirrors the entities.Targetable interface to avoid import cycles.
type Targetable interface {
	Activate()
	Deactivate()
	Toggle()
	TargetID() string
}

// TargetResolver resolves target IDs to Targetable entities.
// This interface is implemented by entities.TargetRegistry.
type TargetResolver interface {
	// Resolve looks up a target by ID, returns nil if not found
	Resolve(id string) Targetable
}

// ActionContext provides context for action execution.
type ActionContext struct {
	// Event is the event that triggered this action
	Event Event
	// Resolver is used to look up target entities
	Resolver TargetResolver
	// Logf is an optional logging function
	Logf func(format string, args ...any)
}

// NewActionContext creates a new action context.
func NewActionContext(event Event, resolver TargetResolver) ActionContext {
	return ActionContext{
		Event:    event,
		Resolver: resolver,
	}
}
