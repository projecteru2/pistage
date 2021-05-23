package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderCommand(t *testing.T) {
	assert := assert.New(t)

	o1, err := RenderCommand("{{a}} {{b}}", map[string]string{"a": "testa", "b": "testb", "c": "testc"}, nil, nil)
	assert.NoError(err)
	assert.Equal(o1, "testa testb")

	o2, err := RenderCommand("{{a}} {{b}}", map[string]string{"a": "testa"}, nil, nil)
	assert.NoError(err)
	assert.Equal(o2, "testa ")

	o3, err := RenderCommand("{{a}} {{vars.b}}", map[string]string{"a": "testa"}, nil, map[string]string{"b": "testb"})
	assert.NoError(err)
	assert.Equal(o3, "testa testb")

	o4, err := RenderCommand("{{a}} {{vars.b}}", nil, nil, map[string]string{"b": "testb"})
	assert.NoError(err)
	assert.Equal(o4, " testb")

	o5, err := RenderCommand("{{a}} {{vars.b}}", map[string]string{"a": "testa"}, nil, nil)
	assert.NoError(err)
	assert.Equal(o5, "testa ")

	o6, err := RenderCommand("{{a}} {{env.TEST}} {{vars.b}}", map[string]string{"a": "testa"}, map[string]string{"TEST": "notest"}, nil)
	assert.NoError(err)
	assert.Equal(o6, "testa notest ")

	o7, err := RenderCommand("{{a}} {{env.TEST}} {{vars.b | default_if_none: 'xxx'}}", map[string]string{"a": "testa"}, map[string]string{"TEST": "notest"}, nil)
	assert.NoError(err)
	assert.Equal(o7, "testa notest xxx")
}
