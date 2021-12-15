package grpc

import (
	"bufio"
	"context"
	"io"
	"net"
	"time"

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

	server  *grpc.Server
	timeout time.Duration
}

func NewGRPCServer(store store.Store, stager *stageserver.StageServer, timeoutSecs int) *GRPCServer {
	return &GRPCServer{
		store:   store,
		stager:  stager,
		timeout: time.Duration(timeoutSecs) * time.Second,
	}
}

func (g *GRPCServer) Serve(l net.Listener, opts ...grpc.ServerOption) {
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

	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	// Discard the output
	g.stager.Add(&common.PistageTask{Ctx: ctx, Pistage: pistage, JobType: common.JobTypeApply, Output: common.ClosableDiscard})

	return &proto.ApplyPistageOnewayReply{
		WorkflowType:       pistage.WorkflowType,
		WorkflowIdentifier: pistage.WorkflowIdentifier,
		Success:            err == nil,
	}, err
}

func (g *GRPCServer) ApplyStream(req *proto.ApplyPistageRequest, stream proto.Pistage_ApplyStreamServer) error {
	pistage, err := common.FromSpec([]byte(req.GetContent()))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(stream.Context(), g.timeout)
	defer cancel()

	// We use a pipe here to retrieve the logs across all jobs within this pistage.
	// Use common.DonCloseWriter to avoid writing end of the pipe being closed by LogTracer.
	// It's a bit tricky here...
	r, w := io.Pipe()
	g.stager.Add(&common.PistageTask{Ctx: ctx, Pistage: pistage, JobType: common.JobTypeApply, Output: common.DonCloseWriter{Writer: w}})

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if err := stream.Send(&proto.ApplyPistageStreamReply{
			WorkflowType:       pistage.WorkflowType,
			WorkflowIdentifier: pistage.WorkflowIdentifier,
			Log:                scanner.Text(),
		}); err != nil {
			logrus.WithError(err).Error("[GRPCServer] error sending ApplyPistageStreamReply")
			return err
		}
	}

	return nil
}

func (g *GRPCServer) RollbackOneway(ctx context.Context, req *proto.RollbackPistageRequest) (*proto.RollbackReply, error) {
	pistage, err := common.FromSpec([]byte(req.GetContent()))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	// Discard the output
	g.stager.Add(&common.PistageTask{Ctx: ctx, Pistage: pistage, JobType: common.JobTypeRollback, Output: common.ClosableDiscard})

	return &proto.RollbackReply{
		WorkflowType:       pistage.WorkflowType,
		WorkflowIdentifier: pistage.WorkflowIdentifier,
		Success:            err == nil,
	}, err
}

func (g *GRPCServer) RollbackStream(req *proto.RollbackPistageRequest, stream proto.Pistage_RollbackStreamServer) error {
	pistage, err := common.FromSpec([]byte(req.GetContent()))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(stream.Context(), g.timeout)
	defer cancel()

	// generate output
	r, w := io.Pipe()
	g.stager.Add(&common.PistageTask{Ctx: ctx, Pistage: pistage, JobType: common.JobTypeRollback, Output: common.DonCloseWriter{Writer: w}})

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if err := stream.Send(&proto.RollbackPistageStreamReply{
			WorkflowType:       pistage.WorkflowType,
			WorkflowIdentifier: pistage.WorkflowIdentifier,
			Log:                scanner.Text(),
		}); err != nil {
			logrus.WithError(err).Error("[GRPCServer] error sending ApplyPistageStreamReply")
			return err
		}
	}
	return nil
}

func (g *GRPCServer) GetWorkflowRuns(ctx context.Context, req *proto.GetWorkflowDetailsRequest) (*proto.GetWorkflowDetailsReply, error) {
	workflowRuns, err := g.store.GetPistageRunsByWorkflowIdentifier(req.WorkflowIdentifier)
	if err != nil {
		return nil, err
	}

	runs := make([]*proto.WorkflowRun, 0, len(workflowRuns))
	for _, workflowRun := range workflowRuns {
		runs = append(runs, &proto.WorkflowRun{
			StartTime:    workflowRun.Start,
			EndTime:      workflowRun.End,
			WorkflowType: workflowRun.WorkflowType,
			Status:       string(workflowRun.Status),
		})
	}

	return &proto.GetWorkflowDetailsReply{
		WorkflowIdentifier: req.WorkflowIdentifier,
		Runs:               runs,
	}, nil
}
