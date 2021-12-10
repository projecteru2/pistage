package helpers

import (
	"context"

	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/executors"
	"github.com/projecteru2/pistage/executors/eru"
	"github.com/projecteru2/pistage/executors/shell"
	"github.com/projecteru2/pistage/executors/ssh"
	"github.com/projecteru2/pistage/store"
)

// initEru initializes eru executor provider.
func initEru( ctx context.Context, config *common.Config, store store.Store) error {
	eruProvider, err := eru.NewEruJobExecutorProvider(config, store, ctx)
	if err != nil {
		return err
	}
	executors.RegisterExecutorProvider(eruProvider)
	return nil
}

// initShell initializes shell executor provider.
func initShell(ctx context.Context, config *common.Config, store store.Store) error {
	localshellProvider, err := shell.NewShellJobExecutorProvider(config, store)
	if err != nil {
		return err
	}
	executors.RegisterExecutorProvider(localshellProvider)
	return nil
}

// initSSH initializes ssh executor provider.
func initSSH(ctx context.Context, config *common.Config, store store.Store) error {
	sshProvider, err := ssh.NewSSHJobExecutorProvider(config, store)
	if err != nil {
		return err
	}
	executors.RegisterExecutorProvider(sshProvider)
	return nil
}

var initializers = map[string]func(context.Context, *common.Config, store.Store) error{
	"eru":   initEru,
	"shell": initShell,
	"ssh":   initSSH,
}

// InitExecutorProvider initiates and registers executor providers.
func InitExecutorProvider( ctx context.Context, config *common.Config, store store.Store) error {
	for _, provider := range config.JobExecutors {
		f, ok := initializers[provider]
		if !ok {
			continue
		}
		if err := f(ctx, config, store); err != nil {
			return err
		}
	}
	return nil
}
