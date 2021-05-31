package helpers

import (
	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/executors"
	"github.com/projecteru2/phistage/executors/eru"
	"github.com/projecteru2/phistage/executors/shell"
	"github.com/projecteru2/phistage/executors/ssh"
	"github.com/projecteru2/phistage/store"
)

// initEru initializes eru executor provider.
func initEru(config *common.Config, store store.Store) error {
	eruProvider, err := eru.NewEruJobExecutorProvider(config, store)
	if err != nil {
		return err
	}
	executors.RegisterExecutorProvider(eruProvider)
	return nil
}

// initShell initializes shell executor provider.
func initShell(config *common.Config, store store.Store) error {
	localshellProvider, err := shell.NewShellJobExecutorProvider(config, store)
	if err != nil {
		return err
	}
	executors.RegisterExecutorProvider(localshellProvider)
	return nil
}

// initSSH initializes ssh executor provider.
func initSSH(config *common.Config, store store.Store) error {
	sshProvider, err := ssh.NewSSHJobExecutorProvider(config, store)
	if err != nil {
		return err
	}
	executors.RegisterExecutorProvider(sshProvider)
	return nil
}

var initializers = map[string]func(*common.Config, store.Store) error{
	"eru":   initEru,
	"shell": initShell,
	"ssh":   initSSH,
}

// InitExecutorProvider initiates and registers executor providers.
func InitExecutorProvider(config *common.Config, store store.Store) error {
	for _, provider := range config.JobExecutors {
		f, ok := initializers[provider]
		if !ok {
			continue
		}
		if err := f(config, store); err != nil {
			return err
		}
	}
	return nil
}
