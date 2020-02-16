package action

import (
	"github.com/projecteru2/aa/errors"
)

// Actions .
type Actions []Action

// Check .
func (as Actions) Check() error {
	if as.Len() < 1 {
		return errors.Annotatef(errors.ErrInvalidValue, "actions are empty")
	}

	for _, act := range as {
		if err := act.Check(); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// Parse .
func (as *Actions) Parse(list interface{}) (err error) {
	raw, ok := list.([]interface{})
	if !ok {
		return errors.Annotatef(errors.ErrInvalidType,
			"must be []interface{} type, but %v", list)
	}

	acts := make(Actions, len(raw))
	for i, value := range raw {
		if acts[i], err = as.parseAction(value); err != nil {
			return
		}
	}

	*as = acts

	return
}

func (as *Actions) parseAction(dict interface{}) (Action, error) {
	raw, ok := dict.(map[string]interface{})
	if !ok {
		return nil, errors.Annotatef(errors.ErrInvalidType, "must be map[string]interface{} type")
	}

	atom := &Atom{}
	return atom, atom.Parse(raw)
}

// Add .
func (as *Actions) Add(act ...Action) {
	*as = append(*as, act...)
}

// Equal .
func (as Actions) Equal(other Actions) bool {
	if as.Len() != other.Len() {
		return false
	}

	for i := 0; i < as.Len(); i++ {
		if !as[i].Equal(other[i]) {
			return false
		}
	}

	return true
}

// Len .
func (as Actions) Len() int { return len(as) }

// Action .
type Action interface {
	Equal(other Action) bool
	Check() error

	Target() string
	IsDependency() bool

	GetAsync() bool
	GetImage() *Image
	GetCommand() *Command
	GetParams() Parameters

	SetImage(*Image)
}
