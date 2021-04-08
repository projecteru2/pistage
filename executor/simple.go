package executor

import (
	"context"

	"github.com/projecteru2/aa/action"
	"github.com/projecteru2/aa/errors"
	"github.com/projecteru2/aa/orch"
)

// Simple executor.
type Simple struct {
	orch  orch.Orchestrator
	store Store
}

// NewSimple .
func NewSimple() (*Simple, error) {
	eru, err := orch.NewEru()
	if err != nil {
		return nil, err
	}

	return &Simple{
		orch:  eru,
		store: &localStore{},
	}, nil
}

// AsyncStart .
func (s *Simple) AsyncStart(ctx context.Context, complex *action.Complex) (string, error) {
	return s.start(ctx, complex, true)
}

// SyncStart .
func (s *Simple) SyncStart(ctx context.Context, complex *action.Complex) (string, error) {
	return s.start(ctx, complex, false)
}

func (s *Simple) start(ctx context.Context, complex *action.Complex, async bool) (string, error) {
	jg, err := s.parse(complex)
	if err != nil {
		return "", errors.Trace(err)
	}

	if err := jg.save(ctx); err != nil {
		return "", errors.Trace(err)
	}

	if async {
		// TODO
		// actually, we should booting a labmda as a workload to waiting
		// the really executor has been done.
		go jg.run(ctx)
	} else {
		jg.run(ctx)
	}

	return jg.id, nil
}

func (s *Simple) parse(complex *action.Complex) (*JobGroup, error) {
	jg := NewJobGroup()
	jg.params = complex.Params
	jg.orch = s.orch
	jg.store = s.store

	for name, acts := range complex.Groups {
		if err := jg.add(name, acts); err != nil {
			return jg, errors.Trace(err)
		}
	}

	return jg, nil
}
