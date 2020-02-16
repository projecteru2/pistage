package orch

import (
	"context"
)

// Orchestrator .
type Orchestrator interface {
	Lambda(context.Context, LambdaOptions) (string, <-chan Message, error)
	Execute(context.Context, ExecuteOptions) (<-chan Message, error)

	GetContainerID(ctx context.Context, app, entry string, labels []string) (string, error)
}

// Message .
type Message struct {
	EOF   bool
	Error error
	ID    string
	Data  []byte
}
