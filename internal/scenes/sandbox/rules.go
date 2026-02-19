package sandbox

import (
	"github.com/torsten/GoP/internal/entities"
	"github.com/torsten/GoP/internal/rules"
)

// targetResolver adapts entities.TargetRegistry to rules.TargetResolver.
type targetResolver struct {
	registry *entities.TargetRegistry
}

// newTargetResolver creates a new adapter for the given registry.
func newTargetResolver(registry *entities.TargetRegistry) *targetResolver {
	return &targetResolver{registry: registry}
}

// Resolve implements rules.TargetResolver.
func (r *targetResolver) Resolve(id string) rules.Targetable {
	target := r.registry.Resolve(id)
	if target == nil {
		return nil
	}
	// Return a wrapper that adapts entities.Targetable to rules.Targetable
	return &targetableAdapter{target: target}
}

// targetableAdapter adapts entities.Targetable to rules.Targetable.
type targetableAdapter struct {
	target entities.Targetable
}

// Activate implements rules.Targetable.
func (a *targetableAdapter) Activate() {
	a.target.Activate()
}

// Deactivate implements rules.Targetable.
func (a *targetableAdapter) Deactivate() {
	a.target.Deactivate()
}

// Toggle implements rules.Targetable.
func (a *targetableAdapter) Toggle() {
	a.target.Toggle()
}

// TargetID implements rules.Targetable.
func (a *targetableAdapter) TargetID() string {
	return a.target.TargetID()
}
