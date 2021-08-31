package main

import (
	"fmt"
	"os"

	"github.com/projecteru2/pistage/cmd/pistage/version"
	"github.com/projecteru2/pistage/cmd/pistagecli/commands"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(version.Version())
	}

	app := &cli.App{
		Name:    "pistagecli",
		Version: version.VERSION,
		Commands: []*cli.Command{
			commands.ApplyCommands(),
			commands.VariablesCommands(),
			commands.JobCommands(),
			commands.StepCommands(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "host",
				Aliases: []string{"H"},
				Value:   ":9736",
				Usage:   "Pistage address",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Errorln("Failed to run pistagecli")
		return
	}
}
