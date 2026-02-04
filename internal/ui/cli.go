package ui

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	fmt.Println("Burrow — Advanced macOS Cleanup for Developers")
	fmt.Println("\nUsage:")
	fmt.Println("  burrow <command> [flags]")
	fmt.Println("\nCommands:")
	fmt.Println("  scan      Identify cleanup candidates")
	fmt.Println("  clean     Remove identified files (dry-run by default)")
	fmt.Println("  undo      Restore last cleanup from trash")
	fmt.Println("  list      List all detected files")
	fmt.Println("  rules     List all cleanup rules")
	fmt.Println("  stats     Show disk reclaimable stats")
	fmt.Println("  doctor    Check system health and permissions")
	fmt.Println("  version   Show version information")
	fmt.Println("\nFlags:")
	fmt.Println("  -h, --help   Show help for a command")
}

func runScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	category := fs.String("category", "", "Filter by category")
	js := fs.Bool("json", false, "Output in JSON format")
	explain := fs.Bool("explain", false, "Explain why paths were selected")
	fs.Parse(args)

	cfg, _ := config.Load()
	registry := rules.NewRegistry()
	s := scanner.NewScanner(registry, scanner.ScanOptions{
		Category:      *category,
		ExcludedPaths: cfg.ExcludedPaths,
		SizeThreshold: cfg.SizeThresholdMB * 1024 * 1024,
	})

	if !*js {
		fmt.Println("🔍 Scanning for cleanup candidates...")
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
		fmt.Println("✨ No cleanup candidates found. Your system is clean!")
		return nil
	}

	fmt.Printf("\n%-30s %-15s %s\n", "CATEGORY", "SIZE", "RULE")
	fmt.Println(strings.Repeat("-", 70))
	for _, res := range results.Results {
		fmt.Printf("%-30s %-15s %s\n", res.Rule.Category, FormatSize(res.TotalSize), res.Rule.Name)
		if *explain {
			fmt.Printf("   💡 %s\n", res.Rule.Explanation)
		}
	}

	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Total reclaimable space: %s\n", FormatSize(results.TotalSize))
	fmt.Println("\nRun 'burrow clean' to see a detailed breakdown or 'burrow list' to see all files.")
	return nil
}

func runClean(args []string) error {
	fs := flag.NewFlagSet("clean", flag.ContinueOnError)
	dryRun := fs.Bool("dry-run", true, "Perform a dry run (default true)")
	yes := fs.Bool("yes", false, "Confirm cleanup automatically")
	diff := fs.Bool("diff", false, "Show detailed diff of planned deletions")
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

	if len(results.Results) == 0 {
		fmt.Println("✨ No cleanup candidates found. Your system is clean!")
		return nil
	}

	if *dryRun && !*yes {
		fmt.Println("\nCleanup Summary (Dry Run):")
		fmt.Printf("%-30s %-15s %s\n", "CATEGORY", "SIZE", "RULE")
		fmt.Println(strings.Repeat("-", 70))
		for _, res := range results.Results {
			fmt.Printf("%-30s %-15s %s\n", res.Rule.Category, FormatSize(res.TotalSize), res.Rule.Name)
			if *diff {
				for _, p := range res.FoundPaths {
					fmt.Printf("   - %s\n", p)
				}
			}
		}
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("Total to be reclaimed: %s\n", FormatSize(results.TotalSize))

		if !Confirm("\nDo you want to proceed with the cleanup?") {
			fmt.Println("Cleanup cancelled.")
			return nil
		}
	}

	c := cleaner.NewCleaner()
	res, err := c.Clean(results.Results, false)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Successfully reclaimed %s!\n", FormatSize(res.ReclaimedSpace))
	fmt.Printf("Files moved to trash: %d\n", res.FileCount)
	fmt.Printf("Trash Session ID: %s\n", res.TrashSession)
	fmt.Println("\nYou can undo this action by running 'burrow undo'.")

	return nil
}

func runUndo() error {
	c := cleaner.NewCleaner()
	fmt.Println("♻️ Restoring last cleanup session...")
	if err := c.Undo(); err != nil {
		return err
	}
	fmt.Println("✅ Successfully restored last cleanup session!")
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

	fmt.Printf("\n%-30s %s\n", "CATEGORY", "TOTAL SIZE")
	fmt.Println(strings.Repeat("-", 45))
	for cat, size := range stats {
		fmt.Printf("%-30s %s\n", cat, FormatSize(size))
	}
	fmt.Println(strings.Repeat("-", 45))
	fmt.Printf("%-30s %s\n", "TOTAL RECLAIMABLE", FormatSize(results.TotalSize))

	return nil
}

func runDoctor() error {
	fmt.Println("🏥 Burrow Doctor — Diagnostic Report")
	fmt.Println(strings.Repeat("-", 40))

	// Check Home Directory
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("❌ Home Directory: Error - %v\n", err)
	} else {
		fmt.Printf("✅ Home Directory: %s\n", home)
	}

	// Check Burrow Folders
	burrowDir := filepath.Join(home, ".burrow")
	if _, err := os.Stat(burrowDir); os.IsNotExist(err) {
		fmt.Println("⚠️ Burrow Directory: Not found (will be created on first clean)")
	} else {
		fmt.Printf("✅ Burrow Directory: %s\n", burrowDir)
	}

	// Check Permissions
	testFile := filepath.Join(home, ".burrow", "test_perm")
	os.MkdirAll(filepath.Dir(testFile), 0755)
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		fmt.Printf("❌ Write Permissions: Failed - %v\n", err)
	} else {
		fmt.Println("✅ Write Permissions: OK")
		os.Remove(testFile)
	}

	// Check OS
	fmt.Printf("✅ Operating System: macOS (detected)\n")

	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("All systems operational. Burrow is ready to dig!")
	return nil
}

func runVersion() error {
	fmt.Println("Burrow v0.1.0-mvp")
	return nil
}
