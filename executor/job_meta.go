package executor

import (
	"container/list"
	"context"
	"path/filepath"

	"github.com/projecteru2/aa/action"
	"github.com/projecteru2/aa/config"
	"github.com/projecteru2/aa/errors"
	"github.com/projecteru2/aa/store/meta"
)

type jobMeta struct {
	ID     string                 `json:"id"`
	Params action.Parameters      `json:"params"`
	Jobs   map[string]interface{} `json:"jobs"`

	*meta.Ver `json:"-"`
}

func newJobMeta() *jobMeta {
	return &jobMeta{
		Params: action.Parameters{},
		Jobs:   map[string]interface{}{},
		Ver:    meta.NewVer(),
	}
}

func (m *jobMeta) parse(jg *JobGroup) error {
	m.ID = jg.id
	m.Params = jg.params

	return m.parseJobs(jg.jobs)
}

func (m *jobMeta) parseJobs(jobs Jobs) (err error) {
	dic := map[string]interface{}{}

	for nm, j := range jobs {
		if dic[nm], err = m.parseLink(j.link); err != nil {
			return
		}
	}

	m.Jobs = dic

	return
}

func (m *jobMeta) parseDepJob(dj *DepJob) (items map[string]interface{}, err error) {
	items = map[string]interface{}{}

	for nm, link := range dj.paralled {
		if items[nm], err = m.parseLink(link); err != nil {
			return
		}
	}

	return
}

func (m *jobMeta) parseLink(link *list.List) (interface{}, error) {
	items := []interface{}{}

	for i, cur := 0, link.Front(); i < link.Len(); i, cur = i+1, cur.Next() {
		t, err := m.parseRunner(cur.Value)
		if err != nil {
			return nil, errors.Trace(err)
		}

		items = append(items, t)
	}

	return items, nil
}

func (m *jobMeta) parseRunner(elem interface{}) (interface{}, error) {
	// Actually, the elem is list.List.Element which is a Runner implement,
	// but we are deliberately considering to skip to convert Runner.
	if dj, ok := elem.(*DepJob); ok {
		return m.parseDepJob(dj)
	}

	if task, ok := elem.(*Task); ok {
		return m.parseTask(task)
	}

	return nil, errors.Trace(errors.ErrInvalidType)
}

func (m *jobMeta) parseTask(task *Task) (interface{}, error) {
	return task, nil
}

func (m *jobMeta) save(ctx context.Context) error {
	return meta.Save(meta.Resources{m})
}

// MetaKey .
func (m *jobMeta) MetaKey() string {
	return filepath.Join(config.Conf.EtcdPrefix, "job", m.ID)
}
