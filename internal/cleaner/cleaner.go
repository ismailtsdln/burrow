package cleaner

import (
	"time"

	"github.com/ismailtsdln/burrow/internal/history"
	"github.com/ismailtsdln/burrow/internal/rules"
)

// Cleaner coordinates the cleanup process.
type Cleaner struct {
	trashManager *TrashManager
}

// NewCleaner creates a new cleaner instance.
func NewCleaner() *Cleaner {
	return &Cleaner{
		trashManager: NewTrashManager(),
	}
}

// CleanResult contains the summary of the cleanup action.
type CleanResult struct {
	ReclaimedSpace int64
	FileCount      int
	TrashSession   string
}

// Clean executes the cleanup of the provided results.
func (c *Cleaner) Clean(results []rules.Result, dryRun bool) (*CleanResult, error) {
	var totalSpace int64
	var totalPaths []string

	categoryStats := make(map[string]int64)

	for _, res := range results {
		totalSpace += res.TotalSize
		totalPaths = append(totalPaths, res.FoundPaths...)
		categoryStats[res.Rule.Category] += res.TotalSize
	}

	if dryRun {
		return &CleanResult{
			ReclaimedSpace: totalSpace,
			FileCount:      len(totalPaths),
			TrashSession:   "DRY-RUN",
		}, nil
	}

	session, err := c.trashManager.MoveToTrash(totalPaths)
	if err != nil {
		return nil, err
	}

	// Save to history
	histMgr := history.NewManager()
	histMgr.Save(history.Entry{
		ID:             session,
		Timestamp:      time.Now(),
		ReclaimedBytes: totalSpace,
		FileCount:      len(totalPaths),
		CategoryStats:  categoryStats,
	})

	return &CleanResult{
		ReclaimedSpace: totalSpace,
		FileCount:      len(totalPaths),
		TrashSession:   session,
	}, nil
}

// Undo restores the last cleanup session.
func (c *Cleaner) Undo() error {
	return c.trashManager.RestoreLast()
}
