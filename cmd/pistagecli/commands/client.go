package commands

import (
	"context"
	"time"

	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/projecteru2/pistage/apiserver/grpc/proto"
)

func newClient(c *cli.Context) (proto.PistageClient, error) {
	connection, err := dial(c.Context, c.String("host"))
	if err != nil {
		return nil, err
	}
	return proto.NewPistageClient(connection), nil
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
