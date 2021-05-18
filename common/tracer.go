package common

import (
	"bytes"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
)

// LogTracer traces log output.
type LogTracer struct {
	buffer *bytes.Buffer
	writer io.Writer
	mutex  sync.Mutex
}

// NewLogTracer creates a LogTracer.
func NewLogTracer(id string) *LogTracer {
	buffer := &bytes.Buffer{}
	return &LogTracer{
		buffer: buffer,
		writer: io.MultiWriter(buffer, newLogrusTracer(id)),
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
func (l *LogTracer) Write(p []byte) (int, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.writer.Write(p)
}

type logrusTracer struct {
	entry *logrus.Entry
}

func newLogrusTracer(id string) *logrusTracer {
	return &logrusTracer{
		entry: logrus.WithFields(logrus.Fields{
			"JobRunID": id,
		}),
	}
}

func (l *logrusTracer) Write(p []byte) (int, error) {
	l.entry.Info(string(p))
	return len(p), nil
}
