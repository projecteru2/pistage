package common

import (
	"io"
	"time"
)

type Job struct {
	Name        string            `yaml:"name" json:"name"`
	Image       string            `yaml:"image" json:"image"`
	DependsOn   []string          `yaml:"depends_on" json:"depends_on"`
	Steps       []*Step           `yaml:"steps" json:"steps"`
	Timeout     int               `yaml:"timeout" json:"timeout"`
	Environment map[string]string `yaml:"env" json:"env"`
	Files       []string          `yaml:"files" json:"files"`

	fileCollector FileCollector `yaml:"-" json:"-"`
}

func (j *Job) SetFileCollector(fc FileCollector) {
	j.fileCollector = fc
}

func (j *Job) GetFileCollector() FileCollector {
	return j.fileCollector
}

type Step struct {
	Name        string            `yaml:"name" json:"name"`
	Uses        string            `yaml:"uses" json:"uses"`
	With        map[string]string `yaml:"with" json:"with"`
	Run         []string          `yaml:"run" json:"run"`
	OnError     []string          `yaml:"on_error" json:"on_error"`
	Environment map[string]string `yaml:"env" json:"env"`
}

type JobRunStatus string

var (
	JobRunStatusPending  JobRunStatus = "pending"
	JobRunStatusRunning  JobRunStatus = "running"
	JobRunStatusFinished JobRunStatus = "finished"
	JobRunStatusCanceled JobRunStatus = "canceled"
)

type Run struct {
	ID       string    `json:"id"`
	Phistage string    `json:"phistage"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
}

type JobRun struct {
	ID        string        `json:"id"`
	Phistage  string        `json:"phistage"`
	Job       string        `json:"job"`
	Status    JobRunStatus  `json:"status"`
	Start     time.Time     `json:"start"`
	End       time.Time     `json:"end"`
	LogTracer io.ReadWriter `json:"-"`
}
