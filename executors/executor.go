package executors

import (
	"context"
	"io"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/projecteru2/pistage/common"
)

type PistageExecutor interface {
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


	Rollback(ctx context.Context, jobName string) error

}

// ExecutorProvider is basically the factory of JobExecutor
// It can be registered with its name, which identifies the type.
type ExecutorProvider interface {
	// GetName returns the name of this ExecutorProvider.
	GetName() string

	// GetJobExecutor returns a JobExecutor with the given job and pistage,
	// all job executors in use should be generated from this method.
	GetJobExecutor(job *common.Job, pistage *common.Pistage, output io.Writer) (JobExecutor, error)
}

var executorProviders = make(map[string]ExecutorProvider)

// ErrorExecuteProviderNotFound is returned when fail to find executor provider
var ErrorExecuteProviderNotFound = errors.New("ExecutorProvider not found")

// RegisterExecutorProvider registers the executor provider with its name.
// Executor Providers with the same name can be registered for multiple times,
// latter registration will override former ones.
func RegisterExecutorProvider(ep ExecutorProvider) {
	executorProviders[ep.GetName()] = ep
	logrus.WithField("executor", ep.GetName()).Info("ExecutorProvider registered")
}

// GetExecutorProvider gets executor provider by the given name.
func GetExecutorProvider(name string) ExecutorProvider {
	return executorProviders[name]
}
