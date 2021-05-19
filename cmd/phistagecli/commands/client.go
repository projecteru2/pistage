package commands

import (
	"context"
	"time"

	"github.com/projecteru2/phistage/apiserver/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func newClient(ctx context.Context, addr string) (proto.PhistageClient, error) {
	connection, err := dial(ctx, addr)
	if err != nil {
		return nil, err
	}
	return proto.NewPhistageClient(connection), nil
}

func dial(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: 6 * 60 * time.Second, Timeout: time.Second}),
		grpc.WithBalancerName("round_robin"),
	}
	return grpc.DialContext(ctx, addr, opts...)
}
