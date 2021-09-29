package server

import (
	"net"

	"github.com/projecteru2/pistage/apiserver/grpc"
	"github.com/projecteru2/pistage/cmd/pistage/helpers"
	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/stageserver"

	"github.com/sethvargo/go-signalcontext"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func StartPistage(c *cli.Context) error {
	if err := helpers.SetupLog(c.String("log-level")); err != nil {
		return err
	}

	config, err := common.LoadConfigFromFile(c.String("config"))
	if err != nil {
		return err
	}

	store, err := helpers.InitStorage(config)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := helpers.InitExecutorProvider(config, store); err != nil {
		return err
	}

	l, err := net.Listen("tcp", config.Bind)
	if err != nil {
		return err
	}

	ctx, cancel := signalcontext.OnInterrupt()
	defer cancel()

	s := stageserver.NewStageServer(config, store)
	s.Start()
	logrus.Info("[Stager] started")

	g := grpc.NewGRPCServer(store, s)
	go g.Serve(ctx, l)
	logrus.Info("[GRPCServer] started")

	<-ctx.Done()
	g.Stop()
	s.Stop()
	return nil
}
