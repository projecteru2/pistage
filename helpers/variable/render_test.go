package variable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceVariables(t *testing.T) {
	assert := assert.New(t)

	r1 := ReplaceVariables("{{   $env.XXX}} {{$vars.BBB}} {{$xxx.CCC}}")
	assert.Equal(r1, "{{   __pistage_env__.XXX}} {{__pistage_vars__.BBB}} {{$xxx.CCC}}")
}

func TestBuildTemplateContext(t *testing.T) {
	assert := assert.New(t)

	c := BuildTemplateContext(map[string]string{"a1": "v1", "a2": "v2"}, map[string]string{"X": "1"}, map[string]string{"v1": "c1"})
	assert.NotNil(c)
	assert.Equal(len(c), 4)
	assert.Equal(c["a1"].(string), "v1")
	assert.Equal(c["a2"].(string), "v2")
	assert.Equal(c[pistageEnvVarName].(map[string]string)["X"], "1")
	assert.Equal(c[pistageVarsVarName].(map[string]string)["v1"], "c1")
}
