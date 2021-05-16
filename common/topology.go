package common

import (
	"github.com/pkg/errors"
)

// ErrorNotDAG is returned when the input dependencies is not a DAG
// we can't resolve a non-DAG dependency
var ErrorNotDAG = errors.New("Input should be a DAG")

// no priority queue is used
// since the data won't be too large
// beside our destination differs from topology sorting
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

var emptyVertex = ""

func (t *topo) addEdge(from, to string) {
	if _, ok := t.vertices[from]; !ok {
		t.vertices[from] = 0
	}
	if to != emptyVertex {
		t.vertices[to]++
		t.edges[from] = append(t.edges[from], to)
	}
}

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
		// if no vertices is chosen, there's a circle,
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

func (t *topo) stream() (<-chan string, chan<- string, func()) {
	length := len(t.vertices)
	t.jobCh = make(chan string, length)
	t.finishCh = make(chan string, length)

	go t.checkJobs()
	return t.jobCh, t.finishCh, func() {
		close(t.finishCh)
	}
}

func (t *topo) checkJobs() {
	defer close(t.jobCh)
	var (
		total      = len(t.vertices)
		collected  = 0
		vertices   = []string{}
		unfinished = 0
	)

	for vname, indegree := range t.vertices {
		if indegree != 0 {
			continue
		}
		vertices = append(vertices, vname)
	}
	// if no vertices is chosen, there's a circle,
	// which is a wrong input
	if len(vertices) == 0 {
		return
	}

	collected += len(vertices)
	unfinished += len(vertices)
	for _, vname := range vertices {
		t.jobCh <- vname
	}

	for collected < total {
		for vname := range t.finishCh {
			unfinished--

			tos := t.edges[vname]
			for _, to := range tos {
				t.vertices[to]--
			}
			delete(t.vertices, vname)
			delete(t.edges, vname)

			vertices = []string{}
			for _, vname := range tos {
				if t.vertices[vname] != 0 {
					continue
				}
				vertices = append(vertices, vname)
			}

			if len(vertices) == 0 {
				if unfinished == 0 {
					return
				}
				continue
			}

			collected += len(vertices)
			unfinished += len(vertices)
			for _, vname := range vertices {
				t.jobCh <- vname
			}
		}
	}
}
