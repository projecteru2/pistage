package grpc

import (
	"bufio"
	"context"
	"io"
	"net"

	"github.com/projecteru2/pistage/apiserver/grpc/proto"
	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/stageserver"
	"github.com/projecteru2/pistage/store"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	proto.UnimplementedPistageServer

	store  store.Store
	stager *stageserver.StageServer

	server *grpc.Server
}

const (
	Apply    = "apply"
	Rollback = "rollback"
)

func NewGRPCServer(store store.Store, stager *stageserver.StageServer) *GRPCServer {
	return &GRPCServer{
		store:  store,
		stager: stager,
	}
}

func (g *GRPCServer) Serve(ctx context.Context, l net.Listener, opts ...grpc.ServerOption) {
	g.server = grpc.NewServer(opts...)
	proto.RegisterPistageServer(g.server, g)

	if err := g.server.Serve(l); err != nil {
		logrus.WithError(err).Error("[GRPCServer] serve error")
	}
}

func (g *GRPCServer) Stop() {
	if g.server == nil {
		return
	}

	logrus.Info("[GRPCServer] exiting...")
	g.server.GracefulStop()
	logrus.Info("[GRPCServer] graceful stopped")
}

func (g *GRPCServer) ApplyOneway(ctx context.Context, req *proto.ApplyPistageRequest) (*proto.ApplyPistageOnewayReply, error) {
	pistage, err := common.FromSpec([]byte(req.GetContent()))
	if err != nil {
		return nil, err
	}

	// Discard the output
	g.stager.Add(&common.PistageTask{Pistage: pistage, Output: common.ClosableDiscard})
	return &proto.ApplyPistageOnewayReply{
		Name:    pistage.Name(),
		Success: err == nil,
	}, err
}

func (g *GRPCServer) ApplyStream(req *proto.ApplyPistageRequest, stream proto.Pistage_ApplyStreamServer) error {
	pistage, err := common.FromSpec([]byte(req.GetContent()))
	if err != nil {
		return err
	}

	// We use a pipe here to retrieve the logs across all jobs within this pistage.
	// Use common.DonCloseWriter to avoid writing end of the pipe being closed by LogTracer.
	// It's a bit tricky here...
	r, w := io.Pipe()
	g.stager.Add(&common.PistageTask{Pistage: pistage, Output: common.DonCloseWriter{Writer: w}})

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if err := stream.Send(&proto.ApplyPistageStreamReply{
			Name: pistage.Name(),
			Log:  scanner.Text(),
		}); err != nil {
			logrus.WithError(err).Error("[GRPCServer] error sending ApplyPistageStreamReply")
		}
	}
	return nil
}

func (g *GRPCServer) RollbackStream(req *proto.ApplyPistageRequest, stream proto.Pistage_RollbackStreamServer) error {
	pistage, err := common.FromSpec([]byte(req.GetContent()))
	if err != nil {
		return err
	}
	pistage.JobType = Rollback

	// generate output
	r, w := io.Pipe()
	g.stager.Add(&common.PistageTask{Pistage: pistage, Output: common.DonCloseWriter{Writer: w}})

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		if err := stream.Send(&proto.ApplyPistageStreamReply{
			Name: pistage.Name(),
			Log:  scanner.Text(),
		}); err != nil {
			logrus.WithError(err).Error("[GRPCServer] error sending ApplyPistageStreamReply")
		}
	}
	return nil

}
