package action

import (
	"github.com/projecteru2/aa/errors"
)

// KV indicates key/value parameter pair.
type KV struct {
	Key   string
	Value interface{}
}

// NewKV .
func NewKV(k string, v interface{}) KV {
	return KV{
		Key:   k,
		Value: v,
	}
}

// Check .
func (kv KV) Check() error {
	if len(kv.Key) < 1 {
		return errors.Annotatef(errors.ErrInvalidValue, "key is empty")
	}
	return nil
}

// GetName .
func (kv KV) GetName() string { return kv.Key }

// GetValue .
func (kv KV) GetValue() interface{} { return kv.Value }

// Equal .
func (kv KV) Equal(other Parameter) bool {
	o, ok := other.(KV)
	return ok && kv.Key == o.Key && kv.Value == o.Value
}
