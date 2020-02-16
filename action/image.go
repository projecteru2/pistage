package action

import (
	"github.com/projecteru2/aa/errors"
)

// Image indicates an image of an Atom.
type Image struct {
	Name string
}

// NewImage .
func NewImage(name string) *Image {
	return &Image{Name: name}
}

// Check .
func (i *Image) Check() error {
	if len(i.Name) < 1 {
		return errors.Annotatef(errors.ErrInvalidValue, "name is empty")
	}
	return nil
}

// Equal .
func (i *Image) Equal(other *Image) bool {
	return other != nil && i.Name == other.Name
}
