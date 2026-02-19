package rules

import (
	"fmt"
	"log"
)

// Engine processes events and executes matching rules.
type Engine struct {
	rules    []Rule
	resolver TargetResolver
	fired    map[string]bool // Tracks which "once" rules have fired
}

// NewEngine creates a new rule engine with the given target resolver.
func NewEngine(resolver TargetResolver) *Engine {
	return &Engine{
		rules:    make([]Rule, 0),
		resolver: resolver,
		fired:    make(map[string]bool),
	}
}

// LoadRules adds rules to the engine.
func (e *Engine) LoadRules(rules []Rule) {
	for i := range rules {
		// Default to active if not specified
		if !rules[i].Active {
			rules[i].Active = true
		}
	}
	e.rules = append(e.rules, rules...)
}

// LoadRuleSet loads all rules from a rule set.
func (e *Engine) LoadRuleSet(ruleSet RuleSet) {
	e.LoadRules(ruleSet.Rules)
}

// Clear removes all rules from the engine.
func (e *Engine) Clear() {
	e.rules = make([]Rule, 0)
	e.fired = make(map[string]bool)
}

// ProcessEvent checks all rules against the event and executes matching actions.
func (e *Engine) ProcessEvent(event Event) {
	ctx := NewActionContext(event, e.resolver)

	for i := range e.rules {
		rule := &e.rules[i]

		// Skip inactive rules
		if !rule.Active {
			continue
		}

		// Check if rule matches the event
		if !rule.matchesEvent(event) {
			continue
		}

		// Check if this is a "once" rule that already fired
		if rule.Once && e.fired[rule.ID] {
			continue
		}

		// Execute actions
		log.Printf("[rules] rule '%s' triggered by event '%s' from '%s'", rule.ID, event.Type, event.Source)
		ExecuteActions(ctx, rule.Actions)

		// Mark as fired if once rule
		if rule.Once {
			e.fired[rule.ID] = true
		}
	}
}

// Rules returns the current rules (for debugging).
func (e *Engine) Rules() []Rule {
	return e.rules
}

// RuleCount returns the number of loaded rules.
func (e *Engine) RuleCount() int {
	return len(e.rules)
}

// Stats returns engine statistics for debugging.
func (e *Engine) Stats() string {
	onceCount := 0
	for _, r := range e.rules {
		if r.Once {
			onceCount++
		}
	}
	return fmt.Sprintf("Rules: %d total, %d once-rules, %d fired", len(e.rules), onceCount, len(e.fired))
}
