package common

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var (
	// ErrorJobHasNoName is returned when the job has no name.
	// This is only returned when directly loading a job from yaml specification.
	// When loading from a pistage, this won't be returned since jobs in pistage
	// is a map, the names are the keys.
	ErrorJobHasNoName = errors.New("Job has no name")

	// ErrorStepHasNoName is returned when the step has no name.
	ErrorStepHasNoName = errors.New("Step has no name")
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

func LoadJob(content []byte) (*Job, error) {
	j := &Job{}
	err := yaml.Unmarshal(content, j)
	if err != nil {
		return nil, err
	}
	if j.Name == "" {
		return nil, ErrorJobHasNoName
	}
	return j, nil
}

type Step struct {
	Name        string            `yaml:"name" json:"name"`
	Uses        string            `yaml:"uses" json:"uses"`
	With        map[string]string `yaml:"with" json:"with"`
	Run         []string          `yaml:"run" json:"run"`
	OnError     []string          `yaml:"on_error" json:"on_error"`
	Environment map[string]string `yaml:"env" json:"env"`
}

func LoadStep(content []byte) (*Step, error) {
	s := &Step{}
	err := yaml.Unmarshal(content, s)
	if err != nil {
		return nil, err
	}
	if s.Name == "" {
		return nil, ErrorStepHasNoName
	}
	return s, nil
}

// RunStatus is the status of a Run or a JobRun
type RunStatus string

var (
	RunStatusPending  RunStatus = "pending"
	RunStatusRunning  RunStatus = "running"
	RunStatusFinished RunStatus = "finished"
	RunStatusFailed   RunStatus = "failed"
	RunStatusCanceled RunStatus = "canceled"
)

type Run struct {
	ID      string    `json:"id"`
	Pistage string    `json:"pistage"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
}

type JobRun struct {
	ID        string             `json:"id"`
	Pistage   string             `json:"pistage"`
	Job       string             `json:"job"`
	Status    RunStatus          `json:"status"`
	Start     time.Time          `json:"start"`
	End       time.Time          `json:"end"`
	LogTracer io.ReadWriteCloser `json:"-"`
}

var (
	// ErrorInputIsRequired is returned when a value for KhoriumStepInput is required but not given.
	ErrorInputIsRequired = errors.New("Input is required")

	// ErrorMustSpecifyRun is returned when run is not given in the specification.
	ErrorMustSpecifyRun = errors.New("Must specify run")

	// ErrorMustSpecifyMain is returned when main is not given in run part.
	ErrorMustSpecifyMain = errors.New("Must specify main")
)

// KhoriumStep is the predefined step.
// It can be used as a step to execute during a job.
type KhoriumStep struct {
	Name        string                       `yaml:"name" json:"name"`
	Description string                       `yaml:"description" json:"description"`
	Inputs      map[string]*KhoriumStepInput `yaml:"inputs" json:"inputs"`
	Run         *KhoriumStepRun              `yaml:"run" json:"run"`

	// Files contains all the files within this KhoriumStep, filename with path as key, content as value.
	// They can be binary executable, or scripts, as a tarball.
	Files map[string][]byte `yaml:"-" json:"-"`
}

func (ks *KhoriumStep) Validate() error {
	if ks.Name == "" {
		return ErrorStepHasNoName
	}
	if ks.Run == nil {
		return ErrorMustSpecifyRun
	}
	if ks.Run.Main == "" {
		return ErrorMustSpecifyMain
	}
	return nil
}

// BuildEnvironmentVariables builds an environment variables map for the input.
// If the value is required but not given, will return an error.
// The values in KhoriumStepInput will be set as an environment variable in the format
// KHORIUMSTEP_INPUT_${upper case of the input name}.
func (ks *KhoriumStep) BuildEnvironmentVariables(vars map[string]string) (map[string]string, error) {
	envs := map[string]string{}
	for name, input := range ks.Inputs {
		key := fmt.Sprintf("KHORIUMSTEP_INPUT_%s", strings.ToUpper(name))
		value, ok := vars[name]

		switch {
		case input.Required && !ok:
			return nil, errors.WithMessagef(ErrorInputIsRequired, "input: %s", name)
		case ok:
			envs[key] = value
		default:
			envs[key] = input.Default
		}
	}
	return envs, nil
}

// KhoriumStepInput is the inputs of KhoriumStep.
type KhoriumStepInput struct {
	Description string `yaml:"description" json:"description"`
	Default     string `yaml:"default" json:"default"`
	Required    bool   `yaml:"required" json:"required"`
}

// KhoriumStepRun is the command to run of KhoriumStep
type KhoriumStepRun struct {
	Main string `yaml:"main" json:"main"`
	Post string `yaml:"post" json:"post"`
}

func LoadKhoriumStep(content []byte) (*KhoriumStep, error) {
	ks := &KhoriumStep{}
	err := yaml.Unmarshal(content, ks)
	if err != nil {
		return nil, err
	}
	if ks.Files == nil {
		ks.Files = map[string][]byte{}
	}
	return ks, ks.Validate()
}
