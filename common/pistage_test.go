package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateHash(t *testing.T) {
	a := assert.New(t)
	p1 := &Pistage{
		WorkflowType:       "t",
		WorkflowIdentifier: "i",
		Jobs: map[string]*Job{
			"job1": {},
			"job2": {},
			"job3": {},
		},
	}
	p2 := &Pistage{
		WorkflowType:       "t",
		WorkflowIdentifier: "i",
		Jobs: map[string]*Job{
			"job3": {},
			"job2": {},
			"job1": {},
		},
	}

	a.NoError(p1.GenerateHash())
	a.NoError(p2.GenerateHash())
	a.Equal(p1.ContentHash, p2.ContentHash)
}
