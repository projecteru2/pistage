package action

import (
	"testing"

	"github.com/projecteru2/aa/test/assert"
)

func TestParseAtomFailedAsCycleRef(t *testing.T) {
	// TODO
}

func TestParseAtom(t *testing.T) {
	raw := `
name: echo
image: alpine
run: echo {{ .ENV_TXT }}
with:
  ENV_TXT: hello, world
`

	params := Parameters{
		"ENV_TXT": KV{
			Key:   "ENV_TXT",
			Value: "hello, world",
		},
	}

	exp := &Complex{
		Name: "echo",
		Groups: Groups{
			"main": Actions{
				&Atom{
					ImageName:  "alpine",
					Image:      &Image{Name: "alpine"},
					RawCommand: "echo {{ .ENV_TXT }}",
					Command:    &Command{Raw: "echo {{ .ENV_TXT }}"},
					Params:     params,
				},
			},
		},
		Params: params,
	}

	act, err := Parse(raw)
	assert.NilError(t, err)
	assert.NotNil(t, act)

	assert.True(t, exp.Equal(act))
}

func TestParseComplex(t *testing.T) {
	raw := `
name: Checkout and Build
uses:
  group0:
    - dep: Checkout code
      with:
        image: alpine
        repo: "{{ .repo }}"
        dir: "{{ .dir }}"

    - run: cd {{ .dir }} && make test
      image: alpine
      async: true
      on:
        failed: echo Oops
        retry: 1
        reentry: make clean

    - dep: Build binary
      with:
        workdir: /tmp/aa

  group1:
    - run: echo {{ .txt }}
      image: alpine
      with:
        txt: finished

with:
  dir: /tmp/aa
`

	exp := &Complex{
		Name: "Checkout and Build",
		Groups: Groups{
			"group0": Actions{
				&Atom{
					Dep: "Checkout code",
					Params: Parameters{
						"image": KV{Key: "image", Value: "alpine"},
						"repo":  Var{Name: "repo", Raw: "{{ .repo }}"},
						"dir":   Var{Name: "dir", Raw: "{{ .dir }}"},
					},
				},
				&Atom{
					ImageName:  "alpine",
					Image:      &Image{Name: "alpine"},
					RawCommand: "cd {{ .dir }} && make test",
					Command:    &Command{Raw: "cd {{ .dir }} && make test"},
					Async:      true,
				},
				&Atom{
					Dep: "Build binary",
					Params: Parameters{
						"workdir": KV{Key: "workdir", Value: "/tmp/aa"},
					},
				},
			},
			"group1": Actions{
				&Atom{
					ImageName:  "alpine",
					Image:      &Image{Name: "alpine"},
					RawCommand: "echo {{ .txt }}",
					Command:    &Command{Raw: "echo {{ .txt }}"},
					Params: Parameters{
						"txt": KV{Key: "txt", Value: "finished"},
					},
				},
			},
		},
		Params: Parameters{
			"dir": KV{Key: "dir", Value: "/tmp/aa"},
		},
	}

	act, err := Parse(raw)
	assert.NilError(t, err)
	assert.NotNil(t, act)
	assert.True(t, exp.Equal(act))
}

func TestParseInvalidYaml(t *testing.T) {
	_, err := Parse("#")
	assert.Error(t, err)
}

func TestParseComplexFailedAsInvalidName(t *testing.T) {
	_, err := Parse("namex: namex")
	assert.Error(t, err)

	_, err = Parse("name: 0")
	assert.Error(t, err)
}
