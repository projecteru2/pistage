package executor

import (
	"context"
	"sync"

	"github.com/projecteru2/aa/action"
	"github.com/projecteru2/aa/errors"
)

// Store .
type Store interface {
	GetComplex(context.Context, string) (*action.Complex, error)
}

type localStore struct {
	complexes sync.Map
}

// GetAction .
func (c *localStore) GetComplex(ctx context.Context, name string) (*action.Complex, error) {
	if v, exists := c.complexes.Load(name); exists {
		complex, ok := v.(*action.Complex)
		if !ok {
			c.complexes.Delete(name)
			return nil, errors.Annotatef(errors.ErrInvalidType, "%s is not an action.Action", name)
		}
		return complex, nil
	}

	complex, err := action.LoadComplex(name)
	if err != nil {
		return nil, errors.Trace(err)
	}

	c.complexes.Store(name, complex)

	return complex, nil
}
