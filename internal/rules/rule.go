package rules

// WhenClause defines when a rule should trigger.
type WhenClause struct {
	// Event is the event type to match
	Event string `yaml:"event"`
	// Source is an optional specific source ID to match
	// If empty, matches any source
	Source string `yaml:"source,omitempty"`
}

// ActionSpec defines an action to execute when a rule triggers.
type ActionSpec struct {
	// Type is the action type: activate, deactivate, toggle
	Type string `yaml:"type"`
	// Target is the target entity ID
	Target string `yaml:"target"`
	// Params contains optional action parameters
	Params map[string]any `yaml:"params,omitempty"`
}

// Rule is a single rule definition.
type Rule struct {
	// ID is the unique identifier for this rule
	ID string `yaml:"id"`
	// When defines the conditions for this rule to trigger
	When WhenClause `yaml:"when"`
	// Actions is the list of actions to execute when triggered
	Actions []ActionSpec `yaml:"actions"`
	// Once indicates if this rule should only fire once
	Once bool `yaml:"once,omitempty"`
	// Active indicates if this rule is enabled
	Active bool `yaml:"active,omitempty"`
}

// RuleSet is a collection of rules loaded from a file.
type RuleSet struct {
	// Rules is the list of rules in this set
	Rules []Rule `yaml:"rules"`
}

// matchesEvent checks if this rule matches the given event.
func (r *Rule) matchesEvent(event Event) bool {
	// Check if rule is active (default to true if not specified)
	if !r.Active {
		return false
	}

	// Check event type
	if r.When.Event != event.Type {
		return false
	}

	// Check source if specified
	if r.When.Source != "" && r.When.Source != event.Source {
		return false
	}

	return true
}
