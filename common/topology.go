package common

import (
	"github.com/pkg/errors"
)

// ErrorNotDAG is returned when the input dependencies is not a DAG
// we can't resolve a non-DAG dependency
var ErrorNotDAG = errors.New("Input should be a DAG")

// no priority queue is used
// since the data won't be too large
// besides our destination differs from topology sorting
type topo struct {
	vertices map[string]int
	edges    map[string][]string
	jobCh    chan string
	finishCh chan string
}

func newTopo() *topo {
	return &topo{
		vertices: map[string]int{},
		edges:    map[string][]string{},
	}
}

// addDependencies adds dependencies for name,
// if name has no dependencies, just simply leave dependencies empty.
func (t *topo) addDependencies(name string, dependencies ...string) {
	t.vertices[name] = 0
	for _, dependency := range dependencies {
		if _, ok := t.vertices[dependency]; !ok {
			t.vertices[dependency] = 0
		}
		t.vertices[name]++
		t.edges[dependency] = append(t.edges[dependency], name)
	}
}

func (t *topo) checkCyclic() error {
	_, err := t.graph()
	return err
}

// graph returns the global execution graph for the jobs.
// Note that this is not that high efficiency.
// For example, if A->B->C, A->D->E, this graph shows that C, E can be executed parallelly,
// and B, D parallelly after C and E both finishes.
// But the fact is that as soon as E finishes, D can be started, the same situation is also for B.
// When really need this efficiency, use stream instead.
func (t *topo) graph() ([][]string, error) {
	var (
		r         [][]string
		total     = len(t.vertices)
		collected = 0
	)
	for collected < total {
		vertices := []string{}
		for vname, indegree := range t.vertices {
			if indegree != 0 {
				continue
			}
			vertices = append(vertices, vname)
		}
		// if no vertices is chosen, there's a cycle,
		// which is a wrong input
		if len(vertices) == 0 {
			return nil, ErrorNotDAG
		}

		// brute force
		// just ignore this...
		for _, vname := range vertices {
			tos := t.edges[vname]
			for _, to := range tos {
				t.vertices[to]--
			}
			delete(t.vertices, vname)
			delete(t.edges, vname)
		}

		r = append(r, vertices)
		collected += len(vertices)
	}
	return r, nil
}

const EmptyVertexName = ""

// stream returns a job channel to get jobs from,
// a finish channel indicating that one job is finished,
// and a callback function to tell that all jobs are finished.
// Remember to send a finished job name back to finish channel,
// and call the callback when and only when all jobs are finished.
// This method do not check if the topo is a valid DAG. So do remember to check cyclic before calling this method.
func (t *topo) stream() (<-chan string, chan<- string) {
	length := len(t.vertices)
	t.jobCh = make(chan string, length)

	// finishCh receives the run status of vnames.
	// If finishCh receives a EmptyVertexName, it means at least one of the jobs failed, no need to proceed.
	t.finishCh = make(chan string, length)

	go func() {
		t.iterate()
		for {
			if len(t.vertices) == 0 {
				close(t.jobCh)
				return
			}
			select {
			case vname := <-t.finishCh:
				if vname == EmptyVertexName {
					close(t.jobCh)
					return
				}
				for _, to := range t.edges[vname] {
					t.vertices[to]--
				}
				delete(t.edges, vname)
			}
			t.iterate()
		}
	}()
	// t.finishCh is never closed.
	// Don't worry it will be GCed.
	return t.jobCh, t.finishCh
}

func (t *topo) iterate() {
	vnames := []string{}
	// push all nodes with 0 indegree to the process channel
	for vname, indegree := range t.vertices {
		if indegree == 0 {
			vnames = append(vnames, vname)
		}
	}
	for _, vname := range vnames {
		delete(t.vertices, vname)
		t.jobCh <- vname
	}
}
