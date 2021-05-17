package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
// ${root}/phistage/${sha1 of phistage name}/run/${run id}
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
	// run path
	if err := os.MkdirAll(filepath.Join(rootPath, "run"), 0700); err != nil {
		return err
	}

	content, err := json.Marshal(phistage)
	if err != nil {
		return err
	}

	sha1OfFile, err := Sha1HexDigest(content)
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

// ${root}/run/${id}/jobrun/${jobrun id}
// ${root}/run/${id}/run
// ${root}/phistage/${sha1 of phistage name}/run/${run id}
func (fs *FileSystemStore) CreateRun(ctx context.Context, run *common.Run) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	run.ID = fs.snowflake.Generate().String()

	// root
	rootPath := filepath.Join(fs.root, "run", run.ID)
	if err := os.MkdirAll(rootPath, 0700); err != nil {
		return err
	}

	// jobrun
	jobRunPath := filepath.Join(rootPath, "jobrun")
	if err := os.MkdirAll(jobRunPath, 0700); err != nil {
		return err
	}

	// run
	content, err := json.Marshal(run)
	if err != nil {
		return err
	}
	if err := writeIfNotExist(filepath.Join(rootPath, "run"), content); err != nil {
		return err
	}

	// phistage run
	sha1OfPhistageName, err := Sha1HexDigest(run.Phistage)
	if err != nil {
		return err
	}
	filename := filepath.Join(fs.root, "phistage", sha1OfPhistageName, "run", run.ID)
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

func (fs *FileSystemStore) UpdateRun(ctx context.Context, run *common.Run) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	runPath := filepath.Join(fs.root, "run", run.ID, "run")

	content, err := json.Marshal(run)
	if err != nil {
		return err
	}
	return overrideFile(runPath, content)
}

// ${root}/phistage/${sha1 of phistage name}/run/${run id}
func (fs *FileSystemStore) GetRunsByPhistage(ctx context.Context, name string) ([]*common.Run, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	pattern := filepath.Join(fs.root, "phistage", sha1OfName, "run", "*")
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

// ${root}/run/${id}/jobrun/${jobrun id}
func (fs *FileSystemStore) CreateJobRun(ctx context.Context, run *common.Run, jobRun *common.JobRun) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	jobRun.ID = fs.snowflake.Generate().String()

	content, err := json.Marshal(jobRun)
	if err != nil {
		return err
	}

	jobRunPath := filepath.Join(fs.root, "run", run.ID, "jobrun", jobRun.ID)
	return writeIfNotExist(jobRunPath, content)
}

// ${root}/run/${id}/jobrun/${jobrun id}
func (fs *FileSystemStore) GetJobRun(ctx context.Context, runID, jobRunID string) (*common.JobRun, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	jobRunPath := filepath.Join(fs.root, "run", runID, "jobrun", jobRunID)
	content, err := ioutil.ReadFile(jobRunPath)
	if err != nil {
		return nil, err
	}

	jobRun := &common.JobRun{}
	if err := json.Unmarshal(content, jobRun); err != nil {
		return nil, err
	}
	return jobRun, nil
}

// ${root}/run/${id}/jobrun/${jobrun id}
func (fs *FileSystemStore) UpdateJobRun(ctx context.Context, run *common.Run, jobRun *common.JobRun) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	content, err := json.Marshal(jobRun)
	if err != nil {
		return err
	}

	jobRunPath := filepath.Join(fs.root, "run", run.ID, "jobrun", jobRun.ID)
	return overrideFile(jobRunPath, content)
}

// ${root}/run/${id}/jobrun/${jobrun id}
// ${root}/run/${id}/jobrun/${jobrun id}.log
func (fs *FileSystemStore) FinishJobRun(ctx context.Context, run *common.Run, jobRun *common.JobRun) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	jobRun.End = time.Now()
	jobRun.Status = common.JobRunStatusFinished
	content, err := json.Marshal(jobRun)
	if err != nil {
		return err
	}

	jobRunPath := filepath.Join(fs.root, "run", run.ID, "jobrun", jobRun.ID)
	if err := overrideFile(jobRunPath, content); err != nil {
		return err
	}

	jobRunLogPath := filepath.Join(fs.root, "run", run.ID, "jobrun", fmt.Sprintf("%s.log", jobRun.ID))
	f, err := os.OpenFile(jobRunLogPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, jobRun.LogTracer); err != nil {
		return err
	}
	return f.Close()
}

// ${root}/run/${id}/jobrun/${jobrun id}
// ${root}/run/${id}/jobrun/${jobrun id}.log
func (fs *FileSystemStore) GetJobRuns(ctx context.Context, runID string) ([]*common.JobRun, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	pattern := filepath.Join(fs.root, "run", runID, "jobrun", "*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var jobRuns []*common.JobRun
	for _, m := range matches {
		if strings.HasSuffix(m, ".log") {
			continue
		}
		jobRunID := filepath.Base(m)
		jobRun, err := fs.GetJobRun(ctx, runID, jobRunID)
		if err != nil {
			return nil, err
		}
		jobRuns = append(jobRuns, jobRun)
	}
	return jobRuns, nil
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
