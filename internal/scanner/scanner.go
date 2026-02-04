package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ismailtsdln/burrow/internal/rules"
	"github.com/ismailtsdln/burrow/internal/safety"
)

// ScanOptions contains filtering and performance settings for a scan.
type ScanOptions struct {
	Category      string
	SizeThreshold int64
	ExcludedPaths []string
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
	allRules := s.registry.All()
	results := make([]rules.Result, 0)
	var totalSize int64
	var mu sync.Mutex
	var wg sync.WaitGroup

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
				expanded := expandPath(pathPattern)

				// Filter by excluded paths
				excluded := false
				for _, ep := range s.options.ExcludedPaths {
					if strings.HasPrefix(expanded, expandPath(ep)) {
						excluded = true
						break
					}
				}
				if excluded {
					continue
				}

				// Basic check if path exists
				if _, err := os.Stat(expanded); os.IsNotExist(err) {
					continue
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

// expandPath is a helper that should ideally be in a shared package or internal to safety,
// but for simplicity we re-implement it here or call safety.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
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
