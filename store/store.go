package store

import (
	"context"

	"github.com/projecteru2/pistage/common"
)

// Store is the interface for storage.
type Store interface {
	// Snapshot
	CreatePistageSnapshot(pistage *common.Pistage) (string, error)
	GetPistageBySnapshotID(id string) (*common.Pistage, error)

	// Pistage
	CreatePistageRun(pistage *common.Pistage, version string) (string, error)
	GetPistageRun(id string) (*common.Run, error)
	UpdatePistageRun(run *common.Run) error
	GetLatestPistageRunByNamespaceAndFlowIdentifier(workflowNamespace string,
		workflowIdentifier string) (pistageRun *common.Run, err error)

	// JobRun
	CreateJobRun(run *common.Run, jobRun *common.JobRun) error
	GetJobRun(id string) (*common.JobRun, error)
	UpdateJobRun(jobRun *common.JobRun) error
	GetJobRunsByPistageRunId(id string) ([]*common.JobRun, error)

	// Register
	GetRegisteredKhoriumStep(ctx context.Context, name string) (*common.KhoriumStep, error)

	Close() error
}
