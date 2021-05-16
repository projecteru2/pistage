package common

// Config is global config for phistage
type Config struct {
	// DefaultJobExecutor is used for jobs who doesn't declare what
	// executor to use, default value will be eru.
	DefaultJobExecutor string `yaml:"default_job_executor"`
	// DefaultJobExecuteTimeout is the timeout for the entire job,
	// including all steps with the job.
	DefaultJobExecuteTimeout int `yaml:"default_job_execute_timeout"`

	// Number of stager workers
	StagerWorkers int `yaml:"stager_works"`

	// Eru config
	EruAddress  string `yaml:"eru_address"`
	EruUsername string `yaml:"eru_username"`
	EruPassword string `yaml:"eru_password"`

	// FileSystem storage config
	FileSystemStoreRoot string `yaml:"filesystem_store_root"`
}
