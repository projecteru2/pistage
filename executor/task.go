package executor

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/projecteru2/aa/action"
	"github.com/projecteru2/aa/errors"
	"github.com/projecteru2/aa/log"
	"github.com/projecteru2/aa/metrics"
	"github.com/projecteru2/aa/orch"
	"github.com/projecteru2/aa/uuid"
)

// Tasks .
type Tasks []*Task

// Len .
func (ts Tasks) Len() int { return len(ts) }

// Task .
type Task struct {
	ID         string        `json:"id"`
	Status     string        `json:"status"`
	Stdout     string        `json:"stdout,omitempty"`
	Stderr     string        `json:"stderr,omitempty"`
	ReturnCode int           `json:"return_code"`
	Action     action.Action `json:"action"`
	ResourceID string        `json:"resource_id,omitempty"`
	UUID       string        `json:"uuid"`

	wg   sync.WaitGroup
	orch orch.Orchestrator
	node *list.Element
}

// NewTask .
func NewTask(act action.Action, orch orch.Orchestrator) (t *Task, err error) {
	t = &Task{
		Status: StatusPending,
		Action: act,
		orch:   orch,
	}

	t.UUID, err = uuid.New()

	return
}

func (t *Task) wait() error {
	t.wg.Wait()
	return nil
}

func (t *Task) run(ctx context.Context) error {
	if err := t.save(ctx); err != nil {
		return errors.Trace(err)
	}

	if t.Action.GetImage() == nil {
		return t.execOnPrev(ctx)
	}

	return t.lambda(ctx)
}

func (t *Task) lambda(ctx context.Context) error {
	prog, err := t.getProg()
	if err != nil {
		return errors.Annotatef(err, "render `%s` prog failed", t.Action.GetCommand().Raw)
	}

	opts := orch.NewLambdaOptions(t.Action.GetImage().Name, prog, t.timeout())
	opts.Labels["uuid"] = t.UUID

	noti, err := t.orch.Lambda(ctx, opts)
	if err != nil {
		return errors.Trace(err)
	}
	log.Debugf("run a lambda for <%s> in %s", prog, t.ResourceID)

	if err := t.subscribe(ctx, noti); err != nil {
		return errors.Trace(err)
	}

	return t.save(ctx)
}

func (t *Task) execOnPrev(ctx context.Context) error {
	prog, err := t.getProg()
	if err != nil {
		return errors.Trace(err)
	}

	rid, err := prevResID(t.node)
	if err != nil {
		return errors.Trace(err)
	}

	t.ResourceID = rid
	log.Debugf("exec: %s on prev workload: %s", prog, t.ResourceID)

	opts := orch.NewExecuteOptions(t.ResourceID, "/", []string{prog}, []string{})
	noti, err := t.orch.Execute(ctx, opts)
	if err != nil {
		return errors.Trace(err)
	}

	if err := t.subscribe(ctx, noti); err != nil {
		return errors.Trace(err)
	}

	return t.save(ctx)
}

func (t *Task) subscribe(ctx context.Context, noti <-chan orch.Message) error {
	if !t.Action.GetAsync() {
		return t.doWait(ctx, noti)
	}

	t.wg.Add(1)

	go func() {
		defer t.wg.Done()

		if err := t.doWait(ctx, noti); err != nil {
			log.ErrorStack(err)
			metrics.IncrError()
		}
	}()

	return nil
}

func (t *Task) doWait(ctx context.Context, noti <-chan orch.Message) error {
	prog, err := t.getProg()
	if err != nil {
		return errors.Trace(err)
	}

	for msg := range noti {
		select {
		case <-ctx.Done():
			return errors.Trace(ctx.Err())
		default:
		}

		if t.ResourceID == "" && msg.ID != "" {
			t.ResourceID = msg.ID
		}
		switch {
		case msg.Error != nil:
			log.ErrorStack(msg.Error)
			fallthrough
		case msg.EOF:
			break
		}

		log.Debugf("lambda <%s> in %s: %s", prog, t.ResourceID, msg.Data)
	}

	return nil
}

func (t *Task) getProg() (string, error) {
	cmd := t.Action.GetCommand()
	params := t.Action.GetParams()
	return cmd.Program(params)
}

func (t *Task) save(ctx context.Context) error {
	// TODO
	return nil
}

func (t *Task) timeout() time.Duration {
	// TODO
	return time.Hour * 24
}
