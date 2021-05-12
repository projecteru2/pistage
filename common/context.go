package common

import (
	"context"
)

// FileCollector collects files in workload.
// Then copy the files to target workload when necessary.
type FileCollector interface {
	// Collect collects files with given paths from workload identified by workloadID
	Collect(ctx context.Context, workloadID string, files []string) error
	// CopyTo copies files to the workload identified by workloadID
	CopyTo(ctx context.Context, workloadID string, files []string) error
	// Files returns all the file names
	Files() []string
}
