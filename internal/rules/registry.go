package rules

// Registry manages the collection of cleanup rules.
type Registry struct {
	rules []CleanupRule
}

// NewRegistry initializes a new rules registry with default rules.
func NewRegistry() *Registry {
	r := &Registry{}
	r.registerDefaultRules()

	// Load custom rules
	if custom, err := LoadCustomRules(); err == nil && len(custom) > 0 {
		r.rules = append(r.rules, custom...)
	}

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
		{
			Name:         "Go Module Cache",
			Category:     "Package Managers",
			Paths:        []string{"~/go/pkg/mod"},
			RiskLevel:    RiskSafe,
			Description:  "Delete Go module cache.",
			Explanation:  "Determined by GOMODCACHE, this directory holds downloaded modules. Deleting it forces a redownload of dependencies on the next build, which is safe but consumes bandwidth.",
			RuleVersion:  "1.1.0",
			IntroducedIn: "0.2.0",
		},
		{
			Name:         "Yarn Cache",
			Category:     "Package Managers",
			Paths:        []string{"~/Library/Caches/Yarn"},
			RiskLevel:    RiskSafe,
			Description:  "Delete Yarn package cache.",
			Explanation:  "Yarn stores every downloaded package in a global cache. Deleting this frees up space but will make future yarn installs slower until the cache is repopulated.",
			RuleVersion:  "1.1.0",
			IntroducedIn: "0.2.0",
		},
		{
			Name:         "CocoaPods Cache",
			Category:     "Package Managers",
			Paths:        []string{"~/Library/Caches/CocoaPods"},
			RiskLevel:    RiskSafe,
			Description:  "Delete CocoaPods cache.",
			Explanation:  "CocoaPods caches pod specs and sources. Cleaning this directory is safe and useful for resolving pod installation issues.",
			RuleVersion:  "1.1.0",
			IntroducedIn: "0.2.0",
		},
		{
			Name:         "Composer Cache",
			Category:     "Package Managers",
			Paths:        []string{"~/.composer/cache"},
			RiskLevel:    RiskSafe,
			Description:  "Delete PHP Composer cache.",
			Explanation:  "Composer caches downloaded PHP packages. Safe to delete; packages will be re-downloaded as needed.",
			RuleVersion:  "1.1.0",
			IntroducedIn: "0.2.0",
		},
		{
			Name:         "Ruby Gem Cache",
			Category:     "Package Managers",
			Paths:        []string{"~/.gem/specs"},
			RiskLevel:    RiskSafe,
			Description:  "Delete Ruby Gem specs cache.",
			Explanation:  "Caches spec files for RubyGems. Safe to delete.",
			RuleVersion:  "1.1.0",
			IntroducedIn: "0.2.0",
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
			Name:         "Xcode Simulators",
			Category:     "Developer Tools",
			Paths:        []string{"~/Library/Developer/CoreSimulator/Devices"},
			RiskLevel:    RiskCaution,
			Description:  "Delete all Xcode Simulator devices.",
			Explanation:  "This deletes all simulated iOS/watchOS/tvOS devices. You will lose any apps installed on them and their data. Xcode will recreate fresh, empty simulators the next time you launch it or run a test.",
			RuleVersion:  "1.1.0",
			IntroducedIn: "0.2.0",
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
		{
			Name:         "Go Build Cache",
			Category:     "Developer Tools",
			Paths:        []string{"~/.cache/go-build"},
			RiskLevel:    RiskSafe,
			Description:  "Delete Go build cache.",
			Explanation:  "Go caches compiled packages to speed up builds. Deleting this is safe but will make the next build roughly as slow as a fresh build.",
			RuleVersion:  "1.1.0",
			IntroducedIn: "0.2.0",
		},
		{
			Name:         "JetBrains IDE Caches",
			Category:     "Developer Tools",
			Paths:        []string{"~/Library/Caches/JetBrains"},
			RiskLevel:    RiskCaution,
			Description:  "Delete IntelliJ/WebStorm/Goland caches.",
			Explanation:  "JetBrains IDEs store indexes and caches here. Deleting this will force the IDE to re-index all projects upon next launch, which can take significant time.",
			RuleVersion:  "1.1.0",
			IntroducedIn: "0.2.0",
		},
		{
			Name:     "VS Code Cache",
			Category: "Developer Tools",
			Paths: []string{
				"~/Library/Application Support/Code/Cache",
				"~/Library/Application Support/Code/Code Cache",
				"~/Library/Application Support/Code/CachedData",
				"~/Library/Application Support/Code/User/workspaceStorage",
			},
			RiskLevel:    RiskCaution,
			Description:  "Delete VS Code caches and workspace storage.",
			Explanation:  "Deletes generic caches and cached workspace data. Deleting workspaceStorage will not delete your code, but may reset local workspace state (UI layout, opened files history) for projects. Useful if VS Code is acting buggy.",
			RuleVersion:  "1.1.0",
			IntroducedIn: "0.2.0",
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
			Name:     "Electron App Caches",
			Category: "System",
			Paths: []string{
				"~/Library/Application Support/Slack/Cache",
				"~/Library/Application Support/Slack/Code Cache",
				"~/Library/Application Support/Discord/Cache",
				"~/Library/Application Support/Discord/Code Cache",
			},
			RiskLevel:    RiskSafe,
			Description:  "Delete caches for Slack and Discord.",
			Explanation:  "Electron apps (Slack, Discord) tend to accumulate large amounts of cache data over time. Deleting these is generally safe and forces the apps to fetch fresh data, often resolving UI glitches.",
			RuleVersion:  "1.1.0",
			IntroducedIn: "0.2.0",
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
			Paths:        []string{"~/.docker"},
			RiskLevel:    RiskManual,
			Description:  "Inspect Docker configuration and context.",
			Explanation:  "Burrow tracks the configuration size. To clean actual containers and images, run 'docker system prune'. Burrow does not directly delete Docker artifacts to prevent accidental data loss of persistent volumes.",
			RuleVersion:  "1.0.0",
			IntroducedIn: "0.1.0",
		},
	}
}
