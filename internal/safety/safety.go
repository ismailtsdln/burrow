package safety

import (
	"os"
	"path/filepath"
	"strings"
)

// IsSafe returns true if the path is safe to delete.
func IsSafe(path string) (bool, string) {
	absPath, err := filepath.Abs(ExpandPath(path))
	if err != nil {
		return false, "Invalid path"
	}

	// 1. Guard against home directory and root
	home, _ := os.UserHomeDir()
	if absPath == home || absPath == "/" {
		return false, "Cannot delete home or root directory"
	}

	// 2. Guard against System Integrity Protection (SIP) and system paths
	systemPaths := []string{
		"/System",
		"/Library/Apple",
		"/usr/bin",
		"/usr/sbin",
		"/bin",
		"/sbin",
	}
	for _, p := range systemPaths {
		if strings.HasPrefix(absPath, p) {
			return false, "Path is protected by System Integrity Protection (SIP)"
		}
	}

	// 3. Guard against Git repositories
	if isGitRepo(absPath) {
		return false, "Path contains Git metadata (.git)"
	}

	// 4. Guard against common user directories
	userDocPaths := []string{
		filepath.Join(home, "Documents"),
		filepath.Join(home, "Desktop"),
		filepath.Join(home, "Downloads"),
	}
	for _, p := range userDocPaths {
		if absPath == p {
			return false, "Path is a common user document directory"
		}
	}

	return true, ""
}

// ExpandPath replaces ~ with the user's home directory.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// isGitRepo checks if the path or any of its parents/children contain .git.
func isGitRepo(path string) bool {
	// Check if the path itself is .git
	if filepath.Base(path) == ".git" {
		return true
	}

	// Check for .git in the current directory or any subdirectory
	found := false
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() && info.Name() == ".git" {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	if found {
		return true
	}

	// Traverse up to check if we are inside a git repo
	curr := path
	for {
		if _, err := os.Stat(filepath.Join(curr, ".git")); err == nil {
			return true
		}
		parent := filepath.Dir(curr)
		if parent == curr || parent == "/" || parent == "." {
			break
		}
		curr = parent
	}

	return false
}
