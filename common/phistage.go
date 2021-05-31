package common

import (
	"io"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var ErrorJobNotFound = errors.New("Job not found")

type Phistage struct {
	Name        string            `yaml:"name" json:"name"`
	Jobs        map[string]*Job   `yaml:"jobs" json:"jobs"`
	Environment map[string]string `yaml:"env" json:"env"`
	Executor    string            `yaml:"executor" json:"executor"`
}

// init set name to all jobs.
func (p *Phistage) init() {
	for jobName, job := range p.Jobs {
		job.Name = jobName
	}
}

// validate currently checks only if the dependency graph contains a cycle.
func (p *Phistage) validate() error {
	tp := newTopo()
	for _, job := range p.Jobs {
		tp.addDependencies(job.Name, job.DependsOn...)
	}
	return tp.checkCyclic()
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
func (p *Phistage) JobDependencies() ([][]*Job, error) {
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
func (p *Phistage) JobStream() (<-chan string, chan<- string, func()) {
	tp := newTopo()
	for _, job := range p.Jobs {
		tp.addDependencies(job.Name, job.DependsOn...)
	}
	return tp.stream()
}

// GetJob gets job by the given names.
func (p *Phistage) GetJob(name string) (*Job, error) {
	job, ok := p.Jobs[name]
	if !ok {
		return nil, ErrorJobNotFound
	}
	return job, nil
}

// GetJobs gets job list by the given names.
func (p *Phistage) GetJobs(names []string) []*Job {
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

// FromSpec build a Phistage from a spec file.
func FromSpec(content []byte) (*Phistage, error) {
	p := &Phistage{}
	err := yaml.Unmarshal(content, p)
	if err != nil {
		return nil, err
	}

	p.init()
	return p, p.validate()
}

// MarshalPhistage marshals phistage back into yaml format.
func MarshalPhistage(phistage *Phistage) ([]byte, error) {
	return yaml.Marshal(phistage)
}

// PhistageTask contains a phistage and an output tracing stream.
// Tracing stream is used to trace this process.
type PhistageTask struct {
	// Phistage holds the phistage to execute.
	Phistage *Phistage

	// Output is the tracing stream for logs.
	// It's an io.WriteCloser, closing this output indicates that
	// all logs have been written into this stream, the phistage
	// has finished.
	// Do remember to close the Output, or find some other methods to
	// control the halt of the process.
	Output io.WriteCloser
}
