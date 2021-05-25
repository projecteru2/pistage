package commands

import (
	"context"
	"io/ioutil"

	"github.com/projecteru2/phistage/apiserver/grpc/proto"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func registerStep(c *cli.Context) error {
	content, err := ioutil.ReadFile(c.String("file"))
	if err != nil {
		return err
	}

	client, err := newClient(c)
	if err != nil {
		return err
	}

	reply, err := client.RegisterStep(context.TODO(), &proto.RegisterStepRequest{Content: string(content)})
	if err != nil {
		return err
	}

	if !reply.GetSuccess() {
		logrus.Error("Failed to register step")
	}
	return nil
}

func StepCommands() *cli.Command {
	return &cli.Command{
		Name:  "step",
		Usage: "Control of steps",
		Subcommands: []*cli.Command{
			{
				Name:  "register",
				Usage: "Register a step for others to use",
				Action: func(c *cli.Context) error {
					return registerStep(c)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Required: true,
						Usage:    "File contains the step, in yaml format",
					},
				},
			},
		},
	}
}
