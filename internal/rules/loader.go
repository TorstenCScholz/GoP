package rules

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// ParseYAML parses rules from YAML data.
func ParseYAML(data []byte) (RuleSet, error) {
	var ruleSet RuleSet
	if err := yaml.Unmarshal(data, &ruleSet); err != nil {
		return RuleSet{}, fmt.Errorf("failed to parse YAML rules: %w", err)
	}
	return ruleSet, nil
}

// ParseJSON parses rules from JSON data.
func ParseJSON(data []byte) (RuleSet, error) {
	var ruleSet RuleSet
	if err := json.Unmarshal(data, &ruleSet); err != nil {
		return RuleSet{}, fmt.Errorf("failed to parse JSON rules: %w", err)
	}
	return ruleSet, nil
}

// LoadYAML loads rules from YAML data into the engine.
func (e *Engine) LoadYAML(data []byte) error {
	ruleSet, err := ParseYAML(data)
	if err != nil {
		return err
	}
	e.LoadRuleSet(ruleSet)
	return nil
}

// LoadJSON loads rules from JSON data into the engine.
func (e *Engine) LoadJSON(data []byte) error {
	ruleSet, err := ParseJSON(data)
	if err != nil {
		return err
	}
	e.LoadRuleSet(ruleSet)
	return nil
}
