package run

import (
	"github.com/urfave/cli/v2"

	"github.com/projecteru2/aa/config"
	"github.com/projecteru2/aa/errors"
	"github.com/projecteru2/aa/executor"
	"github.com/projecteru2/aa/log"
	"github.com/projecteru2/aa/store"
)

// Run .
func Run(fn cli.ActionFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		filepaths := c.StringSlice("config")
		if err := setup(filepaths); err != nil {
			return errors.Trace(err)
		}

		if err := log.Setup(config.Conf.LogLevel, config.Conf.LogFile); err != nil {
			return errors.Trace(err)
		}

		err := fn(c)
		if err != nil {
			log.ErrorStack(err)
		}

		return err
	}
}

func setup(filepaths []string) error {
	if err := config.Conf.ParseFiles(filepaths...); err != nil {
		return errors.Trace(err)
	}

	return store.Setup(config.Conf.MetaType)
}

// NewExecutor .
func NewExecutor(c *cli.Context) executor.Executor {
	return executor.NewSimple()
}
