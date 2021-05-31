package shell

import (
	"io"

	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/executors"
	"github.com/projecteru2/phistage/store"
)

type ShellJobExecutorProvider struct {
	config *common.Config
	store  store.Store
}

func NewShellJobExecutorProvider(config *common.Config, store store.Store) (*ShellJobExecutorProvider, error) {
	return &ShellJobExecutorProvider{
		config: config,
		store:  store,
	}, nil
}

func (ls *ShellJobExecutorProvider) GetName() string {
	return "shell"
}

func (ls *ShellJobExecutorProvider) GetJobExecutor(job *common.Job, phistage *common.Phistage, output io.Writer) (executors.JobExecutor, error) {
	return NewShellJobExecutor(job, phistage, output, ls.store, ls.config)
}
