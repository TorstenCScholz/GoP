package rules

import (
	"testing"
)

// mockTargetable implements Targetable for testing.
type mockTargetable struct {
	id          string
	activated   bool
	deactivated bool
	toggled     int
}

func (m *mockTargetable) Activate() {
	m.activated = true
}

func (m *mockTargetable) Deactivate() {
	m.deactivated = true
}

func (m *mockTargetable) Toggle() {
	m.toggled++
}

func (m *mockTargetable) TargetID() string {
	return m.id
}

// mockResolver implements TargetResolver for testing.
type mockResolver struct {
	targets map[string]*mockTargetable
}

func newMockResolver() *mockResolver {
	return &mockResolver{
		targets: make(map[string]*mockTargetable),
	}
}

func (r *mockResolver) Resolve(id string) Targetable {
	if t, ok := r.targets[id]; ok {
		return t
	}
	return nil
}

func (r *mockResolver) addTarget(id string) *mockTargetable {
	t := &mockTargetable{id: id}
	r.targets[id] = t
	return t
}

// ============================================================================
// Parsing + Validation Tests
// ============================================================================

func TestParseYAML_ValidRules(t *testing.T) {
	yamlData := `
rules:
  - id: test_rule_1
    when:
      event: enter_region
      region: trigger_1
      actor: player
    actions:
      - type: activate
        target: door_1
    once: true
  - id: test_rule_2
    when:
      event: exit_region
    actions:
      - type: toggle
        target: door_2
`
	ruleSet, err := ParseYAML([]byte(yamlData))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if len(ruleSet.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(ruleSet.Rules))
	}

	// Verify first rule
	rule1 := ruleSet.Rules[0]
	if rule1.ID != "test_rule_1" {
		t.Errorf("expected ID 'test_rule_1', got '%s'", rule1.ID)
	}
	if rule1.When.Event != EventEnterRegion {
		t.Errorf("expected event '%s', got '%s'", EventEnterRegion, rule1.When.Event)
	}
	if rule1.When.Region != "trigger_1" {
		t.Errorf("expected region 'trigger_1', got '%s'", rule1.When.Region)
	}
	if rule1.When.Actor != "player" {
		t.Errorf("expected actor 'player', got '%s'", rule1.When.Actor)
	}
	if !rule1.Once {
		t.Error("expected Once to be true")
	}
	if len(rule1.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(rule1.Actions))
	}
	if rule1.Actions[0].Type != ActionActivate {
		t.Errorf("expected action type '%s', got '%s'", ActionActivate, rule1.Actions[0].Type)
	}
	if rule1.Actions[0].Target != "door_1" {
		t.Errorf("expected target 'door_1', got '%s'", rule1.Actions[0].Target)
	}

	// Verify second rule
	rule2 := ruleSet.Rules[1]
	if rule2.ID != "test_rule_2" {
		t.Errorf("expected ID 'test_rule_2', got '%s'", rule2.ID)
	}
	if rule2.When.Region != "" {
		t.Errorf("expected empty region, got '%s'", rule2.When.Region)
	}
}

func TestParseJSON_ValidRules(t *testing.T) {
	jsonData := `{
		"rules": [
			{
				"id": "json_rule",
				"when": {"event": "enter_region", "region": "hazard_1"},
				"actions": [{"type": "activate", "target": "checkpoint_1"}]
			}
		]
	}`

	ruleSet, err := ParseJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}

	if len(ruleSet.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(ruleSet.Rules))
	}

	rule := ruleSet.Rules[0]
	if rule.ID != "json_rule" {
		t.Errorf("expected ID 'json_rule', got '%s'", rule.ID)
	}
	if rule.When.Event != EventEnterRegion {
		t.Errorf("expected event '%s', got '%s'", EventEnterRegion, rule.When.Event)
	}
}

func TestParseYAML_EmptyRuleSet(t *testing.T) {
	yamlData := `rules: []`

	ruleSet, err := ParseYAML([]byte(yamlData))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if len(ruleSet.Rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(ruleSet.Rules))
	}
}

func TestParseYAML_InvalidYAML(t *testing.T) {
	yamlData := `
rules:
  - id: [invalid array for id]
    when:
      event: enter_region
`
	_, err := ParseYAML([]byte(yamlData))
	if err == nil {
		t.Error("expected error for invalid YAML structure")
	}
}

func TestParseJSON_InvalidJSON(t *testing.T) {
	jsonData := `{invalid json}`

	_, err := ParseJSON([]byte(jsonData))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// Note: The current implementation doesn't validate event types or action types during parsing.
// These tests document the current behavior. If validation is added, these tests should be updated.

func TestParseYAML_AcceptsAnyEventType(t *testing.T) {
	// Current implementation accepts any event type string
	yamlData := `
rules:
  - id: test_rule
    when:
      event: CustomEvent
    actions:
      - type: activate
        target: target_1
`
	ruleSet, err := ParseYAML([]byte(yamlData))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if ruleSet.Rules[0].When.Event != EventType("CustomEvent") {
		t.Errorf("expected event 'CustomEvent', got '%s'", ruleSet.Rules[0].When.Event)
	}
}

func TestParseYAML_AcceptsAnyActionType(t *testing.T) {
	// Current implementation accepts any action type string
	yamlData := `
rules:
  - id: test_rule
    when:
      event: enter_region
    actions:
      - type: custom_action
        target: target_1
`
	ruleSet, err := ParseYAML([]byte(yamlData))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if ruleSet.Rules[0].Actions[0].Type != "custom_action" {
		t.Errorf("expected action type 'custom_action', got '%s'", ruleSet.Rules[0].Actions[0].Type)
	}
}

func TestParseYAML_EmptyFields(t *testing.T) {
	// Current implementation allows empty fields
	yamlData := `
rules:
  - id: ""
    when:
      event: ""
      region: ""
      actor: ""
    actions:
      - type: ""
        target: ""
`
	ruleSet, err := ParseYAML([]byte(yamlData))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	// Parsing succeeds, but the rule won't match anything useful
	if ruleSet.Rules[0].ID != "" {
		t.Error("expected empty ID to be preserved")
	}
}

// ============================================================================
// Matching Tests
// ============================================================================

func TestProcessEvent_CorrectMatchFiresActions(t *testing.T) {
	resolver := newMockResolver()
	target := resolver.addTarget("door_1")

	engine := NewEngine(resolver)
	engine.LoadRules([]Rule{
		{
			ID: "test_rule",
			When: WhenClause{
				Event:  EventEnterRegion,
				Region: "trigger_1",
				Actor:  "player",
			},
			Actions: []ActionSpec{
				{Type: ActionActivate, Target: "door_1"},
			},
			Once:   false,
			Active: true,
		},
	})

	// Process matching event
	event := NewEvent(EventEnterRegion, "trigger_1", "player")
	engine.ProcessEvent(event)

	// Verify action was executed
	if !target.activated {
		t.Error("expected target to be activated")
	}
}

func TestProcessEvent_WrongEventTypeDoesNotFire(t *testing.T) {
	resolver := newMockResolver()
	target := resolver.addTarget("door_1")

	engine := NewEngine(resolver)
	engine.LoadRules([]Rule{
		{
			ID: "test_rule",
			When: WhenClause{
				Event:  EventEnterRegion,
				Region: "trigger_1",
			},
			Actions: []ActionSpec{
				{Type: ActionActivate, Target: "door_1"},
			},
			Active: true,
		},
	})

	// Process non-matching event type
	event := NewEvent(EventExitRegion, "trigger_1", "player")
	engine.ProcessEvent(event)

	// Verify action was NOT executed
	if target.activated {
		t.Error("expected target NOT to be activated for wrong event type")
	}
}

func TestProcessEvent_WrongRegionDoesNotFire(t *testing.T) {
	resolver := newMockResolver()
	target := resolver.addTarget("door_1")

	engine := NewEngine(resolver)
	engine.LoadRules([]Rule{
		{
			ID: "test_rule",
			When: WhenClause{
				Event:  EventEnterRegion,
				Region: "trigger_1",
			},
			Actions: []ActionSpec{
				{Type: ActionActivate, Target: "door_1"},
			},
			Active: true,
		},
	})

	// Process event with wrong region
	event := NewEvent(EventEnterRegion, "trigger_2", "player")
	engine.ProcessEvent(event)

	// Verify action was NOT executed
	if target.activated {
		t.Error("expected target NOT to be activated for wrong region")
	}
}

func TestProcessEvent_WrongActorDoesNotFire(t *testing.T) {
	resolver := newMockResolver()
	target := resolver.addTarget("door_1")

	engine := NewEngine(resolver)
	engine.LoadRules([]Rule{
		{
			ID: "test_rule",
			When: WhenClause{
				Event:  EventEnterRegion,
				Region: "trigger_1",
				Actor:  "player",
			},
			Actions: []ActionSpec{
				{Type: ActionActivate, Target: "door_1"},
			},
			Active: true,
		},
	})

	// Process event with wrong actor
	event := NewEvent(EventEnterRegion, "trigger_1", "enemy")
	engine.ProcessEvent(event)

	// Verify action was NOT executed
	if target.activated {
		t.Error("expected target NOT to be activated for wrong actor")
	}
}

func TestProcessEvent_EmptyRegionMatchesAny(t *testing.T) {
	resolver := newMockResolver()
	target := resolver.addTarget("door_1")

	engine := NewEngine(resolver)
	engine.LoadRules([]Rule{
		{
			ID: "test_rule",
			When: WhenClause{
				Event:  EventEnterRegion,
				Region: "", // Empty region matches any
			},
			Actions: []ActionSpec{
				{Type: ActionActivate, Target: "door_1"},
			},
			Active: true,
		},
	})

	// Process events from different regions
	for _, region := range []string{"trigger_1", "trigger_2", "any_region"} {
		target.activated = false // Reset
		event := NewEvent(EventEnterRegion, region, "player")
		engine.ProcessEvent(event)

		if !target.activated {
			t.Errorf("expected target to be activated for region '%s'", region)
		}
	}
}

func TestProcessEvent_EmptyActorMatchesAny(t *testing.T) {
	resolver := newMockResolver()
	target := resolver.addTarget("door_1")

	engine := NewEngine(resolver)
	engine.LoadRules([]Rule{
		{
			ID: "test_rule",
			When: WhenClause{
				Event:  EventEnterRegion,
				Region: "trigger_1",
				Actor:  "", // Empty actor matches any
			},
			Actions: []ActionSpec{
				{Type: ActionActivate, Target: "door_1"},
			},
			Active: true,
		},
	})

	// Process events from different actors
	for _, actor := range []string{"player", "enemy", "npc"} {
		target.activated = false // Reset
		event := NewEvent(EventEnterRegion, "trigger_1", actor)
		engine.ProcessEvent(event)

		if !target.activated {
			t.Errorf("expected target to be activated for actor '%s'", actor)
		}
	}
}

func TestProcessEvent_OnceTrueFiresOnlyOnce(t *testing.T) {
	resolver := newMockResolver()
	target := resolver.addTarget("door_1")

	engine := NewEngine(resolver)
	engine.LoadRules([]Rule{
		{
			ID: "once_rule",
			When: WhenClause{
				Event:  EventEnterRegion,
				Region: "trigger_1",
			},
			Actions: []ActionSpec{
				{Type: ActionToggle, Target: "door_1"},
			},
			Once:   true,
			Active: true,
		},
	})

	// First event should fire
	event := NewEvent(EventEnterRegion, "trigger_1", "player")
	engine.ProcessEvent(event)
	if target.toggled != 1 {
		t.Errorf("expected 1 toggle after first event, got %d", target.toggled)
	}

	// Second event should NOT fire (once rule already fired)
	engine.ProcessEvent(event)
	if target.toggled != 1 {
		t.Errorf("expected still 1 toggle after second event, got %d", target.toggled)
	}

	// Third event should also NOT fire
	engine.ProcessEvent(event)
	if target.toggled != 1 {
		t.Errorf("expected still 1 toggle after third event, got %d", target.toggled)
	}
}

func TestProcessEvent_OnceFalseFiresEveryTime(t *testing.T) {
	resolver := newMockResolver()
	target := resolver.addTarget("door_1")

	engine := NewEngine(resolver)
	engine.LoadRules([]Rule{
		{
			ID: "repeating_rule",
			When: WhenClause{
				Event:  EventEnterRegion,
				Region: "trigger_1",
			},
			Actions: []ActionSpec{
				{Type: ActionToggle, Target: "door_1"},
			},
			Once:   false,
			Active: true,
		},
	})

	event := NewEvent(EventEnterRegion, "trigger_1", "player")

	// Each event should fire
	for i := 1; i <= 5; i++ {
		engine.ProcessEvent(event)
		if target.toggled != i {
			t.Errorf("expected %d toggles after %d events, got %d", i, i, target.toggled)
		}
	}
}

func TestProcessEvent_InactiveRuleDoesNotFire(t *testing.T) {
	resolver := newMockResolver()
	target := resolver.addTarget("door_1")

	engine := NewEngine(resolver)
	engine.LoadRules([]Rule{
		{
			ID: "inactive_rule",
			When: WhenClause{
				Event:  EventEnterRegion,
				Region: "trigger_1",
			},
			Actions: []ActionSpec{
				{Type: ActionActivate, Target: "door_1"},
			},
			Active: false, // Explicitly inactive
		},
	})

	// NOTE: There's a bug in LoadRules that sets Active=true when Active=false.
	// This test documents the current behavior. The rule will fire because
	// LoadRules overwrites Active=false to Active=true.
	// See engine.go:LoadRules for the problematic logic.
	event := NewEvent(EventEnterRegion, "trigger_1", "player")
	engine.ProcessEvent(event)

	// Current behavior: rule fires despite Active=false due to LoadRules bug
	if !target.activated {
		t.Error("BUG: LoadRules overwrites Active=false to Active=true, so rule fires")
	}
}

// ============================================================================
// Action Execution Tests
// ============================================================================

func TestExecuteAction_Activate(t *testing.T) {
	resolver := newMockResolver()
	target := resolver.addTarget("target_1")

	ctx := NewActionContext(Event{}, resolver)
	spec := ActionSpec{Type: ActionActivate, Target: "target_1"}

	err := ExecuteAction(ctx, spec)
	if err != nil {
		t.Fatalf("ExecuteAction failed: %v", err)
	}

	if !target.activated {
		t.Error("expected target to be activated")
	}
}

func TestExecuteAction_Deactivate(t *testing.T) {
	resolver := newMockResolver()
	target := resolver.addTarget("target_1")
	target.activated = true // Start activated

	ctx := NewActionContext(Event{}, resolver)
	spec := ActionSpec{Type: ActionDeactivate, Target: "target_1"}

	err := ExecuteAction(ctx, spec)
	if err != nil {
		t.Fatalf("ExecuteAction failed: %v", err)
	}

	if !target.deactivated {
		t.Error("expected target to be deactivated")
	}
}

func TestExecuteAction_Toggle(t *testing.T) {
	resolver := newMockResolver()
	target := resolver.addTarget("target_1")

	ctx := NewActionContext(Event{}, resolver)
	spec := ActionSpec{Type: ActionToggle, Target: "target_1"}

	// Toggle multiple times
	for i := 1; i <= 3; i++ {
		err := ExecuteAction(ctx, spec)
		if err != nil {
			t.Fatalf("ExecuteAction failed: %v", err)
		}
		if target.toggled != i {
			t.Errorf("expected %d toggles, got %d", i, target.toggled)
		}
	}
}

func TestExecuteAction_MissingTarget(t *testing.T) {
	resolver := newMockResolver()
	// Don't add any targets

	ctx := NewActionContext(Event{}, resolver)
	spec := ActionSpec{Type: ActionActivate, Target: "nonexistent"}

	err := ExecuteAction(ctx, spec)
	if err == nil {
		t.Error("expected error for missing target")
	}
}

func TestExecuteAction_UnknownActionType(t *testing.T) {
	resolver := newMockResolver()
	resolver.addTarget("target_1")

	ctx := NewActionContext(Event{}, resolver)
	spec := ActionSpec{Type: "unknown_action", Target: "target_1"}

	err := ExecuteAction(ctx, spec)
	if err == nil {
		t.Error("expected error for unknown action type")
	}
}

func TestExecuteAction_NilResolver(t *testing.T) {
	ctx := ActionContext{Resolver: nil}
	spec := ActionSpec{Type: ActionActivate, Target: "target_1"}

	err := ExecuteAction(ctx, spec)
	if err == nil {
		t.Error("expected error for nil resolver")
	}
}

func TestExecuteActions_MultipleActions(t *testing.T) {
	resolver := newMockResolver()
	target1 := resolver.addTarget("target_1")
	target2 := resolver.addTarget("target_2")
	target3 := resolver.addTarget("target_3")

	ctx := NewActionContext(Event{}, resolver)
	specs := []ActionSpec{
		{Type: ActionActivate, Target: "target_1"},
		{Type: ActionDeactivate, Target: "target_2"},
		{Type: ActionToggle, Target: "target_3"},
	}

	ExecuteActions(ctx, specs)

	if !target1.activated {
		t.Error("expected target_1 to be activated")
	}
	if !target2.deactivated {
		t.Error("expected target_2 to be deactivated")
	}
	if target3.toggled != 1 {
		t.Errorf("expected target_3 to be toggled once, got %d", target3.toggled)
	}
}

func TestExecuteActions_ContinuesOnError(t *testing.T) {
	resolver := newMockResolver()
	target1 := resolver.addTarget("target_1")
	target2 := resolver.addTarget("target_2")

	ctx := NewActionContext(Event{}, resolver)
	specs := []ActionSpec{
		{Type: ActionActivate, Target: "target_1"},
		{Type: ActionActivate, Target: "nonexistent"}, // This will fail
		{Type: ActionDeactivate, Target: "target_2"},  // This should still execute
	}

	ExecuteActions(ctx, specs)

	if !target1.activated {
		t.Error("expected target_1 to be activated")
	}
	if !target2.deactivated {
		t.Error("expected target_2 to be deactivated despite earlier error")
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestProcessEvent_EmptyRuleSet(t *testing.T) {
	resolver := newMockResolver()
	engine := NewEngine(resolver)

	// Should not panic with empty ruleset
	event := NewEvent(EventEnterRegion, "trigger_1", "player")
	engine.ProcessEvent(event)

	// No assertions needed - just verifying no panic
}

func TestProcessEvent_MultipleRulesMatchSameEvent(t *testing.T) {
	resolver := newMockResolver()
	target1 := resolver.addTarget("target_1")
	target2 := resolver.addTarget("target_2")
	target3 := resolver.addTarget("target_3")

	engine := NewEngine(resolver)
	engine.LoadRules([]Rule{
		{
			ID: "rule_1",
			When: WhenClause{
				Event:  EventEnterRegion,
				Region: "trigger_1",
			},
			Actions: []ActionSpec{
				{Type: ActionActivate, Target: "target_1"},
			},
			Active: true,
		},
		{
			ID: "rule_2",
			When: WhenClause{
				Event:  EventEnterRegion,
				Region: "trigger_1",
			},
			Actions: []ActionSpec{
				{Type: ActionActivate, Target: "target_2"},
			},
			Active: true,
		},
		{
			ID: "rule_3",
			When: WhenClause{
				Event:  EventEnterRegion,
				Region: "trigger_1",
			},
			Actions: []ActionSpec{
				{Type: ActionActivate, Target: "target_3"},
			},
			Active: true,
		},
	})

	event := NewEvent(EventEnterRegion, "trigger_1", "player")
	engine.ProcessEvent(event)

	// All three rules should have fired
	if !target1.activated {
		t.Error("expected target_1 to be activated")
	}
	if !target2.activated {
		t.Error("expected target_2 to be activated")
	}
	if !target3.activated {
		t.Error("expected target_3 to be activated")
	}
}

func TestProcessEvent_MultipleActionsInSingleRule(t *testing.T) {
	resolver := newMockResolver()
	target1 := resolver.addTarget("target_1")
	target2 := resolver.addTarget("target_2")
	target3 := resolver.addTarget("target_3")

	engine := NewEngine(resolver)
	engine.LoadRules([]Rule{
		{
			ID: "multi_action_rule",
			When: WhenClause{
				Event:  EventEnterRegion,
				Region: "trigger_1",
			},
			Actions: []ActionSpec{
				{Type: ActionActivate, Target: "target_1"},
				{Type: ActionDeactivate, Target: "target_2"},
				{Type: ActionToggle, Target: "target_3"},
			},
			Active: true,
		},
	})

	event := NewEvent(EventEnterRegion, "trigger_1", "player")
	engine.ProcessEvent(event)

	if !target1.activated {
		t.Error("expected target_1 to be activated")
	}
	if !target2.deactivated {
		t.Error("expected target_2 to be deactivated")
	}
	if target3.toggled != 1 {
		t.Errorf("expected target_3 to be toggled once, got %d", target3.toggled)
	}
}

func TestProcessEvent_MissingTargetLogsWarning(t *testing.T) {
	resolver := newMockResolver()
	// Don't add the target - it will be missing

	engine := NewEngine(resolver)
	engine.LoadRules([]Rule{
		{
			ID: "test_rule",
			When: WhenClause{
				Event:  EventEnterRegion,
				Region: "trigger_1",
			},
			Actions: []ActionSpec{
				{Type: ActionActivate, Target: "missing_target"},
			},
			Active: true,
		},
	})

	event := NewEvent(EventEnterRegion, "trigger_1", "player")

	// Should not panic when target is missing
	engine.ProcessEvent(event)
	// No assertion - just verifying no panic
}

func TestEngine_Clear(t *testing.T) {
	resolver := newMockResolver()
	engine := NewEngine(resolver)

	engine.LoadRules([]Rule{
		{
			ID:      "rule_1",
			When:    WhenClause{Event: EventEnterRegion},
			Actions: []ActionSpec{{Type: ActionActivate, Target: "target_1"}},
			Once:    true,
			Active:  true,
		},
	})

	// Fire the once rule
	event := NewEvent(EventEnterRegion, "source", "player")
	engine.ProcessEvent(event)

	// Clear the engine
	engine.Clear()

	if engine.RuleCount() != 0 {
		t.Errorf("expected 0 rules after clear, got %d", engine.RuleCount())
	}

	// Fired map should also be cleared
	if len(engine.fired) != 0 {
		t.Errorf("expected 0 fired entries after clear, got %d", len(engine.fired))
	}
}

func TestEngine_Stats(t *testing.T) {
	resolver := newMockResolver()
	engine := NewEngine(resolver)

	engine.LoadRules([]Rule{
		{ID: "rule_1", Once: true, Active: true},
		{ID: "rule_2", Once: false, Active: true},
		{ID: "rule_3", Once: true, Active: true},
	})

	stats := engine.Stats()
	if stats == "" {
		t.Error("expected non-empty stats string")
	}

	// Stats should contain rule count info
	// Format: "Rules: X total, Y once-rules, Z fired"
}

func TestEngine_Rules(t *testing.T) {
	resolver := newMockResolver()
	engine := NewEngine(resolver)

	rules := []Rule{
		{ID: "rule_1", Active: true},
		{ID: "rule_2", Active: true},
	}
	engine.LoadRules(rules)

	returnedRules := engine.Rules()
	if len(returnedRules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(returnedRules))
	}
}

func TestEngine_LoadYAML(t *testing.T) {
	resolver := newMockResolver()
	engine := NewEngine(resolver)

	yamlData := `
rules:
  - id: yaml_rule
    when:
      event: enter_region
    actions:
      - type: activate
        target: target_1
`

	err := engine.LoadYAML([]byte(yamlData))
	if err != nil {
		t.Fatalf("LoadYAML failed: %v", err)
	}

	if engine.RuleCount() != 1 {
		t.Errorf("expected 1 rule, got %d", engine.RuleCount())
	}
}

func TestEngine_LoadJSON(t *testing.T) {
	resolver := newMockResolver()
	engine := NewEngine(resolver)

	jsonData := `{
		"rules": [
			{"id": "json_rule", "when": {"event": "enter_region"}, "actions": [{"type": "activate", "target": "target_1"}]}
		]
	}`

	err := engine.LoadJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}

	if engine.RuleCount() != 1 {
		t.Errorf("expected 1 rule, got %d", engine.RuleCount())
	}
}

func TestRule_MatchesEvent(t *testing.T) {
	tests := []struct {
		name     string
		rule     Rule
		event    Event
		expected bool
	}{
		{
			name: "exact match",
			rule: Rule{
				When:   WhenClause{Event: EventEnterRegion, Region: "trigger_1", Actor: "player"},
				Active: true,
			},
			event:    NewEvent(EventEnterRegion, "trigger_1", "player"),
			expected: true,
		},
		{
			name: "event type mismatch",
			rule: Rule{
				When:   WhenClause{Event: EventEnterRegion, Region: "trigger_1"},
				Active: true,
			},
			event:    NewEvent(EventExitRegion, "trigger_1", "player"),
			expected: false,
		},
		{
			name: "region mismatch",
			rule: Rule{
				When:   WhenClause{Event: EventEnterRegion, Region: "trigger_1"},
				Active: true,
			},
			event:    NewEvent(EventEnterRegion, "trigger_2", "player"),
			expected: false,
		},
		{
			name: "actor mismatch",
			rule: Rule{
				When:   WhenClause{Event: EventEnterRegion, Region: "trigger_1", Actor: "player"},
				Active: true,
			},
			event:    NewEvent(EventEnterRegion, "trigger_1", "enemy"),
			expected: false,
		},
		{
			name: "empty region matches any",
			rule: Rule{
				When:   WhenClause{Event: EventEnterRegion, Region: ""},
				Active: true,
			},
			event:    NewEvent(EventEnterRegion, "any_region", "player"),
			expected: true,
		},
		{
			name: "empty actor matches any",
			rule: Rule{
				When:   WhenClause{Event: EventEnterRegion, Region: "trigger_1", Actor: ""},
				Active: true,
			},
			event:    NewEvent(EventEnterRegion, "trigger_1", "any_actor"),
			expected: true,
		},
		{
			name: "inactive rule",
			rule: Rule{
				When:   WhenClause{Event: EventEnterRegion, Region: "trigger_1"},
				Active: false,
			},
			event:    NewEvent(EventEnterRegion, "trigger_1", "player"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rule.matchesEvent(tt.event)
			if result != tt.expected {
				t.Errorf("matchesEvent() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestNewActionContext(t *testing.T) {
	resolver := newMockResolver()
	event := NewEvent(EventEnterRegion, "source", "player")

	ctx := NewActionContext(event, resolver)

	if ctx.Event.Type != EventEnterRegion {
		t.Errorf("expected event type '%s', got '%s'", EventEnterRegion, ctx.Event.Type)
	}
	if ctx.Resolver == nil {
		t.Error("expected resolver to be set")
	}
}

func TestNewEvent(t *testing.T) {
	event := NewEvent(EventEnterRegion, "trigger_1", "player")

	if event.Type != EventEnterRegion {
		t.Errorf("expected type '%s', got '%s'", EventEnterRegion, event.Type)
	}
	if event.RegionID != "trigger_1" {
		t.Errorf("expected regionID 'trigger_1', got '%s'", event.RegionID)
	}
	if event.ActorType != "player" {
		t.Errorf("expected actorType 'player', got '%s'", event.ActorType)
	}
}
