package eru

import (
	"context"

	coreclient "github.com/projecteru2/core/client"
	corepb "github.com/projecteru2/core/rpc/gen"
	coretypes "github.com/projecteru2/core/types"
	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/executors"
)

type EruJobExecutorProvider struct {
	config *common.Config
	eru    corepb.CoreRPCClient
}

func NewEruJobExecutorProvider(config *common.Config) (*EruJobExecutorProvider, error) {
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
	}, nil
}

func (ep *EruJobExecutorProvider) GetName() string {
	return "eru"
}

func (ep *EruJobExecutorProvider) GetJobExecutor(job *common.Job, phistage *common.Phistage) (executors.JobExecutor, error) {
	return NewEruJobExecutor(job, phistage, ep.eru)
}
