package rules

import (
	"encoding/json"
	"os"

	"github.com/ismailtsdln/burrow/internal/safety"
)

// LoadCustomRules loads rules from ~/.config/burrow/custom_rules.json
func LoadCustomRules() ([]CleanupRule, error) {
	path := safety.ExpandPath("~/.config/burrow/custom_rules.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil // No custom rules file, which is fine
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var customRules []CleanupRule
	if err := json.Unmarshal(data, &customRules); err != nil {
		return nil, err
	}

	// Set defaults for custom rules
	for i := range customRules {
		if customRules[i].Category == "" {
			customRules[i].Category = "Custom"
		}
		if customRules[i].RiskLevel == "" {
			customRules[i].RiskLevel = RiskManual // Default to manual/caution for safety
		}
		customRules[i].IntroducedIn = "custom"
		customRules[i].RuleVersion = "1.0.0"
	}

	return customRules, nil
}
