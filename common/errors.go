package common

import (
	"github.com/pkg/errors"
)

var (
	// ErrExecutionError is returned when there's
	// something wrong when executing the shell script.
	// Usually this is due to the non-zero exiting code.
	ErrExecutionError = errors.New("Execution error")
)
