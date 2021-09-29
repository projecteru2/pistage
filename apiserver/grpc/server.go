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

func (g *GRPCServer) SetVariables(ctx context.Context, req *proto.SetVariablesRequest) (*proto.SetVariablesReply, error) {
	err := g.store.SetVariablesForPistage(ctx, req.GetName(), req.GetVariables())
	return &proto.SetVariablesReply{
		Name:    req.GetName(),
		Success: err == nil,
	}, err
}

func (g *GRPCServer) GetVariables(ctx context.Context, req *proto.GetVariablesRequest) (*proto.GetVariablesReply, error) {
	vars, err := g.store.GetVariablesForPistage(ctx, req.GetName())
	if err != nil {
		return nil, err
	}
	return &proto.GetVariablesReply{
		Name:      req.GetName(),
		Variables: vars,
	}, nil
}

func (g *GRPCServer) ApplyOneway(ctx context.Context, req *proto.ApplyPistageRequest) (*proto.ApplyPistageOnewayReply, error) {
	pistage, err := common.FromSpec([]byte(req.GetContent()))
	if err != nil {
		return nil, err
	}

	// Discard the output
	g.stager.Add(&common.PistageTask{Pistage: pistage, Output: common.ClosableDiscard})
	return &proto.ApplyPistageOnewayReply{
		Name:    pistage.Name,
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
			Name: pistage.Name,
			Log:  scanner.Text(),
		}); err != nil {
			logrus.WithError(err).Error("[GRPCServer] error sending ApplyPistageStreamReply")
		}
	}
	return nil
}

func (g *GRPCServer) GetPistage(ctx context.Context, req *proto.GetPistageRequest) (*proto.GetPistageReply, error) {
	pistage, err := g.store.GetPistage(ctx, req.GetName())
	if err != nil {
		return nil, err
	}

	content, err := common.MarshalPistage(pistage)
	if err != nil {
		return nil, err
	}

	return &proto.GetPistageReply{
		Name:    pistage.Name,
		Content: string(content),
	}, nil
}

func (g *GRPCServer) DeletePistage(ctx context.Context, req *proto.DeletePistageRequest) (*proto.DeletePistageReply, error) {
	err := g.store.DeletePistage(ctx, req.GetName())
	return &proto.DeletePistageReply{
		Name:    req.GetName(),
		Success: err == nil,
	}, err
}

func (g *GRPCServer) GetRunsByPistage(ctx context.Context, req *proto.GetRunsByPistageRequest) (*proto.GetRunsByPistageReply, error) {
	runs, err := g.store.GetRunsByPistage(ctx, req.GetName())
	if err != nil {
		return nil, err
	}

	pbRuns := []*proto.Run{}
	for _, run := range runs {
		pbRuns = append(pbRuns, toGRPCRun(run))
	}
	return &proto.GetRunsByPistageReply{
		Name: req.GetName(),
		Runs: pbRuns,
	}, nil
}

func (g *GRPCServer) GetJobRunsByPistage(ctx context.Context, req *proto.GetJobRunsByPistageRequest) (*proto.GetJobRunsByPistageReply, error) {
	jobRuns, err := g.store.GetJobRuns(ctx, req.GetRunID())
	if err != nil {
		return nil, err
	}

	pbJobRuns := []*proto.JobRun{}
	for _, jobRun := range jobRuns {
		pbJobRuns = append(pbJobRuns, toGRPCJobRun(jobRun))
	}
	return &proto.GetJobRunsByPistageReply{
		Name:    req.GetName(),
		RunID:   req.GetRunID(),
		JobRuns: pbJobRuns,
	}, nil
}

func (g *GRPCServer) RegisterJob(ctx context.Context, req *proto.RegisterJobRequest) (*proto.RegisterJobReply, error) {
	job, err := common.LoadJob([]byte(req.GetContent()))
	if err != nil {
		return nil, err
	}

	if err := g.store.RegisterJob(ctx, job); err != nil {
		return nil, err
	}

	return &proto.RegisterJobReply{
		Name:    job.Name,
		Success: true,
	}, nil
}

func (g *GRPCServer) RegisterStep(ctx context.Context, req *proto.RegisterStepRequest) (*proto.RegisterStepReply, error) {
	step, err := common.LoadStep([]byte(req.GetContent()))
	if err != nil {
		return nil, err
	}

	if err := g.store.RegisterStep(ctx, step); err != nil {
		return nil, err
	}

	return &proto.RegisterStepReply{
		Name:    step.Name,
		Success: true,
	}, nil
}
