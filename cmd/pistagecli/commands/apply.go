package commands

import (
	"context"
	"io"
	"io/ioutil"

	"github.com/projecteru2/pistage/apiserver/grpc/proto"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func applyOneway(c *cli.Context) error {
	content, err := ioutil.ReadFile(c.String("file"))
	if err != nil {
		return err
	}

	client, err := newClient(c)
	if err != nil {
		return err
	}

	reply, err := client.ApplyOneway(context.TODO(), &proto.ApplyPistageRequest{Content: string(content)})
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

func applyStream(c *cli.Context) error {
	content, err := ioutil.ReadFile(c.String("file"))
	if err != nil {
		return err
	}

	client, err := newClient(c)
	if err != nil {
		return err
	}

	stream, err := client.ApplyStream(context.TODO(), &proto.ApplyPistageRequest{Content: string(content)})
	if err != nil {
		return err
	}

	for {
		message, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		logrus.Infof("[%s:%s] %s", message.WorkflowNamespace, message.WorkflowIdentifier, message.Log)
	}
	return nil
}

func apply(c *cli.Context) error {
	if c.Bool("stream") {
		return applyStream(c)
	}
	return applyOneway(c)
}

func rollback(c *cli.Context) error {
	if c.Bool("stream") {
		return rollbackStream(c)
	}
	return rollbackOneway(c)
}

func rollbackOneway(c *cli.Context) error {
	content, err := ioutil.ReadFile(c.String("file"))
	if err != nil {
		return err
	}

	client, err := newClient(c)
	if err != nil {
		return err
	}

	reply, err := client.RollbackOneway(context.TODO(), &proto.RollbackPistageRequest{Content: string(content)})
	if err != nil {
		return err
	}

	if reply.GetSuccess() {
		logrus.Info("Rollbacked")
	} else {
		logrus.Error("Failed to Rollback")
	}
	return nil
}

func rollbackStream(c *cli.Context) error {
	content, err := ioutil.ReadFile(c.String("file"))
	if err != nil {
		return err
	}

	client, err := newClient(c)
	if err != nil {
		return err
	}

	stream, err := client.RollbackStream(context.TODO(), &proto.RollbackPistageRequest{Content: string(content)})
	if err != nil {
		return err
	}

	for {
		message, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		logrus.Infof("[%s:%s] %s", message.WorkflowNamespace, message.WorkflowIdentifier, message.Log)
	}
	return nil
}

func ApplyCommands() *cli.Command {
	return &cli.Command{
		Name:  "apply",
		Usage: "Apply a Pistage",
		Action: func(c *cli.Context) error {
			return apply(c)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Value:   "pistage.yml",
				Usage:   "Pistage yaml description file",
			},
			&cli.BoolFlag{
				Name:  "stream",
				Value: false,
				Usage: "If set, will wait and print all the logs from pistage",
			},
		},
	}
}

func RollbackCommands() *cli.Command {
	return &cli.Command{
		Name:  "rollback",
		Usage: "rollback some steps",
		Action: func(c *cli.Context) error {
			return rollback(c)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Value:   "pistage.yml",
				Usage:   "Pistage yaml description file",
			},
			&cli.BoolFlag{
				Name:  "stream",
				Value: false,
				Usage: "If set, will wait and print all the logs from pistage",
			},
		},
	}
}
