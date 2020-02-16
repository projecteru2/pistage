package executor

import (
	"container/list"
	"context"
	"testing"

	"github.com/projecteru2/aa/action"
	"github.com/projecteru2/aa/executor/mocks"
	"github.com/projecteru2/aa/orch"
	orchmocks "github.com/projecteru2/aa/orch/mocks"
	"github.com/projecteru2/aa/test/assert"
	"github.com/projecteru2/aa/test/mock"
)

func TestJobAddShouldInheritResource(t *testing.T) {
	raw := `
name: Checkout and Build
uses:
  group0:
    - dep: Checkout code
    - run: cd {{ .dir }} && make test

  group1:
    - run: echo hello
      image: alpine
    - run: echo ok
`
	comp, err := action.Parse(raw)
	assert.NilError(t, err)
	assert.NotNil(t, comp)

	simple := mockSimple()

	checkoutComplex := &action.Complex{
		Name: "Checkout code",
		Groups: action.Groups{
			action.MainGroupName: action.Actions{
				action.NewAtom("",
					&action.Image{Name: "alpine"},
					&action.Command{Raw: "git clone {{ .repo }} /tmp"},
					nil,
				),
			},
		},
	}
	checkoutRunners := map[string]*list.List{action.MainGroupName: list.New()}
	task, err := NewTask(checkoutComplex.Groups[action.MainGroupName][0], simple.orch)
	assert.NilError(t, err)
	checkoutRunners[action.MainGroupName].PushBack(task)

	lc, ok := simple.store.(*mocks.Store)
	assert.True(t, ok)
	assert.NotNil(t, lc)
	// Complex so-called "Checkout code"
	lc.On("GetComplex", mock.Anything, mock.Anything).Return(checkoutComplex, nil).Once()

	jg, err := simple.parse(comp)
	assert.NilError(t, err)
	assert.Equal(t, 2, jg.jobs.Len())

	eru, ok := simple.orch.(*orchmocks.Orchestrator)
	assert.True(t, ok)

	ch := func() <-chan orch.Message {
		c := make(chan orch.Message, 1)
		c <- orch.Message{EOF: true}
		close(c)
		return c
	}
	eru.On("Lambda", mock.Anything, mock.Anything).Return("c0", ch(), nil).Once()
	eru.On("Execute", mock.Anything, mock.Anything).Return(ch(), nil).Once()

	g0 := jg.jobs["group0"]
	assert.NilError(t, g0.run(context.Background()))

	eru.On("Lambda", mock.Anything, mock.Anything).Return("c1", ch(), nil).Once()
	eru.On("Execute", mock.Anything, mock.Anything).Return(ch(), nil).Once()

	g1 := jg.jobs["group1"]
	assert.NilError(t, g1.run(context.Background()))
}

func TestJobAddFailedAsHeadActionHasnotImage(t *testing.T) {
	// TODO
}

func TestJobAddFailedAsCycleLinked(t *testing.T) {
	// TODO
}

func TestJobAddSimple(t *testing.T) {
	atom := action.NewAtom("",
		&action.Image{Name: "alpine"},
		&action.Command{Raw: "echo {{ .txt }}"},
		action.Parameters{"txt": action.NewKV("txt", "hello, world")},
	)
	name := "grp"
	acts := action.Actions{atom}

	jg := NewJobGroup()
	assert.NilError(t, jg.add(name, acts))
	assert.Equal(t, 1, jg.jobs.Len())
	assert.Equal(t, "", jg.id)

	job, exists := jg.jobs[name]
	assert.True(t, exists)
	assert.NotNil(t, job)
	assert.Equal(t, 1, job.link.Len())

	elem := job.link.Back()
	assert.NotNil(t, elem)
	assert.NotNil(t, elem.Value)

	task, ok := elem.Value.(*Task)
	assert.True(t, ok)
	assert.Equal(t, "", task.ID)
	assert.Equal(t, StatusPending, task.Status)
	assert.Equal(t, "", task.Stdout)
	assert.Equal(t, "", task.Stderr)
	assert.Equal(t, 0, task.ReturnCode)
	assert.True(t, atom.Equal(task.Action))
	assert.Equal(t, "", task.ResourceID)
}
