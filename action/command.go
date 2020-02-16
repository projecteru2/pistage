package action

import (
	"bytes"
	"text/template"

	"github.com/projecteru2/aa/errors"
)

// Command indicates a command within an image.
type Command struct {
	Raw string
}

// NewCommand .
func NewCommand(raw string) *Command {
	return &Command{Raw: raw}
}

// Check .
func (c *Command) Check() error {
	if len(c.Raw) < 1 {
		return errors.Annotatef(errors.ErrInvalidValue, "raw is empty")
	}
	return nil
}

// Equal .
func (c *Command) Equal(other *Command) bool {
	return c.Raw == other.Raw
}

// Program renders from raw and replaces all placeholders.
func (c *Command) Program(params Parameters) (string, error) {
	tmpl, err := template.New("prog").Parse(c.Raw)
	if err != nil {
		return "", errors.Trace(err)
	}

	var wr bytes.Buffer
	if err := tmpl.Execute(&wr, params.dict()); err != nil {
		return "", errors.Trace(err)
	}

	return wr.String(), nil
}
