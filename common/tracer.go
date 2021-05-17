package common

import (
	"bytes"
	"io"
	"os"
	"sync"
)

type LogTracer struct {
	buffer *bytes.Buffer
	mutex  sync.Mutex
}

func NewLogTracer() *LogTracer {
	return &LogTracer{
		buffer: &bytes.Buffer{},
	}
}

func (l *LogTracer) Read(p []byte) (int, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.buffer.Read(p)
}

func (l *LogTracer) Write(p []byte) (int, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	writer := io.MultiWriter(l.buffer, os.Stdout)
	return writer.Write(p)
}
