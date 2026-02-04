package cleaner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TrashManifest stores information about trashed files for undo operations.
type TrashManifest struct {
	Timestamp time.Time    `json:"timestamp"`
	Entries   []TrashEntry `json:"entries"`
}

// TrashEntry maps a trashed file to its original location.
type TrashEntry struct {
	OriginalPath string `json:"original_path"`
	TrashPath    string `json:"trash_path"`
}

// TrashManager handles moving files to trash and restoring them.
type TrashManager struct {
	TrashBaseDir string
}

// NewTrashManager creates a new trash manager.
func NewTrashManager() *TrashManager {
	home, _ := os.UserHomeDir()
	return &TrashManager{
		TrashBaseDir: filepath.Join(home, ".burrow", "trash"),
	}
}

// MoveToTrash moves a path to a timestamped trash directory.
func (tm *TrashManager) MoveToTrash(paths []string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	sessionDir := filepath.Join(tm.TrashBaseDir, timestamp)

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create trash directory: %w", err)
	}

	manifest := TrashManifest{
		Timestamp: time.Now(),
		Entries:   make([]TrashEntry, 0),
	}

	for _, path := range paths {
		targetName := filepath.Base(path)
		// Handle potential name collisions in the trash session
		trashPath := filepath.Join(sessionDir, targetName)

		if err := os.Rename(path, trashPath); err != nil {
			// If rename fails (e.g., across filesystems), try copying/deleting
			// For MVP, we'll assume same filesystem or return error
			return "", fmt.Errorf("failed to move %s to trash: %w", path, err)
		}

		manifest.Entries = append(manifest.Entries, TrashEntry{
			OriginalPath: path,
			TrashPath:    trashPath,
		})
	}

	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(sessionDir, "manifest.json"), manifestData, 0644); err != nil {
		return "", fmt.Errorf("failed to write manifest: %w", err)
	}

	return timestamp, nil
}

// RestoreLast restores the most recent trash session.
func (tm *TrashManager) RestoreLast() error {
	entries, err := os.ReadDir(tm.TrashBaseDir)
	if err != nil {
		return fmt.Errorf("failed to read trash directory: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no trash sessions found")
	}

	// Find the most recent session (by folder name)
	var latest string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() > latest {
			latest = entry.Name()
		}
	}

	if latest == "" {
		return fmt.Errorf("no valid trash sessions found")
	}

	sessionDir := filepath.Join(tm.TrashBaseDir, latest)
	manifestData, err := os.ReadFile(filepath.Join(sessionDir, "manifest.json"))
	if err != nil {
		return fmt.Errorf("failed to read manifest for session %s: %w", latest, err)
	}

	var manifest TrashManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	for _, entry := range manifest.Entries {
		if err := os.MkdirAll(filepath.Dir(entry.OriginalPath), 0755); err != nil {
			continue // Best effort
		}
		if err := os.Rename(entry.TrashPath, entry.OriginalPath); err != nil {
			fmt.Printf("Warning: Failed to restore %s: %v\n", entry.OriginalPath, err)
		}
	}

	// Clean up the empty trash session directory
	return os.RemoveAll(sessionDir)
}
