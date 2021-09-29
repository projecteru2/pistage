package shell

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/projecteru2/pistage/helpers"
)

// ShellFileCollector collects or sends files from or to working environment.
// Node: all paths of files should be relative to working dir,
// working dir should be passed as identifier parameter.
// This is very much like SSHFileCollector.
type ShellFileCollector struct {
	mutex     sync.Mutex
	filesLock sync.Mutex
	files     map[string][]byte
}

func NewShellFileCollector() *ShellFileCollector {
	return &ShellFileCollector{
		files: map[string][]byte{},
	}
}

func (s *ShellFileCollector) SetFiles(files map[string][]byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.files = files
}

// Collect collects files.
// For an ShellFileCollector, identifier represents the current working dir,
// identifier will be used to do a file path join with files.
func (s *ShellFileCollector) Collect(ctx context.Context, identifier string, files []string) error {
	if len(files) == 0 {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, file := range files {
		path := filepath.Join(identifier, file)
		// We don't allow to collect files outside identifier.
		if !strings.HasPrefix(path, identifier) {
			continue
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		s.files[file] = content
	}
	return nil
}

// CopyTo copies files to the destination.
// For an ShellFileCollector, identifier represents the target working dir,
// identifier will be used to do a file path join with files.
func (s *ShellFileCollector) CopyTo(ctx context.Context, identifier string, files []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data := map[string][]byte{}
	if len(files) == 0 {
		data = s.files
	} else {
		for _, filename := range files {
			content, ok := s.files[filename]
			if !ok {
				continue
			}
			data[filename] = content
		}
	}

	if len(data) == 0 {
		return nil
	}

	for file, content := range data {
		// create the essential directory.
		path := filepath.Join(identifier, file)
		// We don't allow to copy files to outside of identifier.
		if !strings.HasPrefix(path, identifier) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		if err := helpers.WriteIfNotExist(path, content); err != nil {
			return err
		}
	}
	return nil
}

// Files returns all file names including path this collector holds.
func (s *ShellFileCollector) Files() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var files []string
	for file := range s.files {
		files = append(files, file)
	}
	return files
}
