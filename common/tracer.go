package common

import (
	"bytes"
	"io"
	"os"
	"sync"
)

// LogTracer traces log output.
type LogTracer struct {
	buffer *bytes.Buffer
	writer io.Writer
	mutex  sync.Mutex
}

// NewLogTracer creates a LogTracer.
func NewLogTracer() *LogTracer {
	buffer := &bytes.Buffer{}
	return &LogTracer{
		buffer: buffer,
		writer: io.MultiWriter(buffer, os.Stdout),
	}
}

// Read implements io.Reader.
func (l *LogTracer) Read(p []byte) (int, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.buffer.Read(p)
}

// Write implements io.Writer.
// Also Write will write data to os.Stdout for output.
// TODO maybe we should use logrus format to write this log line.
func (l *LogTracer) Write(p []byte) (int, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.writer.Write(p)
}
