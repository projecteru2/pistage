package common

import (
	"github.com/juju/errors"
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
}

func newTopo() *topo {
	return &topo{
		vertices: map[string]int{},
		edges:    map[string][]string{},
	}
}

func (t *topo) addEdge(from, to string) {
	if _, ok := t.vertices[from]; !ok {
		t.vertices[from] = 0
	}
	t.vertices[to]++
	t.edges[from] = append(t.edges[from], to)
}

func (t *topo) sort() ([][]string, error) {
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
