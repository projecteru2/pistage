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
	"gopkg.in/yaml.v3"

	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/helpers"
	"github.com/projecteru2/pistage/store"
)

type FileSystemStore struct {
	root           string
	mutex          sync.Mutex
	snowflake      *snowflake.Node
	khoriumManager *store.KhoriumManager
}

func NewFileSystemStore(root string, khoriumManager *store.KhoriumManager) (*FileSystemStore, error) {
	sn, err := store.NewSnowflake()
	if err != nil {
		return nil, err
	}
	return &FileSystemStore{
		root:           root,
		snowflake:      sn,
		khoriumManager: khoriumManager,
	}, nil
}

// ${root}/pistage/${sha1 of pistage name}/meta/current
// ${root}/pistage/${sha1 of pistage name}/meta/${sha1 of file content}
// ${root}/pistage/${sha1 of pistage name}/run/${run id}
func (fs *FileSystemStore) CreatePistage(ctx context.Context, pistage *common.Pistage) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(pistage.Name)
	if err != nil {
		return err
	}

	rootPath := filepath.Join(fs.root, "pistage", sha1OfName)
	metaPath := filepath.Join(rootPath, "meta")
	if err := os.MkdirAll(metaPath, 0700); err != nil {
		return err
	}
	// run path
	if err := os.MkdirAll(filepath.Join(rootPath, "run"), 0700); err != nil {
		return err
	}

	content, err := json.Marshal(pistage)
	if err != nil {
		return err
	}

	sha1OfFile, err := helpers.Sha1HexDigest(content)
	if err != nil {
		return err
	}

	if err := helpers.WriteIfNotExist(filepath.Join(metaPath, sha1OfFile), content); err != nil {
		return err
	}

	if err := helpers.OverWriteFile(filepath.Join(metaPath, "current"), sha1OfFile); err != nil {
		return err
	}
	return nil
}

func (fs *FileSystemStore) GetPistage(ctx context.Context, name string) (*common.Pistage, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	metaPath := filepath.Join(fs.root, "pistage", sha1OfName, "meta")
	sha1OfFile, err := ioutil.ReadFile(filepath.Join(metaPath, "current"))
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(filepath.Join(metaPath, string(sha1OfFile)))
	if err != nil {
		return nil, err
	}

	p := &common.Pistage{}
	if err := json.Unmarshal(content, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (fs *FileSystemStore) DeletePistage(ctx context.Context, name string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
	if err != nil {
		return err
	}

	return os.RemoveAll(filepath.Join(fs.root, "pistage", sha1OfName))
}

// ${root}/run/${id}/jobrun/${jobrun id}
// ${root}/run/${id}/run
// ${root}/pistage/${sha1 of pistage name}/run/${run id}
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
	if err := helpers.WriteIfNotExist(filepath.Join(rootPath, "run"), content); err != nil {
		return err
	}

	// pistage run
	sha1OfPistageName, err := helpers.Sha1HexDigest(run.Pistage)
	if err != nil {
		return err
	}
	filename := filepath.Join(fs.root, "pistage", sha1OfPistageName, "run", run.ID)
	if err := helpers.OverWriteFile(filename, ""); err != nil {
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
	return helpers.OverWriteFile(runPath, content)
}

// ${root}/pistage/${sha1 of pistage name}/run/${run id}
func (fs *FileSystemStore) GetRunsByPistage(ctx context.Context, name string) ([]*common.Run, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	pattern := filepath.Join(fs.root, "pistage", sha1OfName, "run", "*")
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
	return helpers.WriteIfNotExist(jobRunPath, content)
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
	return helpers.OverWriteFile(jobRunPath, content)
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
	if err := helpers.OverWriteFile(jobRunPath, content); err != nil {
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

	sha1OfName, err := helpers.Sha1HexDigest(job.Name)
	if err != nil {
		return err
	}

	content, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return helpers.WriteIfNotExist(filepath.Join(rootPath, sha1OfName), content)
}

// ${root}/registered/job/${sha1 of job name}
func (fs *FileSystemStore) GetRegisteredJob(ctx context.Context, name string) (*common.Job, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
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

	sha1OfName, err := helpers.Sha1HexDigest(step.Name)
	if err != nil {
		return err
	}

	content, err := json.Marshal(step)
	if err != nil {
		return err
	}
	return helpers.WriteIfNotExist(filepath.Join(rootPath, sha1OfName), content)
}

// ${root}/registered/step/${sha1 of step name}
func (fs *FileSystemStore) GetRegisteredStep(ctx context.Context, name string) (*common.Step, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
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

// ${root}/pistage/${sha1 of pistage name}/vars
func (fs *FileSystemStore) SetVariablesForPistage(ctx context.Context, name string, vars map[string]string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
	if err != nil {
		return err
	}

	rootPath := filepath.Join(fs.root, "pistage", sha1OfName)
	if err := os.MkdirAll(rootPath, 0700); err != nil {
		return err
	}

	content, err := json.Marshal(vars)
	if err != nil {
		return err
	}
	return helpers.OverWriteFile(filepath.Join(rootPath, "vars"), content)
}

// ${root}/pistage/${sha1 of pistage name}/vars
func (fs *FileSystemStore) GetVariablesForPistage(ctx context.Context, name string) (map[string]string, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	varsPath := filepath.Join(fs.root, "pistage", sha1OfName, "vars")
	content, err := ioutil.ReadFile(varsPath)
	if err != nil {
		// if not exist, we return empty vars
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	vars := map[string]string{}
	if err := json.Unmarshal(content, &vars); err != nil {
		return nil, err
	}
	return vars, nil
}

// ${root}/registered/khoriumstep/${sha1 of step name}/khoriumstep.yml
func (fs *FileSystemStore) RegisterKhoriumStep(ctx context.Context, ks *common.KhoriumStep) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(ks.Name)
	if err != nil {
		return err
	}

	rootPath := filepath.Join(fs.root, "registered", "khoriumstep", sha1OfName)
	content, err := yaml.Marshal(ks)
	if err != nil {
		return err
	}
	return helpers.WriteIfNotExist(filepath.Join(rootPath, "khoriumstep.yml"), content)
}

// ${root}/registered/khoriumstep/${sha1 of step name}/khoriumstep.yml
// ${root}/registered/khoriumstep/${sha1 of step name}/...
func (fs *FileSystemStore) GetRegisteredKhoriumStep2(ctx context.Context, name string) (*common.KhoriumStep, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	sha1OfName, err := helpers.Sha1HexDigest(name)
	if err != nil {
		return nil, err
	}

	rootPath := filepath.Join(fs.root, "registered", "khoriumstep", sha1OfName)
	content, err := ioutil.ReadFile(filepath.Join(rootPath, "khoriumstep.yml"))
	if err != nil {
		return nil, err
	}

	ks, err := common.LoadKhoriumStep(content)
	if err != nil {
		return nil, err
	}

	traverse := func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		c, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		ks.Files[strings.TrimPrefix(file, rootPath)] = c
		return nil
	}

	if err := filepath.Walk(rootPath, traverse); err != nil {
		return nil, err
	}
	return ks, nil
}

func (fs *FileSystemStore) GetRegisteredKhoriumStep(ctx context.Context, name string) (*common.KhoriumStep, error) {
	return fs.khoriumManager.GetKhoriumStep(ctx, name)
}
