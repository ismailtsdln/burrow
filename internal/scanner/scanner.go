package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ismailtsdln/burrow/internal/rules"
	"github.com/ismailtsdln/burrow/internal/safety"
)

// ScanOptions contains filtering and performance settings for a scan.
type ScanOptions struct {
	Category      string
	SizeThreshold int64
	ExcludedPaths []string
	OlderThan     time.Duration
	LargeFileMode bool
}

// Scanner handles the scanning of the filesystem for cleanup candidates.
type Scanner struct {
	registry *rules.Registry
	options  ScanOptions
}

// NewScanner creates a new scanner instance with options.
func NewScanner(registry *rules.Registry, options ScanOptions) *Scanner {
	return &Scanner{
		registry: registry,
		options:  options,
	}
}

// ScanResults contains the results of a scan.
type ScanResults struct {
	Results   []rules.Result
	TotalSize int64
}

// Scan performs a scan based on the registered rules.
func (s *Scanner) Scan() (*ScanResults, error) {
	results := make([]rules.Result, 0)
	var totalSize int64
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Large File Scan Mode
	if s.options.LargeFileMode {
		dirsToScan := []string{
			"~/Downloads",
			"~/Desktop",
			"~/Documents",
			"~/Movies",
			"~/Pictures",
		}

		threshold := s.options.SizeThreshold
		if threshold == 0 {
			threshold = 100 * 1024 * 1024 // Default 100MB
		}

		for _, dir := range dirsToScan {
			expanded := safety.ExpandPath(dir)
			wg.Add(1)
			go func(d string) {
				defer wg.Done()
				var foundPaths []string
				var ruleSize int64

				filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
					if err != nil || info.IsDir() {
						return nil
					}
					if info.Size() > threshold {
						foundPaths = append(foundPaths, path)
						ruleSize += info.Size()
					}
					return nil
				})

				if len(foundPaths) > 0 {
					mu.Lock()
					results = append(results, rules.Result{
						Rule: rules.CleanupRule{
							Name:        "Large Files (>100MB)",
							Category:    "Large Files",
							Description: fmt.Sprintf("Files larger than %s in %s", formatBytes(threshold), d),
							RiskLevel:   rules.RiskManual,
						},
						FoundPaths: foundPaths,
						TotalSize:  ruleSize,
					})
					totalSize += ruleSize
					mu.Unlock()
				}
			}(expanded)
		}
		wg.Wait()
		return &ScanResults{Results: results, TotalSize: totalSize}, nil
	}

	// Regular Rule-Based Scan
	results = make([]rules.Result, 0)
	totalSize = 0

	allRules := s.registry.All()
	// Reuse existing variables, reset results for standard scan if not in large mode

	for _, rule := range allRules {
		// Filter by category if specified
		if s.options.Category != "" && !strings.EqualFold(rule.Category, s.options.Category) {
			continue
		}

		wg.Add(1)
		go func(r rules.CleanupRule) {
			defer wg.Done()

			var foundPaths []string
			var ruleSize int64

			for _, pathPattern := range r.Paths {
				expanded := safety.ExpandPath(pathPattern)

				// Filter by excluded paths
				excluded := false
				for _, ep := range s.options.ExcludedPaths {
					if strings.HasPrefix(expanded, safety.ExpandPath(ep)) {
						excluded = true
						break
					}
				}
				if excluded {
					continue
				}

				info, err := os.Stat(expanded)
				if err != nil {
					continue
				}

				// Basic check if path exists (redundant but safe)
				if os.IsNotExist(err) {
					continue
				}

				// Filter by Time (OlderThan)
				if s.options.OlderThan > 0 {
					if time.Since(info.ModTime()) < s.options.OlderThan {
						continue
					}
				}

				// Safety check
				if safe, _ := safety.IsSafe(expanded); !safe {
					continue
				}

				size, err := dirSize(expanded)
				if err != nil {
					continue
				}

				// Filter by size threshold
				if s.options.SizeThreshold > 0 && size < s.options.SizeThreshold {
					continue
				}

				foundPaths = append(foundPaths, expanded)
				ruleSize += size
			}

			if len(foundPaths) > 0 {
				mu.Lock()
				results = append(results, rules.Result{
					Rule:       r,
					FoundPaths: foundPaths,
					TotalSize:  ruleSize,
				})
				totalSize += ruleSize
				mu.Unlock()
			}
		}(rule)
	}

	wg.Wait()

	return &ScanResults{
		Results:   results,
		TotalSize: totalSize,
	}, nil
}

// dirSize calculates the total size of a directory.
func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
