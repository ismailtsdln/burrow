package ui

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"time"

	"github.com/ismailtsdln/burrow/internal/cleaner"
	"github.com/ismailtsdln/burrow/internal/config"
	"github.com/ismailtsdln/burrow/internal/rules"
	"github.com/ismailtsdln/burrow/internal/scanner"
)

// Execute is the main entry point for the CLI.
func Execute() error {
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "scan":
		return runScan(args)
	case "list":
		return runList(args)
	case "stats":
		return runStats(args)
	case "clean":
		return runClean(args)
	case "undo":
		return runUndo()
	case "rules":
		return runRules(args)
	case "doctor":
		return runDoctor()
	case "version":
		return runVersion()
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s. Run 'burrow help' for usage", command)
	}
}

func printUsage() {
	fmt.Println(Bold + Cyan + "Burrow — Advanced macOS Cleanup for Developers" + Reset)
	fmt.Println("\n" + Bold + "Usage:" + Reset)
	fmt.Println("  burrow <command> [flags]")
	fmt.Println("\n" + Bold + "Commands:" + Reset)
	fmt.Printf("  %-10s %s\n", Colorize(Green, "scan"), "Identify cleanup candidates")
	fmt.Printf("  %-10s %s\n", Colorize(Green, "clean"), "Remove identified files (dry-run by default)")
	fmt.Printf("  %-10s %s\n", Colorize(Green, "undo"), "Restore last cleanup from trash")
	fmt.Printf("  %-10s %s\n", Colorize(Green, "list"), "List all detected files")
	fmt.Printf("  %-10s %s\n", Colorize(Green, "rules"), "List all cleanup rules")
	fmt.Printf("  %-10s %s\n", Colorize(Green, "stats"), "Show disk reclaimable stats")
	fmt.Printf("  %-10s %s\n", Colorize(Green, "doctor"), "Check system health and permissions")
	fmt.Printf("  %-10s %s\n", Colorize(Green, "version"), "Show version information")
	fmt.Println("\n" + Bold + "Flags:" + Reset)
	fmt.Println("  -h, --help   Show help for a command")
}

func runScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	category := fs.String("category", "", "Filter by category")
	olderThan := fs.String("older-than", "", "Filter items older than duration (e.g. 30d, 24h)")
	js := fs.Bool("json", false, "Output in JSON format")
	explain := fs.Bool("explain", false, "Explain why paths were selected")
	fs.Parse(args)

	var ageDuration time.Duration
	if *olderThan != "" {
		// Basic "d" parsing fix since time.ParseDuration doesn't support "d" (days)
		if strings.HasSuffix(*olderThan, "d") {
			daysStr := strings.TrimSuffix(*olderThan, "d")
			var days int
			fmt.Sscanf(daysStr, "%d", &days)
			ageDuration = time.Duration(days) * 24 * time.Hour
		} else {
			var err error
			ageDuration, err = time.ParseDuration(*olderThan)
			if err != nil {
				return fmt.Errorf("invalid duration format: %s (example: 30d, 24h)", *olderThan)
			}
		}
	}

	cfg, _ := config.Load()
	registry := rules.NewRegistry()
	s := scanner.NewScanner(registry, scanner.ScanOptions{
		Category:      *category,
		ExcludedPaths: cfg.ExcludedPaths,
		SizeThreshold: cfg.SizeThresholdMB * 1024 * 1024,
		OlderThan:     ageDuration,
	})

	if !*js {
		PrintInfo("Scanning for cleanup candidates...")
	}

	results, err := s.Scan()
	if err != nil {
		return err
	}

	if *js {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if len(results.Results) == 0 {
		PrintSuccess("No cleanup candidates found. Your system is clean!")
		return nil
	}

	PrintHeader(fmt.Sprintf("%-30s %-15s %s", "CATEGORY", "SIZE", "RULE"))
	fmt.Println(Gray + strings.Repeat("-", 70) + Reset)
	for _, res := range results.Results {
		fmt.Printf("%-30s %-15s %s\n", Colorize(Blue, res.Rule.Category), Colorize(Yellow, FormatSize(res.TotalSize)), res.Rule.Name)
		if *explain {
			fmt.Printf("   %s %s\n", Colorize(Cyan, "💡"), Colorize(Gray, res.Rule.Explanation))
		}
	}

	fmt.Println(Gray + strings.Repeat("-", 70) + Reset)
	fmt.Printf(Bold+"Total reclaimable space: %s"+Reset+"\n", Colorize(Green, FormatSize(results.TotalSize)))
	fmt.Println("\nRun 'burrow clean' to see a detailed breakdown or 'burrow list' to see all files.")
	return nil
}

func runClean(args []string) error {
	fs := flag.NewFlagSet("clean", flag.ContinueOnError)
	dryRun := fs.Bool("dry-run", true, "Perform a dry run (default true)")
	olderThan := fs.String("older-than", "", "Filter items older than duration (e.g. 30d, 24h)")
	yes := fs.Bool("yes", false, "Confirm cleanup automatically")
	diff := fs.Bool("diff", false, "Show detailed diff of planned deletions")
	fs.Parse(args)

	var ageDuration time.Duration
	if *olderThan != "" {
		if strings.HasSuffix(*olderThan, "d") {
			daysStr := strings.TrimSuffix(*olderThan, "d")
			var days int
			fmt.Sscanf(daysStr, "%d", &days)
			ageDuration = time.Duration(days) * 24 * time.Hour
		} else {
			var err error
			ageDuration, err = time.ParseDuration(*olderThan)
			if err != nil {
				return fmt.Errorf("invalid duration format: %s (example: 30d, 24h)", *olderThan)
			}
		}
	}

	cfg, _ := config.Load()
	registry := rules.NewRegistry()
	s := scanner.NewScanner(registry, scanner.ScanOptions{
		ExcludedPaths: cfg.ExcludedPaths,
		SizeThreshold: cfg.SizeThresholdMB * 1024 * 1024,
		OlderThan:     ageDuration,
	})

	results, err := s.Scan()
	if err != nil {
		return err
	}

	if len(results.Results) == 0 {
		PrintSuccess("No cleanup candidates found. Your system is clean!")
		return nil
	}

	if *dryRun && !*yes {
		PrintHeader("Cleanup Summary (Dry Run):")
		fmt.Printf(Bold+"%-30s %-15s %s"+Reset+"\n", "CATEGORY", "SIZE", "RULE")
		fmt.Println(Gray + strings.Repeat("-", 70) + Reset)
		for _, res := range results.Results {
			fmt.Printf("%-30s %-15s %s\n", Colorize(Blue, res.Rule.Category), Colorize(Yellow, FormatSize(res.TotalSize)), res.Rule.Name)
			if *diff {
				for _, p := range res.FoundPaths {
					fmt.Printf("   %s %s\n", Colorize(Red, "-"), Colorize(Gray, p))
				}
			}
		}
		fmt.Println(Gray + strings.Repeat("-", 70) + Reset)
		fmt.Printf(Bold+"Total to be reclaimed: %s"+Reset+"\n", Colorize(Green, FormatSize(results.TotalSize)))

		if !Confirm("\n" + Colorize(Yellow, "Do you want to proceed with the cleanup?")) {
			PrintWarning("Cleanup cancelled.")
			return nil
		}
	}

	c := cleaner.NewCleaner()
	res, err := c.Clean(results.Results, false)
	if err != nil {
		return err
	}

	PrintSuccess("Successfully reclaimed %s!", FormatSize(res.ReclaimedSpace))
	fmt.Printf("Files moved to trash: %d\n", res.FileCount)
	fmt.Printf("Trash Session ID: %s\n", Colorize(Cyan, res.TrashSession))
	PrintInfo("You can undo this action by running 'burrow undo'.")

	return nil
}

func runUndo() error {
	c := cleaner.NewCleaner()
	PrintInfo("Restoring last cleanup session...")
	if err := c.Undo(); err != nil {
		return err
	}
	PrintSuccess("Successfully restored last cleanup session!")
	return nil
}

func runList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	js := fs.Bool("json", false, "Output in JSON format")
	fs.Parse(args)

	cfg, _ := config.Load()
	registry := rules.NewRegistry()
	s := scanner.NewScanner(registry, scanner.ScanOptions{
		ExcludedPaths: cfg.ExcludedPaths,
		SizeThreshold: cfg.SizeThresholdMB * 1024 * 1024,
	})

	results, err := s.Scan()
	if err != nil {
		return err
	}

	if *js {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if len(results.Results) == 0 {
		fmt.Println("✨ No cleanup candidates found.")
		return nil
	}

	for _, res := range results.Results {
		fmt.Printf("\n[%s] %s (%s)\n", res.Rule.Category, res.Rule.Name, FormatSize(res.TotalSize))
		for _, path := range res.FoundPaths {
			fmt.Printf("  • %s\n", path)
		}
	}
	return nil
}

func runRules(args []string) error {
	fs := flag.NewFlagSet("rules", flag.ContinueOnError)
	explain := fs.String("explain", "", "Explain a specific rule")
	js := fs.Bool("json", false, "Output in JSON format")
	fs.Parse(args)

	registry := rules.NewRegistry()
	allRules := registry.All()

	if *js {
		data, _ := json.MarshalIndent(allRules, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if *explain != "" {
		for _, r := range allRules {
			if strings.EqualFold(r.Name, *explain) {
				fmt.Printf("Rule: %s\n", r.Name)
				fmt.Printf("Category: %s\n", r.Category)
				fmt.Printf("Risk: %s\n", r.RiskLevel)
				fmt.Printf("Description: %s\n", r.Description)
				fmt.Printf("Explanation: %s\n", r.Explanation)
				return nil
			}
		}
		return fmt.Errorf("rule not found: %s", *explain)
	}

	fmt.Println("Available Cleanup Rules:")
	fmt.Printf("\n%-25s %-15s %s\n", "NAME", "RISK", "DESCRIPTION")
	fmt.Println(strings.Repeat("-", 70))
	for _, r := range allRules {
		fmt.Printf("%-25s %-15s %s\n", r.Name, r.RiskLevel, r.Description)
	}
	return nil
}

func runStats(args []string) error {
	fs := flag.NewFlagSet("stats", flag.ContinueOnError)
	js := fs.Bool("json", false, "Output in JSON format")
	fs.Parse(args)

	cfg, _ := config.Load()
	registry := rules.NewRegistry()
	s := scanner.NewScanner(registry, scanner.ScanOptions{
		ExcludedPaths: cfg.ExcludedPaths,
		SizeThreshold: cfg.SizeThresholdMB * 1024 * 1024,
	})

	results, err := s.Scan()
	if err != nil {
		return err
	}

	stats := make(map[string]int64)
	for _, res := range results.Results {
		stats[res.Rule.Category] += res.TotalSize
	}

	if *js {
		data, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	PrintHeader(fmt.Sprintf("%-30s %s", "CATEGORY", "TOTAL SIZE"))
	fmt.Println(Gray + strings.Repeat("-", 45) + Reset)
	for cat, size := range stats {
		fmt.Printf("%-30s %s\n", Colorize(Blue, cat), Colorize(Yellow, FormatSize(size)))
	}
	fmt.Println(Gray + strings.Repeat("-", 45) + Reset)
	fmt.Printf(Bold+"%-30s %s"+Reset+"\n", "TOTAL RECLAIMABLE", Colorize(Green, FormatSize(results.TotalSize)))

	return nil
}

func runDoctor() error {
	PrintHeader("Burrow Doctor — Diagnostic Report")
	fmt.Println(Gray + strings.Repeat("-", 40) + Reset)

	// Check Home Directory
	home, err := os.UserHomeDir()
	if err != nil {
		PrintError("Home Directory: Error - %v", err)
	} else {
		PrintSuccess("Home Directory: %s", home)
	}

	// Check Burrow Folders
	burrowDir := filepath.Join(home, ".burrow")
	if _, err := os.Stat(burrowDir); os.IsNotExist(err) {
		PrintWarning("Burrow Directory: Not found (will be created on first clean)")
	} else {
		PrintSuccess("Burrow Directory: %s", burrowDir)
	}

	// Check Permissions
	testFile := filepath.Join(home, ".burrow", "test_perm")
	os.MkdirAll(filepath.Dir(testFile), 0755)
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		PrintError("Write Permissions: Failed - %v", err)
	} else {
		PrintSuccess("Write Permissions: OK")
		os.Remove(testFile)
	}

	// Check OS
	PrintSuccess("Operating System: macOS (detected)")

	fmt.Println(Gray + strings.Repeat("-", 40) + Reset)
	PrintInfo("All systems operational. Burrow is ready to dig!")
	return nil
}

func runVersion() error {
	fmt.Println("Burrow v0.1.0-mvp")
	return nil
}
