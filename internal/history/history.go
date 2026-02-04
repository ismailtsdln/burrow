package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Entry represents a single cleanup session record.
type Entry struct {
	ID             string           `json:"id"`
	Timestamp      time.Time        `json:"timestamp"`
	ReclaimedBytes int64            `json:"reclaimed_bytes"`
	FileCount      int              `json:"file_count"`
	CategoryStats  map[string]int64 `json:"category_stats"`
}

// Manager handles history operations.
type Manager struct {
	historyPath string
}

// NewManager creates a new history manager.
func NewManager() *Manager {
	home, _ := os.UserHomeDir()
	return &Manager{
		historyPath: filepath.Join(home, ".burrow", "history.json"),
	}
}

// Save appends a new entry to the history.
func (m *Manager) Save(entry Entry) error {
	entries, _ := m.Load()
	entries = append(entries, entry)

	// Keep only last 50 entries to avoid bloating
	if len(entries) > 50 {
		entries = entries[len(entries)-50:]
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.historyPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(m.historyPath, data, 0644)
}

// Load returns all history entries sorted by timestamp (newest first).
func (m *Manager) Load() ([]Entry, error) {
	if _, err := os.Stat(m.historyPath); os.IsNotExist(err) {
		return []Entry{}, nil
	}

	data, err := os.ReadFile(m.historyPath)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	// Sort newest first
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	return entries, nil
}
