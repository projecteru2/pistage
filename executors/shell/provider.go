package shell

import (
	"io"

	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/executors"
	"github.com/projecteru2/pistage/store"
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

func (ls *ShellJobExecutorProvider) GetJobExecutor(job *common.Job, pistage *common.Pistage, output io.Writer) (executors.JobExecutor, error) {
	return NewShellJobExecutor(job, pistage, output, ls.store, ls.config)
}
