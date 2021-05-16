package filesystem

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/bwmarrin/snowflake"
	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/store"
)

type FileSystemStore struct {
	root      string
	mutex     sync.Mutex
	snowflake *snowflake.Node
}

func NewFileSystemStore(root string) (*FileSystemStore, error) {
	sn, err := store.NewSnowflake()
	if err != nil {
		return nil, err
	}
	return &FileSystemStore{
		root:      root,
		snowflake: sn,
	}, nil
}

// ${root}/phistage/${sha1 of phistage name}/meta/current
// ${root}/phistage/${sha1 of phistage name}/meta/${sha1 of file content}
// ${root}/phistage/${sha1 of phistage name}/job/${sha1 of job name}/run/${run id}
func (fs *FileSystemStore) CreatePhistage(ctx context.Context, phistage *common.Phistage) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := Sha1HexDigest(phistage.Name)
	if err != nil {
		return err
	}

	rootPath := filepath.Join(fs.root, "phistage", sha1OfName)
	metaPath := filepath.Join(rootPath, "meta")
	if err := os.MkdirAll(metaPath, 0700); err != nil {
		return err
	}

	for _, job := range phistage.Jobs {
		sha1OfJobName, err := Sha1HexDigest(job.Name)
		if err != nil {
			return err
		}
		jobPath := filepath.Join(rootPath, "job", sha1OfJobName, "run")
		if err := os.MkdirAll(jobPath, 0700); err != nil {
			return err
		}
	}

	content, err := json.Marshal(phistage)
	if err != nil {
		return err
	}

	sha1OfFile, err := Sha1HexDigestForBytes(content)
	if err != nil {
		return err
	}

	if err := writeIfNotExist(filepath.Join(metaPath, sha1OfFile), content); err != nil {
		return err
	}

	if err := overrideFile(filepath.Join(metaPath, "current"), sha1OfFile); err != nil {
		return err
	}
	return nil
}

func (fs *FileSystemStore) GetPhistage(ctx context.Context, name string) (*common.Phistage, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	metaPath := filepath.Join(fs.root, "phistage", sha1OfName, "meta")
	sha1OfFile, err := ioutil.ReadFile(filepath.Join(metaPath, "current"))
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(filepath.Join(metaPath, string(sha1OfFile)))
	if err != nil {
		return nil, err
	}

	p := &common.Phistage{}
	if err := json.Unmarshal(content, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (fs *FileSystemStore) DeletePhistage(ctx context.Context, name string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := Sha1HexDigest(name)
	if err != nil {
		return err
	}

	return os.RemoveAll(filepath.Join(fs.root, "phistage", sha1OfName))
}

// ${root}/run/${id}/run
// ${root}/phistage/${sha1 of phistage name}/job/${sha1 of job name}/run/${run id}
func (fs *FileSystemStore) CreateRun(ctx context.Context, run *common.Run) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	run.ID = fs.snowflake.Generate().String()

	runPath := filepath.Join(fs.root, "run", run.ID, "run")
	if err := os.MkdirAll(runPath, 0700); err != nil {
		return err
	}
	content, err := json.Marshal(run)
	if err != nil {
		return err
	}
	if err := writeIfNotExist(runPath, content); err != nil {
		return err
	}

	sha1OfPhistageName, err := Sha1HexDigest(run.Phistage)
	if err != nil {
		return err
	}
	sha1OfJobName, err := Sha1HexDigest(run.Job)
	if err != nil {
		return err
	}
	filename := filepath.Join(fs.root, "phistage", sha1OfPhistageName, "job", sha1OfJobName, "run", run.ID)
	if err := overrideFile(filename, ""); err != nil {
		return err
	}

	return nil
}

func (fs *FileSystemStore) GetRun(ctx context.Context, id string) (*common.Run, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	runPath := filepath.Join(fs.root, "run", id, "run")
	content, err := ioutil.ReadFile(runPath)
	if err != nil {
		return nil, err
	}

	run := &common.Run{}
	if err := json.Unmarshal(content, run); err != nil {
		return nil, err
	}
	return run, nil
}

// ${root}/phistage/${sha1 of phistage name}/job/${sha1 of job name}/run/${run id}
func (fs *FileSystemStore) GetRunByPhistage(ctx context.Context, name string) ([]*common.Run, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	pattern := filepath.Join(fs.root, "phistage", sha1OfName, "job", "*", "run", "*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var runs []*common.Run
	for _, m := range matches {
		runID := filepath.Base(m)
		run, err := fs.GetRun(ctx, runID)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, nil
}

// ${root}/phistage/${sha1 of phistage name}/job/${sha1 of job name}/run/${run id}
func (fs *FileSystemStore) GetRunByJob(ctx context.Context, phistageName, jobName string) ([]*common.Run, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfPhistageName, err := Sha1HexDigest(phistageName)
	if err != nil {
		return nil, err
	}
	sha1OfJobName, err := Sha1HexDigest(jobName)
	if err != nil {
		return nil, err
	}

	pattern := filepath.Join(fs.root, "phistage", sha1OfPhistageName, "job", sha1OfJobName, "run", "*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var runs []*common.Run
	for _, m := range matches {
		runID := filepath.Base(m)
		run, err := fs.GetRun(ctx, runID)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, nil
}

// ${root}/registered/job/${sha1 of job name}
func (fs *FileSystemStore) RegisterJob(ctx context.Context, job *common.Job) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	rootPath := filepath.Join(fs.root, "registered", "job")
	if err := os.MkdirAll(rootPath, 0700); err != nil {
		return err
	}

	sha1OfName, err := Sha1HexDigest(job.Name)
	if err != nil {
		return err
	}

	content, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return writeIfNotExist(filepath.Join(rootPath, sha1OfName), content)
}

// ${root}/registered/job/${sha1 of job name}
func (fs *FileSystemStore) GetRegisteredJob(ctx context.Context, name string) (*common.Job, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(filepath.Join(fs.root, "registered", "job", sha1OfName))
	if err != nil {
		return nil, err
	}

	job := &common.Job{}
	if err := json.Unmarshal(content, job); err != nil {
		return nil, err
	}
	return job, nil
}

// ${root}/registered/step/${sha1 of step name}
func (fs *FileSystemStore) RegisterStep(ctx context.Context, step *common.Step) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	rootPath := filepath.Join(fs.root, "registered", "step")
	if err := os.MkdirAll(rootPath, 0700); err != nil {
		return err
	}

	sha1OfName, err := Sha1HexDigest(step.Name)
	if err != nil {
		return err
	}

	content, err := json.Marshal(step)
	if err != nil {
		return err
	}
	return writeIfNotExist(filepath.Join(rootPath, sha1OfName), content)
}

// ${root}/registered/step/${sha1 of step name}
func (fs *FileSystemStore) GetRegisteredStep(ctx context.Context, name string) (*common.Step, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(filepath.Join(fs.root, "registered", "step", sha1OfName))
	if err != nil {
		return nil, err
	}

	step := &common.Step{}
	if err := json.Unmarshal(content, step); err != nil {
		return nil, err
	}
	return step, nil
}
