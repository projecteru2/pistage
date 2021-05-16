package common

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopologyGraph1(t *testing.T) {
	assert := assert.New(t)

	tp1 := newTopo()
	tp1.addDependencies("C", "A", "B")
	tp1.addDependencies("D", "C")
	tp1.addDependencies("E", "C")

	r1, err := tp1.graph()
	assert.NoError(err)
	assert.Equal(len(r1), 3)
	assert.Equal(len(r1[0]), 2)
	assert.Equal(len(r1[1]), 1)
	assert.Equal(r1[1][0], "C")
	assert.Equal(len(r1[2]), 2)

	tp2 := newTopo()
	tp2.addDependencies("D", "A", "B", "C")
	tp2.addDependencies("E", "D", "C", "F", "H")
	tp2.addDependencies("H", "G")
	tp2.addDependencies("F", "M")
	tp2.addDependencies("G", "M")

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
	tp.addDependencies("A")
	tp.addDependencies("B")
	tp.addDependencies("C")

	tp.addDependencies("F", "D", "E")

	r, err := tp.graph()
	assert.NoError(err)
	assert.Equal(len(r), 2)
}

func TestTopologyStream1(t *testing.T) {
	assert := assert.New(t)

	tp := newTopo()
	tp.addDependencies("C", "A", "B")
	tp.addDependencies("D", "C", "E")
	tp.addDependencies("I", "D")
	tp.addDependencies("E", "B", "G", "H")

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
	tp.addDependencies("C", "A", "B", "D", "E", "F", "G", "H")

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
	tp.addDependencies("A")
	tp.addDependencies("B")
	tp.addDependencies("C")

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