package shell

import (
	"io"

	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/executors"
	"github.com/projecteru2/phistage/store"
)

type LocalShellJobExecutorProvider struct {
	config *common.Config
	store  store.Store
}

func NewLocalShellJobExecutorProvider(config *common.Config, store store.Store) (*LocalShellJobExecutorProvider, error) {
	return &LocalShellJobExecutorProvider{
		config: config,
		store:  store,
	}, nil
}

func (ls *LocalShellJobExecutorProvider) GetName() string {
	return "shell"
}

func (ls *LocalShellJobExecutorProvider) GetJobExecutor(job *common.Job, phistage *common.Phistage, output io.Writer) (executors.JobExecutor, error) {
	return NewLocalShellJobExecutor(job, phistage, output, ls.store, ls.config)
}
