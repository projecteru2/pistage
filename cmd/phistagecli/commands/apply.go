package commands

import (
	"context"
	"io/ioutil"

	"github.com/projecteru2/phistage/apiserver/grpc/proto"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func Apply(c *cli.Context) error {
	content, err := ioutil.ReadFile(c.String("file"))
	if err != nil {
		return err
	}

	client, err := newClient(context.TODO(), c.String("host"))
	if err != nil {
		return err
	}

	reply, err := client.ApplyOneway(context.TODO(), &proto.ApplyPhistageRequest{Content: string(content)})
	if err != nil {
		return err
	}

	if reply.GetSuccess() {
		logrus.Info("Applied")
	} else {
		logrus.Error("Failed to apply")
	}
	return nil
}
