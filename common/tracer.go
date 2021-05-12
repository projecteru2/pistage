package common

import (
	"io"
)

type LogTracer struct {
	Writer io.Writer
}
