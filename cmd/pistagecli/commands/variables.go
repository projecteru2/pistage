package commands

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/projecteru2/pistage/apiserver/grpc/proto"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var errorNoPistageSpecified = errors.New("need to specify Pistage name")

func setVariables(c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return errorNoPistageSpecified
	}

	content, err := ioutil.ReadFile(c.String("file"))
	if err != nil {
		return err
	}

	vars := map[string]string{}
	if err := yaml.Unmarshal(content, vars); err != nil {
		return err
	}

	client, err := newClient(c)
	if err != nil {
		return err
	}

	reply, err := client.SetVariables(context.TODO(), &proto.SetVariablesRequest{
		Name:      name,
		Variables: vars,
	})
	if err != nil {
		return err
	}

	if !reply.GetSuccess() {
		logrus.Error("Failed to set variables")
	}
	return nil
}

func getVariables(c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return errorNoPistageSpecified
	}

	client, err := newClient(c)
	if err != nil {
		return err
	}

	reply, err := client.GetVariables(context.TODO(), &proto.GetVariablesRequest{
		Name: name,
	})
	if err != nil {
		return err
	}

	content, err := yaml.Marshal(reply.GetVariables())
	if err != nil {
		return err
	}

	fmt.Println(string(content))
	return nil
}

func VariablesCommands() *cli.Command {
	return &cli.Command{
		Name:  "vars",
		Usage: "Control variables of a pistage",
		Subcommands: []*cli.Command{
			{
				Name:  "get",
				Usage: "Get variables of a pistage",
				Action: func(c *cli.Context) error {
					return getVariables(c)
				},
			},
			{
				Name:  "set",
				Usage: "Set variables of a pistage",
				Action: func(c *cli.Context) error {
					return setVariables(c)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Required: true,
						Usage:    "File contains the variables, format as an environment variables file",
					},
				},
			},
		},
	}
}
