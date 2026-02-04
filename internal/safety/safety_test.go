package safety

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsSafe(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		path     string
		wantSafe bool
	}{
		{"Home directory", home, false},
		{"Root directory", "/", false},
		{"System directory", "/System", false},
		{"User Documents", filepath.Join(home, "Documents"), false},
		{"User Desktop", filepath.Join(home, "Desktop"), false},
		{"Homebrew Cache", filepath.Join(home, "Library/Caches/Homebrew"), true},
		{"npm Cache", filepath.Join(home, ".npm/_cacache"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSafe, msg := IsSafe(tt.path)
			if gotSafe != tt.wantSafe {
				t.Errorf("IsSafe(%v) = %v (msg: %v), want %v", tt.path, gotSafe, msg, tt.wantSafe)
			}
		})
	}
}

func TestIsGitRepo(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "burrow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	otherDir, _ := os.MkdirTemp("", "burrow-safe-*")
	defer os.RemoveAll(otherDir)

	tests := []struct {
		name   string
		path   string
		isRepo bool
	}{
		{"Git root", tempDir, true},
		{"Inside git", subDir, true},
		{"Not a git repo", otherDir, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isGitRepo(tt.path); got != tt.isRepo {
				t.Errorf("isGitRepo(%v) = %v, want %v", tt.path, got, tt.isRepo)
			}
		})
	}
}
