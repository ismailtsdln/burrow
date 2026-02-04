package rules

// Registry manages the collection of cleanup rules.
type Registry struct {
	rules []CleanupRule
}

// NewRegistry initializes a new rules registry with default rules.
func NewRegistry() *Registry {
	r := &Registry{}
	r.registerDefaultRules()
	return r
}

// All returns all registered cleanup rules.
func (r *Registry) All() []CleanupRule {
	return r.rules
}

// registerDefaultRules populates the registry with built-in rules.
func (r *Registry) registerDefaultRules() {
	r.rules = []CleanupRule{
		// Package Managers
		{
			Name:         "Homebrew Cache",
			Category:     "Package Managers",
			Paths:        []string{"~/Library/Caches/Homebrew"},
			RiskLevel:    RiskSafe,
			Description:  "Delete downloaded Homebrew formulae and bottles.",
			Explanation:  "Homebrew caches downloaded source code and pre-compiled binaries (bottles). Deleting this will reclaim space without affecting installed software. Homebrew will simply re-download what it needs during the next update or install.",
			RuleVersion:  "1.0.0",
			IntroducedIn: "0.1.0",
		},
		{
			Name:         "npm Cache",
			Category:     "Package Managers",
			Paths:        []string{"~/.npm/_cacache", "~/.npm/_logs"},
			RiskLevel:    RiskSafe,
			Description:  "Delete npm global cache and logs.",
			Explanation:  "The npm cache stores package data to avoid redundant network requests. Deleting it is safe because npm handles missing cache entries by fetching them from the registry. It also removes debug logs which are only useful for troubleshooting failed installs.",
			RuleVersion:  "1.0.0",
			IntroducedIn: "0.1.0",
		},
		{
			Name:         "pip Cache",
			Category:     "Package Managers",
			Paths:        []string{"~/Library/Caches/pip"},
			RiskLevel:    RiskSafe,
			Description:  "Delete pip package cache.",
			Explanation:  "Python's pip tool caches wheels and source distributions to speed up re-installation of the same versions. Deleting this cache is safe; pip will re-download the packages from PyPI as needed.",
			RuleVersion:  "1.0.0",
			IntroducedIn: "0.1.0",
		},
		{
			Name:         "Cargo Registry Cache",
			Category:     "Package Managers",
			Paths:        []string{"~/.cargo/registry/cache", "~/.cargo/registry/index"},
			RiskLevel:    RiskSafe,
			Description:  "Delete Rust cargo registry cache.",
			Explanation:  "Cargo caches registry index metadata and downloaded crate source files. While deleting this saves space, the next 'cargo build' will involve a 'Updating crates.io index' step followed by re-downloading dependencies. It has zero impact on compiled binaries.",
			RuleVersion:  "1.0.0",
			IntroducedIn: "0.1.0",
		},

		// Developer Tools
		{
			Name:         "Xcode DerivedData",
			Category:     "Developer Tools",
			Paths:        []string{"~/Library/Developer/Xcode/DerivedData"},
			RiskLevel:    RiskSafe,
			Description:  "Delete Xcode build artifacts and indexes.",
			Explanation:  "DerivedData contains intermediate build products, debug symbols, and module caches. It is the most common source of 'ghost bugs' in Xcode. Deleting it is safe and often recommended; Xcode will rebuild everything from scratch and re-index your projects.",
			RuleVersion:  "1.0.0",
			IntroducedIn: "0.1.0",
		},
		{
			Name:         "Android Build Cache",
			Category:     "Developer Tools",
			Paths:        []string{"~/.android/build-cache"},
			RiskLevel:    RiskSafe,
			Description:  "Delete Android SDK build cache.",
			Explanation:  "The Android build cache stores pre-dexed libraries and other build artifacts. Deleting it is safe; the Android Gradle plugin will re-populate it during subsequent builds.",
			RuleVersion:  "1.0.0",
			IntroducedIn: "0.1.0",
		},
		{
			Name:         "Gradle Cache",
			Category:     "Developer Tools",
			Paths:        []string{"~/.gradle/caches"},
			RiskLevel:    RiskCaution,
			Description:  "Delete Gradle dependency caches.",
			Explanation:  "This directory contains all JARs and artifacts downloaded by Gradle. While safe from a data integrity perspective, deleting it will force every project to re-download all dependencies, which can be extremely slow and consume significant bandwidth.",
			RuleVersion:  "1.0.0",
			IntroducedIn: "0.1.0",
		},

		// System
		{
			Name:         "User Caches",
			Category:     "System",
			Paths:        []string{"~/Library/Caches"},
			RiskLevel:    RiskCaution,
			Description:  "Delete general application caches.",
			Explanation:  "General macOS application caches. While most apps handle missing caches gracefully, some may experience temporary performance degradation or lose local-only state (like unsynced drafts or transient UI preferences). Use with caution.",
			RuleVersion:  "1.0.0",
			IntroducedIn: "0.1.0",
		},
		{
			Name:         "Temporary Files",
			Category:     "System",
			Paths:        []string{"/tmp", "/var/tmp"},
			RiskLevel:    RiskSafe,
			Description:  "Delete system temporary files.",
			Explanation:  "Safe to delete. These files are typically temporary and short-lived.",
			RuleVersion:  "1.0.0",
			IntroducedIn: "0.1.0",
		},

		// Containers (INSPECTION ONLY for MVP)
		{
			Name:         "Docker System Usage",
			Category:     "Containers",
			Paths:        []string{"/var/lib/docker"}, // Note: This is a placeholder path for info
			RiskLevel:    RiskManual,
			Description:  "Inspect Docker disk usage.",
			Explanation:  "MVP: Inspection only. No deletion is performed on container data. Use 'docker system df' for details.",
			RuleVersion:  "1.0.0",
			IntroducedIn: "0.1.0",
		},
	}
}
