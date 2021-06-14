package common

import (
	"context"
)

// FileCollector collects files in workload.
// Then copy the files to target when necessary.
// We have several implementations:
//   - EruFileCollector
//   - ShellFileCollector
//   - SSHFileCollector
// For each implementation, refer to the code for the meaning of
// identifier and files.
type FileCollector interface {
	// Collect collects files with given paths from identifier
	Collect(ctx context.Context, identifier string, files []string) error
	// CopyTo copies files to identifier
	CopyTo(ctx context.Context, identifier string, files []string) error
	// Files returns all the file names
	Files() []string
}
