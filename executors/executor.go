package executors

import (
	"context"

	"github.com/projecteru2/phistage/common"
)

type PhistageExecutor interface {
	Execute(ctx context.Context) error
}

// JobExecutor is an executor to execute the job.
// It is designed stateful, so job is not a parameter to any methods.
// Each job is executed by an independent JobExecutor.
type JobExecutor interface {
	// Prepare does the preparation phase.
	// Usually creates container / virtual machine runtime,
	// setups the environment for job to run.
	Prepare(ctx context.Context) error
	// Execute does the execution phase.
	// Usually actually executes all the steps in the job.
	Execute(ctx context.Context) error
	// Cleanup does the clean up phase.
	// Usually it does cleaning work, collects necessary artifacts,
	// and remove the container / virtual machine runtime.
	Cleanup(ctx context.Context) error
}

// ExecutorProvider is basically the factory of JobExecutor
// It can be registered with its name, which identifies the type.
type ExecutorProvider interface {
	// GetName returns the name of this ExecutorProvider.
	GetName() string
	// GetJobExecutor returns a JobExecutor with the given job and phistage,
	// all job executors in use should be generated from this method.
	GetJobExecutor(job *common.Job, phistage *common.Phistage) (JobExecutor, error)
}

var executorProviders = make(map[string]ExecutorProvider)

// RegisterExecutorProvider registers the executor provider with given name.
// Executor Providers with the same name can be registered for multiple times,
// latter registration will override former ones.
func RegisterExecutorProvider(name string, ep ExecutorProvider) {
	executorProviders[name] = ep
}

// GetExecutorProvider gets executor provider by the given name.
func GetExecutorProvider(name string) ExecutorProvider {
	return executorProviders[name]
}
