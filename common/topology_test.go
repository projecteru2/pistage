package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopology(t *testing.T) {
	assert := assert.New(t)

	tp1 := newTopo()
	tp1.addEdge("A", "C")
	tp1.addEdge("B", "C")
	tp1.addEdge("C", "D")
	tp1.addEdge("C", "E")

	r1, err := tp1.sort()
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

	r2, err := tp2.sort()
	assert.NoError(err)
	assert.Equal(len(r2), 4)
	assert.Equal(len(r2[0]), 4)
	assert.Equal(len(r2[1]), 3)
	assert.Equal(len(r2[2]), 1)
	assert.Equal(r2[2][0], "H")
	assert.Equal(len(r2[3]), 1)
	assert.Equal(r2[3][0], "E")
}
