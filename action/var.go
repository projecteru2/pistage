package action

import (
	"regexp"

	"github.com/projecteru2/aa/errors"
)

const minVarLength = 5

// compiles for "{{ .x }}" within a string.
var varReg = regexp.MustCompile(`\{\{\s*\.\w+\s*\}\}`)

// Var indicates a variable parameter.
type Var struct {
	Raw   string      `json:"raw"`
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// NewVar .
func NewVar(name, raw string) Var {
	return Var{
		Raw:  raw,
		Name: name,
	}
}

// Check .
func (v Var) Check() error {
	switch {
	case len(v.Raw) < minVarLength:
		return errors.Annotatef(errors.ErrInvalidValue, "raw must be greater than %d", minVarLength)

	case len(v.Name) < 1:
		return errors.Annotatef(errors.ErrInvalidValue, "name is empty")

	default:
		return nil
	}
}

// GetName .
func (v Var) GetName() string { return v.Name }

// GetValue .
func (v Var) GetValue() interface{} { return v.Value }

// Equal .
func (v Var) Equal(other Parameter) bool {
	o, ok := other.(Var)
	return ok && v.Name == o.Name && v.Raw == o.Raw && v.Value == o.Value
}

func containVariable(str interface{}) bool {
	raw, ok := str.(string)
	if !ok {
		return false
	}
	return varReg.MatchString(raw)
}
