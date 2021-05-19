package grpc

import (
	"github.com/projecteru2/phistage/apiserver/grpc/proto"
	"github.com/projecteru2/phistage/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toGRPCRun(r *common.Run) *proto.Run {
	return &proto.Run{
		Id:       r.ID,
		Phistage: r.Phistage,
		Start:    timestamppb.New(r.Start),
		End:      timestamppb.New(r.End),
	}
}

func toGRPCJobRun(j *common.JobRun) *proto.JobRun {
	return &proto.JobRun{
		Id:       j.ID,
		Phistage: j.Phistage,
		Job:      j.Job,
		Status:   string(j.Status),
		Start:    timestamppb.New(j.Start),
		End:      timestamppb.New(j.End),
	}
}
