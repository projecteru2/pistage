package common

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Config is global config for pistage
type Config struct {
	Bind               string   `yaml:"bind" default:":9736"`
	StageServerWorkers int      `yaml:"stage_server_workers" default:"10"`
	JobExecutors       []string `yaml:"job_executors" default:"[eru]"`

	DefaultJobExecutor       string `yaml:"default_job_executor" default:"eru"`
	DefaultJobExecuteTimeout int    `yaml:"default_job_execute_timeout" default:"1200"`

	Eru     EruConfig           `yaml:"eru"`
	SSH     SSHConfig           `yaml:"ssh"`
	Storage SQLDataSourceConfig `yaml:"storage"`
	Khorium KhoriumConfig       `yaml:"khorium"`
}

type EruConfig struct {
	Address           string `yaml:"address"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	DefaultPrivileged bool   `yaml:"default_privileged" default:"true"`
	DefaultWorkingDir string `yaml:"default_working_dir" default:"/pistage"`
	DefaultPodname    string `yaml:"default_pod" default:"ci"`
	DefaultJobImage   string `yaml:"default_job_image"`
	DefaultUser       string `yaml:"default_user" default:"root"`
	DefaultNetwork    string `yaml:"default_network" default:"host"`
}

type SSHConfig struct {
	User       string `yaml:"user"`
	PrivateKey string `yaml:"private_key"`
	Address    string `yaml:"address"`
}

type KhoriumConfig struct {
	GitLabUsername    string `yaml:"gitlab_username"`
	GitLabAccessToken string `yaml:"gitlab_access_token"`
}

type SQLDataSourceConfig struct {
	Username     string `yaml:"username" default:"root"`
	Password     string `yaml:"password" default:""`
	Host         string `yaml:"host" default:"localhost"`
	Port         int    `yaml:"port" default:"3306"`
	Database     string `yaml:"database" default:"pistage"`
	MaxLifetime  int    `yaml:"max_lifetime" default:"30"`
	MaxConns     int    `yaml:"max_conns" default:"10"`
	MaxIdleConns int    `yaml:"max_idle_conns" default:"5"`
	Timeout      int    `yaml:"timeout" default:"5"`
	ReadTimeout  int    `yaml:"read_timeout" default:"5"`
	WriteTimeout int    `yaml:"write_timeout" default:"5"`
}

func (c SQLDataSourceConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local&timeout=%ds&readTimeout=%ds&writeTimeout=%ds", c.Username, c.Password, c.Host, c.Port, c.Database, c.Timeout, c.ReadTimeout, c.WriteTimeout)
}

func (c *Config) initDefault() {
	if c.Bind == "" {
		c.Bind = ":9736"
	}
	if c.StageServerWorkers == 0 {
		c.StageServerWorkers = 10
	}
	if len(c.JobExecutors) == 0 {
		c.JobExecutors = []string{"eru"}
	}
	if c.DefaultJobExecutor == "" {
		c.DefaultJobExecutor = "eru"
	}
	if c.DefaultJobExecuteTimeout == 0 {
		c.DefaultJobExecuteTimeout = 1200
	}
	if c.Eru.DefaultWorkingDir == "" {
		c.Eru.DefaultWorkingDir = "/pistage"
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
