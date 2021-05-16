package common

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopologyGraph1(t *testing.T) {
	assert := assert.New(t)

	tp1 := newTopo()
	tp1.addEdge("A", "C")
	tp1.addEdge("B", "C")
	tp1.addEdge("C", "D")
	tp1.addEdge("C", "E")

	r1, err := tp1.graph()
	assert.NoError(err)
	assert.Equal(len(r1), 3)
	assert.Equal(len(r1[0]), 2)
	assert.Equal(len(r1[1]), 1)
	assert.Equal(r1[1][0], "C")
	assert.Equal(len(r1[2]), 2)

	tp2 := newTopo()
	tp2.addEdge("A", "D")
	tp2.addEdge("B", "D")
	tp2.addEdge("C", "D")
	tp2.addEdge("D", "E")
	tp2.addEdge("C", "E")
	tp2.addEdge("F", "E")
	tp2.addEdge("H", "E")
	tp2.addEdge("G", "H")
	tp2.addEdge("M", "F")
	tp2.addEdge("M", "G")

	r2, err := tp2.graph()
	assert.NoError(err)
	assert.Equal(len(r2), 4)
	assert.Equal(len(r2[0]), 4)
	assert.Equal(len(r2[1]), 3)
	assert.Equal(len(r2[2]), 1)
	assert.Equal(r2[2][0], "H")
	assert.Equal(len(r2[3]), 1)
	assert.Equal(r2[3][0], "E")
}

func TestTopologyGraph2(t *testing.T) {
	assert := assert.New(t)

	tp := newTopo()
	tp.addEdge("A", "")
	tp.addEdge("B", "")
	tp.addEdge("C", "")

	tp.addEdge("D", "F")
	tp.addEdge("E", "F")

	r, err := tp.graph()
	assert.NoError(err)
	assert.Equal(len(r), 2)
}

func TestTopologyStream1(t *testing.T) {
	assert := assert.New(t)

	tp := newTopo()
	tp.addEdge("A", "C")
	tp.addEdge("B", "C")
	tp.addEdge("C", "D")
	tp.addEdge("D", "I")
	tp.addEdge("B", "E")
	tp.addEdge("G", "E")
	tp.addEdge("H", "E")
	tp.addEdge("E", "D")

	jobs, finished, finish := tp.stream()
	result := make(chan string)
	r := []string{}
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for e := range result {
			r = append(r, e)
		}
	}()

	for job := range jobs {
		result <- job
		finished <- job
	}
	finish()
	close(result)
	wg.Wait()

	assert.Equal(len(r), 8)
	assert.Equal(r[len(r)-1], "I")
}

func TestTopologyStream2(t *testing.T) {
	assert := assert.New(t)

	tp := newTopo()
	tp.addEdge("A", "C")
	tp.addEdge("B", "C")
	tp.addEdge("D", "C")
	tp.addEdge("E", "C")
	tp.addEdge("F", "C")
	tp.addEdge("G", "C")
	tp.addEdge("H", "C")

	jobs, finished, finish := tp.stream()
	result := make(chan string)
	r := []string{}
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for e := range result {
			r = append(r, e)
		}
	}()

	for job := range jobs {
		result <- job
		finished <- job
	}
	finish()
	close(result)
	wg.Wait()

	assert.Equal(len(r), 8)
	assert.Equal(r[len(r)-1], "C")
}

func TestTopologyStream3(t *testing.T) {
	assert := assert.New(t)

	tp := newTopo()
	tp.addEdge("A", "")
	tp.addEdge("B", "")
	tp.addEdge("C", "")

	jobs, finished, finish := tp.stream()
	result := make(chan string)
	r := []string{}
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for e := range result {
			r = append(r, e)
		}
	}()

	for job := range jobs {
		result <- job
		finished <- job
	}
	finish()
	close(result)
	wg.Wait()

	assert.Equal(len(r), 3)
}
