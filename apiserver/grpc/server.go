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
	g.stager.Add(&common.PistageTask{Pistage: pistage, JobType: common.Apply, Output: common.ClosableDiscard})
	return &proto.ApplyPistageOnewayReply{
		WorkflowNamespace:  pistage.WorkflowNamespace,
		WorkflowIdentifier: pistage.WorkflowIdentifier,
		Success:            err == nil,
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
	g.stager.Add(&common.PistageTask{Pistage: pistage, JobType: common.Apply, Output: common.DonCloseWriter{Writer: w}})

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if err := stream.Send(&proto.ApplyPistageStreamReply{
			WorkflowNamespace:  pistage.WorkflowNamespace,
			WorkflowIdentifier: pistage.WorkflowIdentifier,
			Log:                scanner.Text(),
		}); err != nil {
			logrus.WithError(err).Error("[GRPCServer] error sending ApplyPistageStreamReply")
		}
	}
	return nil
}

func (g *GRPCServer) RollbackOneway(ctx context.Context, req *proto.RollbackPistageRequest) (*proto.RollbackReply, error) {
	pistage, err := common.FromSpec([]byte(req.GetContent()))
	if err != nil {
		return nil, err
	}
	// Discard the output
	g.stager.Add(&common.PistageTask{Pistage: pistage, JobType: common.Rollback, Output: common.ClosableDiscard})
	return &proto.RollbackReply{
		WorkflowNamespace:  pistage.WorkflowNamespace,
		WorkflowIdentifier: pistage.WorkflowIdentifier,
		Success:            err == nil,
	}, err
}

func (g *GRPCServer) RollbackStream(req *proto.RollbackPistageRequest, stream proto.Pistage_RollbackStreamServer) error {
	pistage, err := common.FromSpec([]byte(req.GetContent()))
	if err != nil {
		return err
	}

	// generate output
	r, w := io.Pipe()
	g.stager.Add(&common.PistageTask{Pistage: pistage, JobType: common.Rollback, Output: common.DonCloseWriter{Writer: w}})

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		if err := stream.Send(&proto.RollbackPistageStreamReply{
			WorkflowNamespace:  pistage.WorkflowNamespace,
			WorkflowIdentifier: pistage.WorkflowIdentifier,
			Log:                scanner.Text(),
		}); err != nil {
			logrus.WithError(err).Error("[GRPCServer] error sending ApplyPistageStreamReply")
		}
	}
	return nil
}
