package executor

import (
	"container/list"
	"context"
	"sync"

	"github.com/projecteru2/aa/action"
	"github.com/projecteru2/aa/config"
	"github.com/projecteru2/aa/errors"
	"github.com/projecteru2/aa/log"
	"github.com/projecteru2/aa/orch"
)

// JobGroup .
type JobGroup struct {
	id     string
	jobs   Jobs
	params action.Parameters
	orch   orch.Orchestrator
	store  Store
}

// NewJobGroup .
func NewJobGroup() *JobGroup {
	return &JobGroup{
		jobs:   Jobs{},
		params: action.Parameters{},
	}
}

func (g *JobGroup) add(name string, acts action.Actions) error {
	job := newJob(g.store, g.orch)
	if err := job.parse(acts); err != nil {
		return errors.Trace(err)
	}

	g.jobs[name] = job

	return nil
}

func (g *JobGroup) run(ctx context.Context) {
	log.Debugf("Starting a job group: %s", g.id)

	exit := make(chan struct{}, 1)
	notify := make(chan Message)

	// Start a monitor waiting upon all jobs running outcomes.
	mon := make(chan struct{}, 1)
	go func() {
		defer close(mon)
		defer close(notify)

		for {
			select {
			case <-exit:
				log.Infof("exit from monitor")
				return

			case msg := <-notify:
				log.ErrorStack(msg.Err)
			}
		}
	}()

	var wg sync.WaitGroup
	for _, job := range g.jobs {
		wg.Add(1)

		go func(j *Job) {
			defer wg.Done()
			if err := j.run(ctx); err != nil {
				notify <- Message{Err: err}
			}
		}(job)
	}
	wg.Wait()

	// Notifies the monitor preparing to exit.
	close(exit)

	// Makes sure the monitor has been terminated.
	<-mon
}

func (g *JobGroup) save(ctx context.Context) error {
	g.id = "id"

	md, err := g.meta()
	if err != nil {
		return errors.Trace(err)
	}

	return md.save(ctx)
}

func (g *JobGroup) meta() (*jobMeta, error) {
	md := newJobMeta()
	return md, md.parse(g)
}

// Jobs .
type Jobs map[string]*Job

// Len .
func (js Jobs) Len() int { return len(js) }

// DepJob .
type DepJob struct {
	*Job

	// It's very similar with Job, elements are Runner implement,
	// may be Task, and DepJob itself,
	paralled map[string]*list.List

	node *list.Element
}

func (dj *DepJob) wait() error {
	for _, link := range dj.paralled {
		if err := waitLink(link); err != nil {
			return err
		}
	}

	return nil
}

func (dj *DepJob) run(ctx context.Context) error {
	for _, link := range dj.paralled {
		if err := runLink(ctx, link); err != nil {
			// It doesn't deliberately trace the error to
			// avoid appearing a too deep error stack.
			return err
		}
	}

	return nil
}

func (dj *DepJob) tailResID() (string, error) {
	back, err := dj.tail()
	if err != nil {
		return "", errors.Trace(err)
	}

	if t, ok := back.Value.(*Task); ok {
		return t.ResourceID, nil
	}

	if j, ok := back.Value.(*DepJob); ok {
		return j.tailResID()
	}

	return "", errors.Annotatef(errors.ErrInvalidType, "neither *Task nor *DepJob, it's %v", back)
}

func (dj *DepJob) tail() (b *list.Element, err error) {
	if main, exists := dj.paralled[action.MainGroupName]; exists {
		if b = main.Back(); b == nil {
			err = errors.Annotatef(errors.ErrInvalidValue,
				"%s group's tail isn't valid", action.MainGroupName)
		}
		return
	}

	for _, link := range dj.paralled {
		if b := link.Back(); b != nil {
			return b, nil
		}
	}

	return nil, errors.Annotatef(errors.ErrInvalidValue, "*DepJob has not valid tail node")
}

// Job .
type Job struct {
	id    string
	store Store
	orch  orch.Orchestrator

	// The elements are Runner implement, includes DepJob, and Task.
	link *list.List
}

func newJob(store Store, orch orch.Orchestrator) *Job {
	return &Job{
		store: store,
		orch:  orch,
		link:  list.New(),
	}
}

func (j *Job) parse(acts action.Actions) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Conf.ParseDependenciesTimeout.Duration())
	defer cancel()

	for _, act := range acts {
		if err := j.parseAction(ctx, act, j.link); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

func (j *Job) parseAction(ctx context.Context, act action.Action, link *list.List) error {
	if act.IsDependency() {
		return j.parseDep(ctx, act, link)
	}
	return j.parseSingle(act, link)
}

func (j *Job) parseDep(ctx context.Context, ref action.Action, link *list.List) error {
	complex, err := j.store.GetComplex(ctx, ref.Target())
	if err != nil {
		return errors.Trace(err)
	}

	pj := &DepJob{
		Job:      j,
		paralled: map[string]*list.List{},
	}

	for name, acts := range complex.Groups {
		pj.paralled[name] = list.New()
		for i, act := range acts {
			// Merges parameters.
			params := act.GetParams()
			params.Update(ref.GetParams())

			// Update image.
			if i == 0 {
				if img := ref.GetImage(); img != nil {
					act.SetImage(img)
				}
			}

			if err := j.parseAction(ctx, act, pj.paralled[name]); err != nil {
				return errors.Trace(err)
			}
		}
	}

	pj.node = link.PushBack(pj)

	return nil
}

func (j *Job) parseSingle(act action.Action, link *list.List) error {
	if act.IsDependency() {
		return errors.Annotatef(errors.ErrInvalidValue, "it is a dep. action: %s", act.Target())
	}

	t, err := NewTask(act, j.orch)
	if err != nil {
		return errors.Trace(err)
	}

	t.node = link.PushBack(t)

	return nil
}

func (j *Job) run(ctx context.Context) error {
	if err := runLink(ctx, j.link); err != nil {
		return err
	}

	return waitLink(j.link)
}

// Runner .
type Runner interface {
	run(context.Context) error
	wait() error
}

func waitLink(link *list.List) error {
	return foreachLink(link, func(r Runner) error { return r.wait() })
}

func runLink(ctx context.Context, link *list.List) error {
	return foreachLink(link, func(r Runner) error { return r.run(ctx) })
}

func foreachLink(link *list.List, fn func(Runner) error) error {
	for i, cur := 0, link.Front(); i < link.Len(); i, cur = i+1, cur.Next() {
		runner, ok := cur.Value.(Runner)
		if !ok {
			return errors.Annotatef(errors.ErrInvalidType,
				"expected Runner, but %v", cur.Value)
		}

		if err := fn(runner); err != nil {
			return err
		}
	}

	return nil
}

// Gets element's previous running resource ID.
func prevResID(elem *list.Element) (string, error) {
	p, err := prev(elem)
	if err != nil {
		return "", errors.Trace(err)
	}

	if t, ok := p.Value.(*Task); ok {
		return t.ResourceID, nil
	}

	if dj, ok := p.Value.(*DepJob); ok {
		return dj.tailResID()
	}

	return "", errors.Annotatef(errors.ErrInvalidType, "neither *Task nor *DepJob, it's %v", p)
}

// Gets elements's previous node.
func prev(elem *list.Element) (p *list.Element, err error) {
	if p = elem.Prev(); p == nil {
		err = errors.Annotatef(errors.ErrInvalidValue, "has not previous node")
	}
	return
}
