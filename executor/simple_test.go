package executor

import (
	"container/list"
	"context"
	"testing"

	"github.com/projecteru2/aa/action"
	"github.com/projecteru2/aa/executor/mocks"
	"github.com/projecteru2/aa/orch"
	orchmocks "github.com/projecteru2/aa/orch/mocks"
	storemocks "github.com/projecteru2/aa/store/mocks"
	"github.com/projecteru2/aa/test/assert"
	"github.com/projecteru2/aa/test/mock"
)

func TestSimpleStart(t *testing.T) {
	ch := func() <-chan orch.Message {
		c := make(chan orch.Message, 1)
		c <- orch.Message{EOF: true}
		close(c)
		return c
	}()

	simple := mockSimple()
	simple.orch.(*orchmocks.Orchestrator).On("Lambda", mock.Anything, mock.Anything).Return("", ch, nil).Once()

	store, cancel := storemocks.Mocks()
	defer cancel()
	store.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	raw := `
name: echo
image: alpine
run: echo
`
	complex, err := action.Parse(raw)
	assert.NilError(t, err)
	assert.NotNil(t, complex)

	id, err := simple.SyncStart(context.Background(), complex)
	assert.NilError(t, err)
	assert.True(t, len(id) > 1)
}

func TestSimpleParseJobGroup(t *testing.T) {
	complex := createComplex(t)
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

	buildComplex := &action.Complex{
		Name: "Build binary",
		Groups: action.Groups{
			action.MainGroupName: action.Actions{
				action.NewAtom("",
					&action.Image{Name: "alpine"},
					&action.Command{Raw: "cd /tmp && make && make build"},
					nil,
				),
			},
		},
	}
	buildRunners := map[string]*list.List{action.MainGroupName: list.New()}
	task, err = NewTask(buildComplex.Groups[action.MainGroupName][0], simple.orch)
	assert.NilError(t, err)
	buildRunners[action.MainGroupName].PushBack(task)

	lc, ok := simple.store.(*mocks.Store)
	assert.True(t, ok)
	assert.NotNil(t, lc)
	// Complex so-called "Checkout code"
	lc.On("GetComplex", mock.Anything, mock.Anything).Return(checkoutComplex, nil).Once()
	// Complex so-called "Build binary"
	lc.On("GetComplex", mock.Anything, mock.Anything).Return(buildComplex, nil).Once()

	jg, err := simple.parse(complex)
	assert.NilError(t, err)
	assert.Equal(t, 2, jg.jobs.Len())

	// group0 job
	exp := list.New()
	exp.PushBack(&DepJob{paralled: checkoutRunners})
	task, err = NewTask(complex.Groups["group0"][1], nil)
	assert.NilError(t, err)
	exp.PushBack(task)
	exp.PushBack(&DepJob{paralled: buildRunners})
	assertEqualRunners(t, exp, jg.jobs["group0"].link)
	// Overrides the original image
	dj, ok := jg.jobs["group0"].link.Front().Value.(*DepJob)
	assert.True(t, ok)
	task, ok = dj.paralled[action.MainGroupName].Front().Value.(*Task)
	assert.True(t, ok)
	img := task.Action.GetImage()
	assert.NotNil(t, img)
	assert.Equal(t, "brava", img.Name)

	// group1 job
	g1job, exists := jg.jobs["group1"]
	assert.True(t, exists)
	assert.NotNil(t, g1job)
	assert.Equal(t, 1, g1job.link.Len())
	expAtom, ok := complex.Groups["group1"][0].(*action.Atom)
	assertPendingTask(t, expAtom, g1job.link.Back())
	// There're no more elements.
	assert.True(t, ok)
	assert.NotNil(t, expAtom)

	// Asserts job meta.
	md := newJobMeta()
	assert.NilError(t, md.parse(jg))
}

func assertEqualRunners(t *testing.T, exp, real *list.List) {
	assert.Equal(t, exp.Len(), real.Len())

	expElem := exp.Front()
	realElem := real.Front()

	for {
		if expElem == nil {
			return
		}

		assert.NotNil(t, realElem)
		assertEqualRunner(t, expElem, realElem)

		expElem = expElem.Next()
		realElem = realElem.Next()
	}
}

func assertEqualRunner(t *testing.T, exp, real *list.Element) {
	if expTask, ok := exp.Value.(*Task); ok {
		assertEqualTask(t, expTask, real.Value.(*Task))
		return
	}

	assertEqualDJ(t, exp.Value.(*DepJob), real.Value.(*DepJob))
}

func assertEqualTask(t *testing.T, exp, real *Task) {
	assert.Equal(t, exp.ID, real.ID)
	assert.Equal(t, exp.Status, real.Status)
	assert.Equal(t, exp.Stdout, real.Stdout)
	assert.Equal(t, exp.Stderr, real.Stderr)
	assert.Equal(t, exp.ReturnCode, real.ReturnCode)
	assert.Equal(t, exp.ResourceID, real.ResourceID)
	assert.True(t, exp.Action.Equal(real.Action))
}

func assertEqualDJ(t *testing.T, exp, real *DepJob) {
	assert.Equal(t, len(exp.paralled), len(real.paralled))

	for nm, expLink := range exp.paralled {
		realLink, exists := real.paralled[nm]
		assert.True(t, exists)
		assertEqualRunners(t, expLink, realLink)
	}
}

func TestSimpleParseNestedDependencies(t *testing.T) {
	raw := `
name: nested dependencies
uses:
  group0:
    - dep: build
`
	complex, err := action.Parse(raw)
	assert.NilError(t, err)
	assert.NotNil(t, complex)

	simple := mockSimple()

	aptComplex := &action.Complex{
		Name: "apt-install",
		Groups: action.Groups{
			action.MainGroupName: action.Actions{
				action.NewAtom("",
					&action.Image{Name: "alpine"},
					&action.Command{Raw: "apt install {{ .pkg }}"},
					nil,
				),
			},
		},
	}
	aptDJ := &DepJob{
		paralled: map[string]*list.List{action.MainGroupName: list.New()},
	}
	task, err := NewTask(aptComplex.Groups[action.MainGroupName][0], simple.orch)
	assert.NilError(t, err)
	aptDJ.paralled[action.MainGroupName].PushBack(task)

	checkoutComplex := &action.Complex{
		Name: "checkout",
		Groups: action.Groups{
			action.MainGroupName: action.Actions{
				// Depends on "apt-install"
				action.NewAtom("apt-install", nil, nil, nil),
				action.NewAtom("",
					&action.Image{Name: "alpine"},
					&action.Command{Raw: "git clone {{ .repo }}"},
					nil,
				),
			},
		},
	}
	checkoutDJ := &DepJob{
		paralled: map[string]*list.List{action.MainGroupName: list.New()},
	}
	checkoutDJ.paralled[action.MainGroupName].PushBack(aptDJ)
	task, err = NewTask(checkoutComplex.Groups[action.MainGroupName][1], simple.orch)
	assert.NilError(t, err)
	checkoutDJ.paralled[action.MainGroupName].PushBack(task)

	buildComplex := &action.Complex{
		Name: "build",
		Groups: action.Groups{
			action.MainGroupName: action.Actions{
				// Depends on "checkout"
				action.NewAtom("checkout", nil, nil, nil),
				action.NewAtom("",
					&action.Image{Name: "alpine"},
					&action.Command{Raw: "make && make build"},
					nil,
				),
			},
		},
	}
	buildDJ := &DepJob{
		paralled: map[string]*list.List{action.MainGroupName: list.New()},
	}
	buildDJ.paralled[action.MainGroupName].PushBack(checkoutDJ)
	task, err = NewTask(buildComplex.Groups[action.MainGroupName][1], simple.orch)
	assert.NilError(t, err)
	buildDJ.paralled[action.MainGroupName].PushBack(task)

	lc, ok := simple.store.(*mocks.Store)
	assert.True(t, ok)
	assert.NotNil(t, lc)
	// Complex so-called "build"
	lc.On("GetComplex", mock.Anything, mock.Anything).Return(buildComplex, nil).Once()
	// Complex so-called "checkout" which is depended on "build"
	lc.On("GetComplex", mock.Anything, mock.Anything).Return(checkoutComplex, nil).Once()
	// Complex so-called "apt-install" which is depended on "checkout"
	lc.On("GetComplex", mock.Anything, mock.Anything).Return(aptComplex, nil).Once()

	jg, err := simple.parse(complex)
	assert.NilError(t, err)
	assert.Equal(t, 1, jg.jobs.Len())

	exp := list.New()
	exp.PushBack(buildDJ)
	assertEqualRunners(t, exp, jg.jobs["group0"].link)
	logJob(t, jg.jobs["group0"].link)
}

func logJob(t *testing.T, link *list.List) {
	elem := link.Front()

	for i := 0; ; i++ {
		if elem == nil {
			return
		}

		assert.NotNil(t, elem.Value)

		if dj, ok := elem.Value.(*DepJob); ok {
			for _, inner := range dj.paralled {
				logJob(t, inner)
			}
			elem = elem.Next()
			continue
		}

		task, ok := elem.Value.(*Task)
		assert.True(t, ok)
		assert.NotNil(t, task)
		t.Logf("Run [%s] on [%s]\n", task.Action.GetCommand().Raw, task.Action.GetImage().Name)

		elem = elem.Next()
	}
}

func TestSimpleParseDuplicatedDependencies(t *testing.T) {
}

func assertPendingTask(t *testing.T, exp *action.Atom, elem *list.Element) {
	task, ok := elem.Value.(*Task)
	assert.True(t, ok)
	assert.Equal(t, "", task.ID)
	assert.Equal(t, StatusPending, task.Status)
	assert.Equal(t, "", task.Stdout)
	assert.Equal(t, "", task.Stderr)
	assert.Equal(t, 0, task.ReturnCode)
	assert.Equal(t, "", task.ResourceID)
	assert.True(t, exp.Equal(task.Action))
}

func createComplex(t *testing.T) *action.Complex {
	raw := `
name: Checkout and Build
uses:
  group0:
    - dep: Checkout code
      image: brava
      with:
        repo: "{{ .repo }}"
        dir: "{{ .dir }}"

    - run: cd {{ .dir }} && make test
      image: alpine
      on:
        failed: echo Oops
        retry: 1
        reentry: make clean

    - dep: Build binary
      with:
        workdir: /tmp/aa

  group1:
    - run: echo {{ .txt }}
      image: alpine
      with:
        txt: finished

with:
  dir: /tmp/aa
`

	comp, err := action.Parse(raw)
	assert.NilError(t, err)
	assert.NotNil(t, comp)

	return comp
}

func mockSimple() *Simple {
	return &Simple{
		orch:  &orchmocks.Orchestrator{},
		store: &mocks.Store{},
	}
}
