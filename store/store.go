package store

import (
	"context"

	"github.com/projecteru2/phistage/common"
)

// Store is the interface for storage.
type Store interface {
	// Phistage
	CreatePhistage(ctx context.Context, phistage *common.Phistage) error
	GetPhistage(ctx context.Context, name string) (*common.Phistage, error)
	DeletePhistage(ctx context.Context, name string) error

	// Run
	CreateRun(ctx context.Context, run *common.Run) error
	GetRun(ctx context.Context, id string) (*common.Run, error)
	UpdateRun(ctx context.Context, run *common.Run) error
	GetRunsByPhistage(ctx context.Context, name string) ([]*common.Run, error)

	// JobRun
	CreateJobRun(ctx context.Context, run *common.Run, jobRun *common.JobRun) error
	GetJobRun(ctx context.Context, runID, jobRunID string) (*common.JobRun, error)
	UpdateJobRun(ctx context.Context, run *common.Run, jobRun *common.JobRun) error
	FinishJobRun(ctx context.Context, run *common.Run, jobRun *common.JobRun) error
	GetJobRuns(ctx context.Context, runID string) ([]*common.JobRun, error)

	// Register
	RegisterJob(ctx context.Context, job *common.Job) error
	GetRegisteredJob(ctx context.Context, name string) (*common.Job, error)
	RegisterStep(ctx context.Context, step *common.Step) error
	GetRegisteredStep(ctx context.Context, name string) (*common.Step, error)

	// Variables
	SetVariablesForPhistage(ctx context.Context, name string, vars map[string]string) error
	GetVariablesForPhistage(ctx context.Context, name string) (map[string]string, error)
}
