package action

import (
	"github.com/projecteru2/aa/errors"
)

// Atom is the simplest Action.
type Atom struct {
	Dep        string     `json:"dep,omitempty"`
	ImageName  string     `json:"image,omitempty"`
	RawCommand string     `json:"run,omitempty"`
	Params     Parameters `json:"with,omitempty"`
	Async      bool       `json:"async,omitempty"`

	Image   *Image   `json:"-"`
	Command *Command `json:"-"`
}

// NewAtom .
func NewAtom(dep string, image *Image, command *Command, params Parameters) *Atom {
	return &Atom{
		Dep:     dep,
		Image:   image,
		Command: command,
		Params:  params,
	}
}

// IsDependency .
func (a *Atom) IsDependency() bool { return len(a.Dep) > 0 }

// Target returns the real action's name if it's a dep. Action.
func (a *Atom) Target() string { return a.Dep }

// GetAsync .
func (a *Atom) GetAsync() bool { return a.Async }

// SetImage .
func (a *Atom) SetImage(img *Image) { a.Image = img }

// GetImage .
func (a *Atom) GetImage() *Image { return a.Image }

// GetCommand .
func (a *Atom) GetCommand() *Command { return a.Command }

// GetParams .
func (a *Atom) GetParams() Parameters { return a.Params }

// Check .
func (a *Atom) Check() error {
	if a.Params != nil {
		if err := a.Params.Check(); err != nil {
			return errors.Trace(err)
		}
	}

	if len(a.Dep) > 0 {
		return nil
	}

	if a.Image != nil {
		if err := a.Image.Check(); err != nil {
			return errors.Trace(err)
		}
	}

	if a.Command == nil {
		return errors.Annotatef(errors.ErrInvalidValue, "run is empty")
	}
	if err := a.Command.Check(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

// Equal .
func (a *Atom) Equal(other Action) bool {
	switch o, ok := other.(*Atom); {
	case !ok:

	case a.Dep != o.Dep:
	case a.Async != o.Async:

	case a.ImageName != o.ImageName:
	case a.Image == nil && o.Image != nil:
	case a.Image != nil && !a.Image.Equal(o.Image):

	case a.RawCommand != o.RawCommand:
	case a.Command == nil && o.Command != nil:
	case a.Command != nil && !a.Command.Equal(o.Command):

	case a.Params == nil && o.Params != nil:
	case a.Params != nil && !a.Params.Equal(o.Params):

	default:
		return true
	}

	return false
}

// Parse parses Atom fields' value from a raw dict.
func (a *Atom) Parse(dict map[string]interface{}) (err error) {
	if a.Dep, _, err = mustString(dict, KeyDep); err != nil {
		return errors.Trace(err)
	}

	if raw, exits := dict[KeyAsync]; exits {
		var ok bool
		if a.Async, ok = raw.(bool); !ok {
			return errors.Annotatef(errors.ErrInvalidType, "expect bool, but %v", raw)
		}
	}

	switch a.ImageName, _, err = mustString(dict, KeyImage); {
	case err != nil:
		return errors.Trace(err)
	case len(a.ImageName) > 0:
		a.Image = NewImage(a.ImageName)
	}

	switch a.RawCommand, _, err = mustString(dict, KeyRun); {
	case err != nil:
		return errors.Trace(err)
	case len(a.RawCommand) > 0:
		a.Command = NewCommand(a.RawCommand)
	}

	if raw, exists := dict[KeyParams]; exists {
		if err = a.parseParams(raw); err != nil {
			return
		}
	}

	return
}

func (a *Atom) parseParams(raw interface{}) (err error) {
	dict, ok := raw.(map[string]interface{})
	if !ok {
		return errors.Annotatef(errors.ErrInvalidType,
			"must be map[string]interface{} type, but %v", raw)
	}

	params := Parameters{}
	if err = params.Parse(dict); err != nil {
		return
	}

	a.Params = params

	return nil
}
