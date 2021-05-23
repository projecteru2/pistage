package common

import (
	"bytes"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
)

// LogTracer traces log output.
type LogTracer struct {
	buffer  *bytes.Buffer
	writer  io.Writer
	mutex   sync.Mutex
	tracers []io.Writer
}

// NewLogTracer creates a LogTracer.
func NewLogTracer(id string, tracers ...io.Writer) *LogTracer {
	buffer := &bytes.Buffer{}
	writers := []io.Writer{
		buffer,
		newLogrusTracer(id),
	}
	writers = append(writers, tracers...)

	return &LogTracer{
		buffer:  buffer,
		writer:  io.MultiWriter(writers...),
		tracers: tracers,
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

// Close implements io.Closer
func (l *LogTracer) Close() error {
	for _, tracer := range l.tracers {
		if t, ok := tracer.(io.Closer); ok {
			if err := t.Close(); err != nil {
				return err
			}
		}
	}
	return nil
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

// Write implements io.Writer.
func (l *logrusTracer) Write(p []byte) (int, error) {
	l.entry.Info(string(p))
	return len(p), nil
}
