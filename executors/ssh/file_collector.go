package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SSHFileCollector collects or sends files from or to the working environment.
// Note: paths for files should be relative to working dir,
// working dir should be passed as identifier parameter.
type SSHFileCollector struct {
	mutex     sync.Mutex
	filesLock sync.Mutex
	files     map[string][]byte
	client    *ssh.Client
}

func NewSSHFileCollector(client *ssh.Client) *SSHFileCollector {
	return &SSHFileCollector{
		files:  map[string][]byte{},
		client: client,
	}
}

func (s *SSHFileCollector) SetFiles(files map[string][]byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.files = files
}

// Collect collects files.
// For an SSHFileCollector, identifier represents the current working dir,
// identifier will be used to do a file path join with files.
func (s *SSHFileCollector) Collect(ctx context.Context, identifier string, files []string) error {
	if len(files) == 0 {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	sc, err := sftp.NewClient(s.client)
	if err != nil {
		return err
	}
	defer sc.Close()

	for _, file := range files {
		path := filepath.Join(identifier, file)
		// We don't allow to collect files outside identifier.
		if !strings.HasPrefix(path, identifier) {
			continue
		}

		f, err := sc.Open(path)
		if err != nil {
			return err
		}

		content, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		s.files[file] = content
	}
	return nil
}

// CopyTo copies files to the destination.
// For an SSHFileCollector, identifier represents the target working dir,
// identifier will be used to do a file path join with files.
func (s *SSHFileCollector) CopyTo(ctx context.Context, identifier string, files []string) error {
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

	sc, err := sftp.NewClient(s.client)
	if err != nil {
		return err
	}
	defer sc.Close()

	if err := s.createEssentialDirs(ctx, identifier, data); err != nil {
		return err
	}
	for file, content := range data {
		path := filepath.Join(identifier, file)
		// We don't allow to copy files to outside of identifier.
		if !strings.HasPrefix(path, identifier) {
			continue
		}

		local := bytes.NewBuffer(content)
		remote, err := sc.Create(path)
		if err != nil {
			return err
		}
		io.Copy(remote, local)
		remote.Close()
	}
	return nil
}

// createEssentialDirs creates essential dirs for files.
func (s *SSHFileCollector) createEssentialDirs(ctx context.Context, identifier string, files map[string][]byte) error {
	dirs := map[string]struct{}{}
	for path := range files {
		dirs[filepath.Dir(filepath.Join(identifier, path))] = struct{}{}
	}

	dirnames := []string{}
	for path := range dirs {
		dirnames = append(dirnames, path)
	}
	paths := strings.Join(dirnames, " ")
	cmd := fmt.Sprintf("mkdir -p %s", paths)
	return executeCommand(s.client, cmd, identifier, nil, io.Discard)
}

// Files returns all file names including path this collector holds.
func (s *SSHFileCollector) Files() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var files []string
	for file := range s.files {
		files = append(files, file)
	}
	return files
}
