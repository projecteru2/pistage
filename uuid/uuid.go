package uuid

import (
	guid "github.com/google/uuid"

	"github.com/projecteru2/aa/errors"
)

// New .
func New() (string, error) {
	u, err := guid.NewUUID()
	if err != nil {
		return "", errors.Trace(err)
	}
	return u.String(), nil
}

// Check .
func Check(s string) (err error) {
	_, err = guid.Parse(s)
	return
}
