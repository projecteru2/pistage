package main

import (
	"fmt"
	"os"

	"github.com/projecteru2/pistage/cmd/pistage/server"
	"github.com/projecteru2/pistage/cmd/pistage/version"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(version.Version())
	}

	app := &cli.App{
		Name:    "pistage",
		Version: version.VERSION,
		Commands: []*cli.Command{
			{
				Name:   "server",
				Usage:  "Run Pistage server",
				Action: server.StartPistage,
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
				Value:   "pistage.yml",
				Usage:   "Path to config file",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Errorln("Failed to run pistage")
		return
	}
}
