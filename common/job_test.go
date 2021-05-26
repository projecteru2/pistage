package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKhoriumStepBuildEnvironmentVariables(t *testing.T) {
	assert := assert.New(t)

	ks := &KhoriumStep{
		Name:        "test",
		Description: "test",
		Inputs: map[string]*KhoriumStepInput{
			"input1": {
				Description: "input1",
				Default:     "default1",
				Required:    false,
			},
			"input2": {
				Description: "input2",
				Required:    true,
			},
		},
		Run: &KhoriumStepRun{
			Main: "",
			Post: "",
		},
	}

	ev1, err := ks.BuildEnvironmentVariables(map[string]string{"input1": "i1", "input2": "i2", "input3": "i3"})
	assert.NoError(err)
	assert.Equal(len(ev1), 2)
	assert.Equal(ev1["KHORIUMSTEP_INPUT_INPUT1"], "i1")
	assert.Equal(ev1["KHORIUMSTEP_INPUT_INPUT2"], "i2")

	ev2, err := ks.BuildEnvironmentVariables(map[string]string{"input2": "i2", "input3": "i3"})
	assert.NoError(err)
	assert.Equal(len(ev2), 2)
	assert.Equal(ev2["KHORIUMSTEP_INPUT_INPUT1"], "default1")
	assert.Equal(ev2["KHORIUMSTEP_INPUT_INPUT2"], "i2")

	_, err = ks.BuildEnvironmentVariables(map[string]string{"input1": "i1", "input3": "i3"})
	assert.Error(err)
}
