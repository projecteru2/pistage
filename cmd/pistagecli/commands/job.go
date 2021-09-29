package commands

import (
	"context"
	"io/ioutil"

	"github.com/projecteru2/pistage/apiserver/grpc/proto"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func registerJob(c *cli.Context) error {
	content, err := ioutil.ReadFile(c.String("file"))
	if err != nil {
		return err
	}

	client, err := newClient(c)
	if err != nil {
		return err
	}

	reply, err := client.RegisterJob(context.TODO(), &proto.RegisterJobRequest{Content: string(content)})
	if err != nil {
		return err
	}

	if !reply.GetSuccess() {
		logrus.Error("Failed to register job")
	}
	return nil
}

func JobCommands() *cli.Command {
	return &cli.Command{
		Name:  "job",
		Usage: "Control of jobs",
		Subcommands: []*cli.Command{
			{
				Name:  "register",
				Usage: "Register a job for others to use",
				Action: func(c *cli.Context) error {
					return registerJob(c)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Required: true,
						Usage:    "File contains the job, in yaml format",
					},
				},
			},
		},
	}
}
