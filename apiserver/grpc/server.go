package grpc

import (
	"context"
	"net"

	"github.com/projecteru2/phistage/apiserver/grpc/proto"
	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/stager"
	"github.com/projecteru2/phistage/store"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	store  store.Store
	stager *stager.Stager

	server *grpc.Server
}

func NewGRPCServer(store store.Store, stager *stager.Stager) *GRPCServer {
	return &GRPCServer{
		store:  store,
		stager: stager,
	}
}

func (g *GRPCServer) Serve(ctx context.Context, l net.Listener, opts ...grpc.ServerOption) {
	g.server = grpc.NewServer(opts...)
	proto.RegisterPhistageServer(g.server, g)

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
	err := g.store.SetVariablesForPhistage(ctx, req.GetName(), req.GetVariables())
	return &proto.SetVariablesReply{
		Name:    req.GetName(),
		Success: err == nil,
	}, err
}

func (g *GRPCServer) GetVariables(ctx context.Context, req *proto.GetVariablesRequest) (*proto.GetVariablesReply, error) {
	vars, err := g.store.GetVariablesForPhistage(ctx, req.GetName())
	if err != nil {
		return nil, err
	}
	return &proto.GetVariablesReply{
		Name:      req.GetName(),
		Variables: vars,
	}, nil
}

func (g *GRPCServer) ApplyOneway(ctx context.Context, req *proto.ApplyPhistageRequest) (*proto.ApplyPhistageOnewayReply, error) {
	phistage, err := common.FromSpec([]byte(req.GetContent()))
	if err != nil {
		return nil, err
	}

	g.stager.Add(phistage)
	return &proto.ApplyPhistageOnewayReply{
		Name:    phistage.Name,
		Success: err == nil,
	}, err
}

func (g *GRPCServer) ApplyStream(req *proto.ApplyPhistageRequest, stream proto.Phistage_ApplyStreamServer) error {
	phistage, err := common.FromSpec([]byte(req.GetContent()))
	if err != nil {
		return err
	}

	g.stager.Add(phistage)
	// TODO get stream and send them back
	return nil
}

func (g *GRPCServer) GetPhistage(ctx context.Context, req *proto.GetPhistageRequest) (*proto.GetPhistageReply, error) {
	phistage, err := g.store.GetPhistage(ctx, req.GetName())
	if err != nil {
		return nil, err
	}

	content, err := common.MarshalPhistage(phistage)
	if err != nil {
		return nil, err
	}

	return &proto.GetPhistageReply{
		Name:    phistage.Name,
		Content: string(content),
	}, nil
}

func (g *GRPCServer) DeletePhistage(ctx context.Context, req *proto.DeletePhistageRequest) (*proto.DeletePhistageReply, error) {
	err := g.store.DeletePhistage(ctx, req.GetName())
	return &proto.DeletePhistageReply{
		Name:    req.GetName(),
		Success: err == nil,
	}, err
}

func (g *GRPCServer) GetRunsByPhistage(ctx context.Context, req *proto.GetRunsByPhistageRequest) (*proto.GetRunsByPhistageReply, error) {
	runs, err := g.store.GetRunsByPhistage(ctx, req.GetName())
	if err != nil {
		return nil, err
	}

	pbRuns := []*proto.Run{}
	for _, run := range runs {
		pbRuns = append(pbRuns, toGRPCRun(run))
	}
	return &proto.GetRunsByPhistageReply{
		Name: req.GetName(),
		Runs: pbRuns,
	}, nil
}

func (g *GRPCServer) GetJobRunsByPhistage(ctx context.Context, req *proto.GetJobRunsByPhistageRequest) (*proto.GetJobRunsByPhistageReply, error) {
	jobRuns, err := g.store.GetJobRuns(ctx, req.GetRunID())
	if err != nil {
		return nil, err
	}

	pbJobRuns := []*proto.JobRun{}
	for _, jobRun := range jobRuns {
		pbJobRuns = append(pbJobRuns, toGRPCJobRun(jobRun))
	}
	return &proto.GetJobRunsByPhistageReply{
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
