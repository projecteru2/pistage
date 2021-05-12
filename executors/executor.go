package executor

import (
	"context"
)

type PhistageExecutor interface {
	Execute(ctx context.Context) error
}

type JobExecutor interface {
	Prepare(ctx context.Context) error
	Execute(ctx context.Context) error
	Cleanup(ctx context.Context) error
}
