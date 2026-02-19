package rules

import (
	"fmt"
	"log"
)

// Action type constants.
const (
	ActionActivate   = "activate"
	ActionDeactivate = "deactivate"
	ActionToggle     = "toggle"
)

// ExecuteAction executes a single action spec.
func ExecuteAction(ctx ActionContext, spec ActionSpec) error {
	if ctx.Resolver == nil {
		return fmt.Errorf("no resolver in action context")
	}

	target := ctx.Resolver.Resolve(spec.Target)
	if target == nil {
		return fmt.Errorf("target not found: %s", spec.Target)
	}

	switch spec.Type {
	case ActionActivate:
		target.Activate()
		return nil
	case ActionDeactivate:
		target.Deactivate()
		return nil
	case ActionToggle:
		target.Toggle()
		return nil
	default:
		return fmt.Errorf("unknown action type: %s", spec.Type)
	}
}

// ExecuteActions executes multiple actions in sequence.
// If an action fails, it logs the error and continues with the next action.
func ExecuteActions(ctx ActionContext, specs []ActionSpec) {
	for _, spec := range specs {
		if err := ExecuteAction(ctx, spec); err != nil {
			log.Printf("[rules] action failed: %v (target=%s, type=%s)", err, spec.Target, spec.Type)
		}
	}
}
