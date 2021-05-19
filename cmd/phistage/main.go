package main

import (
	"fmt"
	"os"

	"github.com/projecteru2/phistage/cmd/phistage/server"
	"github.com/projecteru2/phistage/cmd/phistage/version"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(version.Version())
	}

	app := &cli.App{
		Name:    "phistage",
		Version: version.VERSION,
		Commands: []*cli.Command{
			{
				Name:  "server",
				Usage: "Run Phistage server",
				Action: func(c *cli.Context) error {
					return server.StartPhistage(c)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "log-level",
						Aliases: []string{"l"},
						Value:   "info",
						Usage:   "Log level for logging",
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "phistage.yml",
				Usage:   "Path to config file",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Errorln("Failed to run phistage")
		return
	}
}
