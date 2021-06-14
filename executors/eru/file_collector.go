package eru

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/projecteru2/phistage/common"
)

var ErrorCopyToContainer = errors.New("Error copy to container")

// EruFileCollector collects or sends files from or to workload.
// Note: the paths of files are absolute.
// TODO: use relative paths, too.
type EruFileCollector struct {
	mutex     sync.Mutex
	filesLock sync.Mutex
	eru       corepb.CoreRPCClient
	files     map[string][]byte
}

func NewEruFileCollector(eru corepb.CoreRPCClient) *EruFileCollector {
	return &EruFileCollector{
		eru:   eru,
		files: map[string][]byte{},
	}
}

func (e *EruFileCollector) SetFiles(files map[string][]byte) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.files = files
}

// Collect collects files from the workload.
// For an EruFileCollector, identifier represents the workload id.
// All paths in files should be absolute, or related to the current working dir.
func (e *EruFileCollector) Collect(ctx context.Context, identifier string, files []string) error {
	if len(files) == 0 {
		return nil
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	var err error
	resp, err := e.eru.Copy(ctx, &corepb.CopyOptions{
		Targets: map[string]*corepb.CopyPaths{
			identifier: {Paths: files},
		},
	})
	if err != nil {
		return err
	}

	var (
		filereaders = make(map[string]*io.PipeReader)
		filewriters = make(map[string]*io.PipeWriter)
		wg          = sync.WaitGroup{}
	)

	for _, file := range files {
		reader, writer := io.Pipe()
		filereaders[file] = reader
		filewriters[file] = writer
	}

	wg.Add(1)
	go func() {
		var message *corepb.CopyMessage
		defer func() {
			for _, writer := range filewriters {
				writer.Close()
			}
			wg.Done()
		}()
		for {
			message, err = resp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}
			if message.Error != "" {
				err = fmt.Errorf(message.Error)
			}

			writer, ok := filewriters[message.Path]
			if !ok {
				continue
			}
			if _, err = writer.Write(message.Data); err != nil {
				return
			}
		}
	}()

	for path, reader := range filereaders {
		wg.Add(1)
		go func(path string, reader io.Reader) {
			defer wg.Done()
			tr := tar.NewReader(reader)
			for {
				header, err := tr.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					return
				}
				// we only have 1 file
				switch header.Typeflag {
				case tar.TypeReg:
					buffer := &bytes.Buffer{}
					_, err = io.Copy(buffer, tr)
					if err != nil {
						return
					}
					e.filesLock.Lock()
					e.files[path] = buffer.Bytes()
					e.filesLock.Unlock()
				default:
				}
			}
		}(path, reader)
	}
	wg.Wait()

	return nil
}

// CopyTo copies files to the workload.
// For an EruFileCollector, identifier represents the workload id.
// All paths in files should be absolute, or related to the current working dir.
func (e *EruFileCollector) CopyTo(ctx context.Context, identifier string, files []string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	data := map[string][]byte{}
	if len(files) == 0 {
		data = e.files
	} else {
		for filename, content := range e.files {
			data[filename] = content
		}
	}

	if len(data) == 0 {
		return nil
	}

	if err := e.createEssentialDirs(ctx, identifier, data); err != nil {
		return err
	}

	resp, err := e.eru.Send(ctx, &corepb.SendOptions{
		Ids:  []string{identifier},
		Data: data,
	})
	if err != nil {
		return err
	}

	for {
		m, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if m.Error != "" {
			return errors.WithMessagef(ErrorCopyToContainer, "path: %s, error: %s", m.Path, m.Error)
		}
	}
	return nil
}

// createEssentialDirs creates all the necessary directories for files.
// Since ERU doesn't provide a `docker cp -` like feature,
// we must ensure all the dirs of the files we send to container exist.
// So we use this function to create all dirs.
func (e *EruFileCollector) createEssentialDirs(ctx context.Context, identifier string, files map[string][]byte) error {
	dirs := map[string]struct{}{}
	for path := range files {
		dirs[filepath.Dir(path)] = struct{}{}
	}

	shell := []string{"/bin/mkdir", "-p"}
	for path := range dirs {
		shell = append(shell, path)
	}

	exec, err := e.eru.ExecuteWorkload(ctx)
	if err != nil {
		return err
	}

	if err := exec.Send(&corepb.ExecuteWorkloadOptions{
		WorkloadId: identifier,
		Commands:   shell,
		Workdir:    "/",
	}); err != nil {
		return err
	}

	for {
		message, err := exec.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		data := string(message.Data)
		if strings.HasPrefix(data, exitMessagePrefix) {
			exitcode, err := strconv.Atoi(strings.TrimPrefix(data, exitMessagePrefix))
			if err != nil {
				return err
			}
			if exitcode != 0 {
				return errors.WithMessagef(common.ErrExecutionError, "exitcode: %d", exitcode)
			}
		}
	}
	return exec.CloseSend()
}

// Files returns all file names including path this collector holds.
func (e *EruFileCollector) Files() []string {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	var files []string
	for file := range e.files {
		files = append(files, file)
	}
	return files
}
