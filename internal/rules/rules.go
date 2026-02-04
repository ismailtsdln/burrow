package rules

// RiskLevel represents the safety level of a cleanup rule.
type RiskLevel string

const (
	RiskSafe    RiskLevel = "Safe"
	RiskCaution RiskLevel = "Caution"
	RiskManual  RiskLevel = "Manual"
)

// CleanupRule defines a single cleanup operation.
type CleanupRule struct {
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	Paths        []string  `json:"paths"`
	RiskLevel    RiskLevel `json:"risk_level"`
	Description  string    `json:"description"`
	Explanation  string    `json:"explanation"`
	RuleVersion  string    `json:"rule_version"`
	IntroducedIn string    `json:"introduced_in"`
}

// Result represents the outcome of a scan for a specific rule.
type Result struct {
	Rule       CleanupRule `json:"rule"`
	FoundPaths []string    `json:"found_paths"`
	TotalSize  int64       `json:"total_size"`
}
