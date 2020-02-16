package action

import (
	"github.com/projecteru2/aa/codec/json"
	"github.com/projecteru2/aa/errors"
)

// Parameters .
type Parameters map[string]Parameter

// MarshalJSON .
func (ps Parameters) MarshalJSON() ([]byte, error) {
	flat := map[string]interface{}{}
	for name, param := range ps {
		flat[name] = param.GetValue()
	}

	return json.Encode(flat, "\t")
}

// UnmarshalJSON .
func (ps *Parameters) UnmarshalJSON(data []byte) error {
	dict := map[string]interface{}{}
	if err := json.Decode(data, &dict); err != nil {
		return errors.Trace(err)
	}

	return ps.Parse(dict)
}

// Check .
func (ps Parameters) Check() error {
	for _, p := range ps {
		if err := p.Check(); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func (ps Parameters) dict() map[string]interface{} {
	d := map[string]interface{}{}

	for _, p := range ps {
		d[p.GetName()] = p.GetValue()
	}

	return d
}

// Dump .
func (ps Parameters) Dump() (string, error) {
	enc, err := json.Encode(ps)
	if err != nil {
		return "", errors.Trace(err)
	}
	return string(enc), nil
}

// Update merges another one.
func (ps *Parameters) Update(other Parameters) {
	if *ps == nil && other != nil {
		*ps = Parameters{}
	}

	for k, v := range other {
		(*ps)[k] = v
	}
}

// Parse parses Parameters.
func (ps *Parameters) Parse(dict map[string]interface{}) (err error) {
	for k, v := range dict {
		switch {
		case containVariable(v):
			err = ps.parseVar(k, v)
		default:
			err = ps.parseKV(k, v)
		}

		if err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

func (ps *Parameters) parseVar(name string, str interface{}) error {
	raw, ok := str.(string)
	if !ok {
		return errors.Annotatef(errors.ErrInvalidType, "must be string type")
	}

	v := NewVar(name, raw)
	if err := v.Check(); err != nil {
		return errors.Trace(err)
	}

	(*ps)[name] = v

	return nil
}

func (ps *Parameters) parseKV(key string, value interface{}) error {
	kv := NewKV(key, value)
	if err := kv.Check(); err != nil {
		return errors.Trace(err)
	}

	(*ps)[key] = kv

	return nil
}

// Equal .
func (ps Parameters) Equal(other Parameters) bool {
	if ps.Len() != other.Len() {
		return false
	}

	for nm, param := range ps {
		switch o, exists := other[nm]; {
		case !exists:
			fallthrough
		case !param.Equal(o):
			return false
		}
	}

	return true
}

// Len .
func (ps Parameters) Len() int { return len(ps) }

// Parameter .
type Parameter interface {
	GetName() string
	GetValue() interface{}
	Equal(other Parameter) bool
	Check() error
}
