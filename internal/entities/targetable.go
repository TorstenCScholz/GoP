package entities

// Targetable represents an entity that can be activated/deactivated by switches, etc.
type Targetable interface {
	// Activate triggers the target (e.g., open a door)
	Activate()
	// Deactivate untriggers the target (e.g., close a door)
	Deactivate()
	// Toggle flips the target state
	Toggle()
	// TargetID returns the unique identifier for this target
	TargetID() string
}

// TargetRegistry manages ID-to-target lookups.
// It provides a decoupled way for switches and other triggers to find their targets.
type TargetRegistry struct {
	targets map[string]Targetable
}

// NewTargetRegistry creates a new registry.
func NewTargetRegistry() *TargetRegistry {
	return &TargetRegistry{
		targets: make(map[string]Targetable),
	}
}

// Register adds a target to the registry.
// If a target with the same ID already exists, it will be overwritten.
func (r *TargetRegistry) Register(t Targetable) {
	if t == nil {
		return
	}
	id := t.TargetID()
	if id == "" {
		return
	}
	r.targets[id] = t
}

// Unregister removes a target from the registry.
func (r *TargetRegistry) Unregister(t Targetable) {
	if t == nil {
		return
	}
	id := t.TargetID()
	if id == "" {
		return
	}
	delete(r.targets, id)
}

// Resolve looks up a target by ID, returns nil if not found.
func (r *TargetRegistry) Resolve(id string) Targetable {
	if id == "" {
		return nil
	}
	return r.targets[id]
}

// HasTarget returns true if a target with the given ID exists.
func (r *TargetRegistry) HasTarget(id string) bool {
	_, exists := r.targets[id]
	return exists
}

// AllTargets returns a slice of all registered targets.
// This is useful for debugging or iteration.
func (r *TargetRegistry) AllTargets() []Targetable {
	result := make([]Targetable, 0, len(r.targets))
	for _, t := range r.targets {
		result = append(result, t)
	}
	return result
}
