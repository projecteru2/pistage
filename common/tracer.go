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

// Close implements io.Closer.
// But Close won't close DonCloseWriter.
func (l *LogTracer) Close() error {
	for _, tracer := range l.tracers {
		if _, ok := tracer.(DonCloseWriter); ok {
			continue
		}
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

// DonCloseWriter wraps an io.Writer, to avoid being closed by LogTracer.
type DonCloseWriter struct {
	io.Writer
}

// Close tries to close the io.Writer if it's an io.WriteCloser.
// And by doing this, DonCloseWriter is an io.WriteCloser,
// which can be used as the log tracer for PhistageTask.
func (dcw DonCloseWriter) Close() error {
	if closer, ok := dcw.Writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// ClosableDiscard is a discard that implements io.WriteCloser
// and won't do anything when writting to this object.
var ClosableDiscard io.WriteCloser = discard{}

type discard struct{}

func (discard) Write(p []byte) (int, error) {
	return len(p), nil
}

func (discard) Close() error {
	return nil
}
