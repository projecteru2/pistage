package act

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/projecteru2/aa/action"
	"github.com/projecteru2/aa/cli/run"
	"github.com/projecteru2/aa/io"
)

// Command .
func Command() *cli.Command {
	return &cli.Command{
		Name: "action",
		Subcommands: []*cli.Command{
			{
				Name:   "register",
				Action: run.Run(register),
			},
			{
				Name:   "execute",
				Action: run.Run(execute),
			},
		},
	}
}

func execute(c *cli.Context) error {
	complex, err := parseComplex(c)
	if err != nil {
		return err
	}

	_, err = run.NewExecutor(c).SyncStart(context.Background(), complex)
	return err
}

func register(c *cli.Context) error {
	complex, err := parseComplex(c)
	if err != nil {
		return err
	}

	if err := complex.Save(context.Background()); err != nil {
		return err
	}

	fmt.Printf("%s has been registered\n", complex.Name)

	return nil
}

func parseComplex(c *cli.Context) (*action.Complex, error) {
	cont, err := readSpec(c)
	if err != nil {
		return nil, err
	}

	return action.Parse(string(cont))
}

func readSpec(c *cli.Context) (string, error) {
	specFile := c.Args().First()
	if len(specFile) < 1 {
		return "", fmt.Errorf("spec filepath is required")
	}

	buf, err := io.ReadFile(specFile)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
