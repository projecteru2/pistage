package action

import (
	"github.com/projecteru2/aa/errors"
)

func mustString(dict map[string]interface{}, key string) (value string, exists bool, err error) {
	var raw interface{}
	if raw, exists = dict[key]; !exists {
		return
	}

	var ok bool
	if value, ok = raw.(string); !ok {
		err = errors.Annotatef(errors.ErrInvalidType, "%s is not string", key)
	}

	return
}
