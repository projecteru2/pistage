package action

import (
	"github.com/projecteru2/aa/codec/json"
	"github.com/projecteru2/aa/errors"
)

// MainGroupName indicates the default group name.
const MainGroupName = "main"

// Groups .
type Groups map[string]Actions

// UnmarshalJSON .
func (gs *Groups) UnmarshalJSON(data []byte) error {
	dict := map[string]interface{}{}
	if err := json.Decode(data, &dict); err != nil {
		return errors.Trace(err)
	}

	return gs.Parse(dict)
}

// Check .
func (gs Groups) Check() error {
	if gs.Len() < 1 {
		return errors.Annotatef(errors.ErrInvalidValue, "groups are empty")
	}

	for _, acts := range gs {
		if err := acts.Check(); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// Parse parses Groups.
func (gs *Groups) Parse(raw interface{}) error {
	dict, ok := raw.(map[string]interface{})
	if !ok {
		return errors.Annotatef(errors.ErrInvalidType,
			"must be map[string]interface{} type, but %v", raw)
	}

	return gs.parse(dict)
}

func (gs *Groups) parse(dict map[string]interface{}) error {
	groups := Groups{}
	for name, value := range dict {
		acts := Actions{}
		if err := acts.Parse(value); err != nil {
			return errors.Trace(err)
		}

		if err := groups.Set(name, acts); err != nil {
			return errors.Trace(err)
		}
	}

	*gs = groups

	return nil
}

// Set .
func (gs *Groups) Set(name string, acts Actions) error {
	if _, exists := (*gs)[name]; exists {
		return errors.Annotatef(errors.ErrKeyExists, name)
	}

	(*gs)[name] = acts

	return nil
}

// Equal .
func (gs Groups) Equal(other Groups) bool {
	if gs.Len() != other.Len() {
		return false
	}

	for nm, acts := range gs {
		switch o, exists := other[nm]; {
		case !exists:
			fallthrough

		case !acts.Equal(o):
			return false
		}
	}

	return true
}

// Len .
func (gs Groups) Len() int { return len(gs) }
