package common

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Config is global config for phistage
type Config struct {
	Bind          string `yaml:"bind" default:":9736"`
	StagerWorkers int    `yaml:"stager_workers" default:"10"`

	DefaultJobExecutor       string `yaml:"default_job_executor" default:"eru"`
	DefaultJobExecuteTimeout int    `yaml:"default_job_execute_timeout" default:"1200"`

	Eru     EruConfig     `yaml:"eru"`
	Storage StorageConfig `yaml:"storage"`
}

type EruConfig struct {
	Address           string `yaml:"address"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	DefaultPrivileged bool   `yaml:"default_privileged" default:"true"`
	DefaultWorkingDir string `yaml:"default_working_dir" default:"/phistage"`
	DefaultPodname    string `yaml:"default_pod" default:"ci"`
	DefaultJobImage   string `yaml:"default_job_image"`
	DefaultUser       string `yaml:"default_user" default:"root"`
	DefaultNetwork    string `yaml:"default_network" default:"host"`
}

type StorageConfig struct {
	Type                string `yaml:"type"`
	FileSystemStoreRoot string `yaml:"filesystem_store_root"`
}

func (c *Config) initDefault() {
	if c.Bind == "" {
		c.Bind = ":9736"
	}
	if c.StagerWorkers == 0 {
		c.StagerWorkers = 10
	}
	if c.DefaultJobExecutor == "" {
		c.DefaultJobExecutor = "eru"
	}
	if c.DefaultJobExecuteTimeout == 0 {
		c.DefaultJobExecuteTimeout = 1200
	}
	if c.Eru.DefaultWorkingDir == "" {
		c.Eru.DefaultWorkingDir = "/phistage"
	}
	if c.Eru.DefaultPodname == "" {
		c.Eru.DefaultPodname = "ci"
	}
	if c.Eru.DefaultUser == "" {
		c.Eru.DefaultUser = "root"
	}
	if c.Eru.DefaultNetwork == "" {
		c.Eru.DefaultNetwork = "host"
	}
}

func LoadConfigFromFile(path string) (*Config, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err := yaml.Unmarshal(content, config); err != nil {
		return nil, err
	}
	config.initDefault()
	return config, nil
}
