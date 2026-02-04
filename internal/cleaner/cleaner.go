package cleaner

import (
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

	for _, res := range results {
		totalSpace += res.TotalSize
		totalPaths = append(totalPaths, res.FoundPaths...)
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
