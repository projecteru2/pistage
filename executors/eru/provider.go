package eru

import (
	"context"
	"io"

	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/executors"
	"github.com/projecteru2/phistage/store"

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
	c, err := coreclient.NewClient(context.TODO(), config.EruAddress, coretypes.AuthConfig{
		Username: config.EruUsername,
		Password: config.EruPassword,
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

func (ep *EruJobExecutorProvider) GetJobExecutor(job *common.Job, phistage *common.Phistage, output io.Writer) (executors.JobExecutor, error) {
	return NewEruJobExecutor(job, phistage, output, ep.eru, ep.store)
}
