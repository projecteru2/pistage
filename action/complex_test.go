package action

import (
	"testing"

	"github.com/projecteru2/aa/test/assert"
)

func TestComplexCheck(t *testing.T) {
	cases := []*Complex{
		{Name: ""},
		{Name: "x", Groups: Groups{}},
		{Name: "x", Groups: Groups{"g": Actions{}}},
		{Name: "x", Groups: Groups{"g": Actions{}}},
		{Name: "x", Groups: Groups{"g": Actions{&Atom{Image: &Image{Name: ""}}}}},
		{Name: "x", Groups: Groups{"g": Actions{&Atom{Command: &Command{Raw: ""}}}}},
		{Name: "x", Groups: Groups{}, Params: Parameters{"k": KV{Key: ""}}},
		{Name: "x", Groups: Groups{}, Params: Parameters{"k": Var{Raw: ""}}},
		{Name: "x", Groups: Groups{}, Params: Parameters{"k": Var{Raw: "{{ .x }}", Name: ""}}},
	}

	for _, c := range cases {
		assert.Error(t, c.Check())
	}
}

func TestComplexEqual(t *testing.T) {
	cases := []struct {
		a, b *Complex
	}{
		{
			&Complex{Name: "a"},
			&Complex{Name: "b"},
		},
		{
			&Complex{},
			&Complex{Groups: Groups{}},
		},
		{
			&Complex{},
			&Complex{Params: Parameters{}},
		},
		{
			&Complex{Groups: Groups{"g0": Actions{}}},
			&Complex{Groups: Groups{"g0": Actions{}, "g1": Actions{}}},
		},
		{
			&Complex{Params: Parameters{"k0": KV{}}},
			&Complex{Params: Parameters{"k0": KV{}, "k1": Var{}}},
		},
		{
			&Complex{Groups: Groups{"g0": Actions{}}},
			&Complex{Groups: Groups{"g1": Actions{}}},
		},
		{
			&Complex{Params: Parameters{"k0": KV{}}},
			&Complex{Params: Parameters{"k1": KV{}}},
		},
		{
			&Complex{Groups: Groups{"g0": Actions{&Atom{Dep: "d"}}}},
			&Complex{Groups: Groups{"g0": Actions{&Atom{Dep: "e"}}}},
		},
		{
			&Complex{Groups: Groups{"g0": Actions{&Atom{}}}},
			&Complex{Groups: Groups{"g0": Actions{&Atom{Image: &Image{}}}}},
		},
		{
			&Complex{Groups: Groups{"g0": Actions{&Atom{Image: &Image{Name: "i"}}}}},
			&Complex{Groups: Groups{"g0": Actions{&Atom{Image: &Image{Name: "j"}}}}},
		},
		{
			&Complex{Groups: Groups{"g0": Actions{&Atom{}}}},
			&Complex{Groups: Groups{"g0": Actions{&Atom{Command: &Command{}}}}},
		},
		{
			&Complex{Groups: Groups{"g0": Actions{&Atom{Command: &Command{Raw: "r"}}}}},
			&Complex{Groups: Groups{"g0": Actions{&Atom{Command: &Command{Raw: "s"}}}}},
		},
		{
			&Complex{Groups: Groups{"g0": Actions{&Atom{}}}},
			&Complex{Groups: Groups{"g0": Actions{&Atom{Params: Parameters{}}}}},
		},
		{
			&Complex{Params: Parameters{"k0": KV{Key: "k0", Value: "v0"}}},
			&Complex{Params: Parameters{"k1": KV{Key: "k1", Value: "v1"}}},
		},
		{
			&Complex{Params: Parameters{"k0": KV{Key: "k0", Value: "v0"}}},
			&Complex{Params: Parameters{"k0": KV{Key: "k0", Value: "v1"}}},
		},
		{
			&Complex{Params: Parameters{"k0": KV{Key: "k0", Value: "v0"}}},
			&Complex{Params: Parameters{"k1": KV{Key: "k1", Value: "v0"}}},
		},
	}

	for _, c := range cases {
		assert.False(t, c.a.Equal(c.b))
	}
}
