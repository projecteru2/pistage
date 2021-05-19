package common

import (
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

// JobDependencies parses the dependency relation of all jobs,
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
	return p, nil
}

// MarshalPhistage marshals phistage back into yaml format.
func MarshalPhistage(phistage *Phistage) ([]byte, error) {
	return yaml.Marshal(phistage)
}
