package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/projecteru2/aa/cli/act"
	"github.com/projecteru2/aa/errors"
	"github.com/projecteru2/aa/ver"
)

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(ver.Version())
	}

	app := cli.App{
		Commands: []*cli.Command{
			act.Command(),
		},
		Flags:   globalFlags(),
		Version: "v",
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(errors.Stack(err))
	}
}

func globalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:     "config",
			Usage:    "config files",
			Required: true,
		},
	}
}
