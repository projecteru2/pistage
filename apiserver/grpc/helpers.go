package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/projecteru2/pistage/apiserver/grpc/proto"
	"github.com/projecteru2/pistage/common"
)

func toGRPCRun(r *common.Run) *proto.Run {
	return &proto.Run{
		Id:       r.ID,
		Pistage: r.Pistage,
		Start:    timestamppb.New(r.Start),
		End:      timestamppb.New(r.End),
	}
}

func toGRPCJobRun(j *common.JobRun) *proto.JobRun {
	return &proto.JobRun{
		Id:       j.ID,
		Pistage: j.Pistage,
		Job:      j.Job,
		Status:   string(j.Status),
		Start:    timestamppb.New(j.Start),
		End:      timestamppb.New(j.End),
	}
}
