package action

import (
	"gopkg.in/yaml.v3"

	"github.com/projecteru2/aa/errors"
)

const (
	// KeyName .
	KeyName = "name"
	// KeyParams .
	KeyParams = "with"
	// KeyDep .
	KeyDep = "dep"
	// KeyRun .
	KeyRun = "run"
	// KeyImage .
	KeyImage = "image"
	// KeyGroups .
	KeyGroups = "uses"
	// KeyAsync .
	KeyAsync = "async"
)

// Parse .
func Parse(s string) (*Complex, error) {
	dict := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(s), &dict); err != nil {
		return nil, errors.Trace(err)
	}

	comp := NewComplex()
	return comp, comp.Parse(dict)
}
