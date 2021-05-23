package main

import (
	"fmt"
	"os"

	"github.com/projecteru2/phistage/cmd/phistage/version"
	"github.com/projecteru2/phistage/cmd/phistagecli/commands"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(version.Version())
	}

	app := &cli.App{
		Name:    "phistagecli",
		Version: version.VERSION,
		Commands: []*cli.Command{
			commands.ApplyCommand(),
			commands.VariablesCommand(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "host",
				Aliases: []string{"H"},
				Value:   ":9736",
				Usage:   "Phistage address",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Errorln("Failed to run phistagecli")
		return
	}
}
