package common

type Job struct {
	Name        string            `yaml:"-"`
	Image       string            `yaml:"image"`
	DependsOn   []string          `yaml:"depends_on"`
	Steps       []*Step           `yaml:"steps"`
	Timeout     int               `yaml:"timeout"`
	Environment map[string]string `yaml:"env"`
	Files       []string          `yaml:"files"`

	fileCollector FileCollector
}

func (j *Job) SetFileCollector(fc FileCollector) {
	j.fileCollector = fc
}

func (j *Job) GetFileCollector() FileCollector {
	return j.fileCollector
}

type Step struct {
	Name        string            `yaml:"name"`
	Uses        string            `yaml:"uses"`
	With        map[string]string `yaml:"with"`
	Run         []string          `yaml:"run"`
	Environment map[string]string `yaml:"env"`
}
