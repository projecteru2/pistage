package common

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/projecteru2/pistage/helpers"
)

var ErrorJobNotFound = errors.New("Job not found")

type Pistage struct {
	WorkflowType       string `yaml:"workflow_type" json:"workflow_type"`
	WorkflowIdentifier string `yaml:"workflow_identifier" json:"workflow_identifier"`

	Jobs        map[string]*Job   `yaml:"jobs" json:"jobs"`
	Environment map[string]string `yaml:"env" json:"env"`
	Executor    string            `yaml:"executor" json:"executor"`

	Content     []byte `yaml:"-" json:"-"`
	ContentHash string `yaml:"-" json:"-"`
}

// init set name to all jobs.
func (p *Pistage) init() {
	for jobName, job := range p.Jobs {
		job.Name = jobName
	}
}

func (p *Pistage) Name() string {
	return p.WorkflowIdentifier
}

// validate currently checks only if the dependency graph contains a cycle.
func (p *Pistage) validate() error {
	tp := newTopo()
	for _, job := range p.Jobs {
		tp.addDependencies(job.Name, job.DependsOn...)
	}
	return tp.checkCyclic()
}

func (p *Pistage) GenerateHash() error {
	if p.ContentHash != "" {
		return nil
	}

	content, err := json.Marshal(p)
	if err != nil {
		return errors.Wrap(err, "marshal pistage")
	}

	sha1OfContent, err := helpers.Sha1HexDigest(content)
	if err != nil {
		return errors.Wrap(err, "generate content hash")
	}

	p.Content = content
	p.ContentHash = sha1OfContent
	return nil
}

func UnmarshalPistage(marshalled []byte) (p *Pistage, err error) {
	err = json.Unmarshal(marshalled, &p)
	return
}

// JobDependencies parses the dependency relationship of all jobs,
// returns a list with the order jobs should be executed.
// The sub list item indicates these jobs can be executed parallelly.
// E.g. if the list returned is
// [
//   ["A", "B"],
//   ["C"],
//   ["D", "E"],
// ]
// this means A and B can be executed parallelly, C needs to be waited till A and B finished,
// then D and E can be executed parallelly.
// In a word, the execution should be (A, B) -> (C) -> (D, E)
func (p *Pistage) JobDependencies() ([][]*Job, error) {
	var (
		jobs [][]*Job
		tp   = newTopo()
	)
	for _, job := range p.Jobs {
		tp.addDependencies(job.Name, job.DependsOn...)
	}

	deps, err := tp.graph()
	if err != nil {
		return nil, err
	}

	for _, jobNames := range deps {
		j := []*Job{}
		for _, jobName := range jobNames {
			job, ok := p.Jobs[jobName]
			if !ok {
				return nil, ErrorJobNotFound
			}
			j = append(j, job)
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

// JobStream parses the dependency relationship of all jobs,
// returns a job channel, a finish channel, and a function indicates that
// all jobs are finished.
// All jobs to be executed should be acquired from job channel,
// if the job is finished, put it back to finish channel, this will continue
// the following parsing; finish function should be invoked after all jobs
// are finished, or when error occurs and early break the execution.
// The channel only contains the names of jobs, so use GetJob method to
// retrieve the real job, since it's too complicated to return a channel of jobs.
func (p *Pistage) JobStream() (<-chan string, chan<- string, func()) {
	tp := newTopo()
	for _, job := range p.Jobs {
		tp.addDependencies(job.Name, job.DependsOn...)
	}
	return tp.stream()
}

// GetJob gets job by the given names.
func (p *Pistage) GetJob(name string) (*Job, error) {
	job, ok := p.Jobs[name]
	if !ok {
		return nil, ErrorJobNotFound
	}
	return job, nil
}

// GetJobs gets job list by the given names.
func (p *Pistage) GetJobs(names []string) []*Job {
	var jobs []*Job
	for _, name := range names {
		job, ok := p.Jobs[name]
		if !ok {
			continue
		}
		jobs = append(jobs, job)
	}
	return jobs
}

// FromSpec build a Pistage from a spec file.
func FromSpec(content []byte) (*Pistage, error) {
	p := &Pistage{}
	err := yaml.Unmarshal(content, p)
	if err != nil {
		return nil, err
	}

	p.init()
	return p, p.validate()
}

// MarshalPistage marshals pistage back into yaml format.
func MarshalPistage(pistage *Pistage) ([]byte, error) {
	return yaml.Marshal(pistage)
}

func EpochMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// PistageTask contains a pistage and an output tracing stream.
// Tracing stream is used to trace this process.
type PistageTask struct {

	// todo add a context
	Ctx context.Context

	// Pistage holds the pistage to execute.
	Pistage *Pistage

	// JobType is used to distinguish cli command kind, like rollback/apply
	JobType string

	// Output is the tracing stream for logs.
	// It's an io.WriteCloser, closing this output indicates that
	// all logs have been written into this stream, the pistage
	// has finished.
	// Do remember to close the Output, or find some other methods to
	// control the halt of the process.
	Output io.WriteCloser
}
