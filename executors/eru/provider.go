package eru

import (
	"context"
	"io"

	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/executors"
	"github.com/projecteru2/pistage/store"

	coreclient "github.com/projecteru2/core/client"
	corepb "github.com/projecteru2/core/rpc/gen"
	coretypes "github.com/projecteru2/core/types"
)

type EruJobExecutorProvider struct {
	config *common.Config
	eru    corepb.CoreRPCClient
	store  store.Store
}

func NewEruJobExecutorProvider(config *common.Config, store store.Store) (*EruJobExecutorProvider, error) {
	c, err := coreclient.NewClient(context.TODO(), config.Eru.Address, coretypes.AuthConfig{
		Username: config.Eru.Username,
		Password: config.Eru.Password,
	})
	if err != nil {
		return nil, err
	}

	return &EruJobExecutorProvider{
		config: config,
		eru:    c.GetRPCClient(),
		store:  store,
	}, nil
}

func (ep *EruJobExecutorProvider) GetName() string {
	return "eru"
}

func (ep *EruJobExecutorProvider) GetJobExecutor(job *common.Job, pistage *common.Pistage, output io.Writer) (executors.JobExecutor, error) {
	return NewEruJobExecutor(job, pistage, output, ep.eru, ep.store, ep.config)
}
