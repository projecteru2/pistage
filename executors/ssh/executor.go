package ssh

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/projecteru2/pistage/common"
	"github.com/projecteru2/pistage/helpers"
	"github.com/projecteru2/pistage/helpers/command"
	"github.com/projecteru2/pistage/helpers/variable"
	"github.com/projecteru2/pistage/store"
)

var (
	sshExecutorRootWorkingDir            = "_pistage"
	sshExecutorKhoriumStepRootWorkingDir = "_pistage_khoriumstep"

	ErrorNoHome = errors.New("No SSH Home")
)

type SSHJobExecutor struct {
	store  store.Store
	config *common.Config

	client *ssh.Client
	home   string

	job     *common.Job
	pistage *common.Pistage

	output         io.Writer
	workingDir     string
	jobEnvironment map[string]string
}

// NewEruJobExecutor creates an ERU executor for this job.
// Since job needs to know its context, pistage is assigned too.
func NewSSHJobExecutor(job *common.Job, pistage *common.Pistage, output io.Writer, client *ssh.Client, store store.Store, config *common.Config) (*SSHJobExecutor, error) {
	// get the current working dir as writable home.
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	// home dir should be $HOME after login
	out, err := session.Output("echo $HOME")
	if err != nil {
		return nil, err
	}

	home := strings.TrimSuffix(string(out), "\n")
	if len(home) == 0 {
		home = "/home/" + config.SSH.User
	}

	return &SSHJobExecutor{
		store:          store,
		config:         config,
		client:         client,
		home:           home,
		job:            job,
		pistage:        pistage,
		output:         output,
		jobEnvironment: pistage.Environment,
	}, nil
}

func executeCommand(client *ssh.Client, cmd, home string, envs map[string]string, output io.Writer) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	envExports := command.RenderEnvironmentForSSH(envs)

	commandShards := []string{
		envExports,
		fmt.Sprintf("cd %s", home),
		cmd,
	}
	commandToExecute := strings.Join(commandShards, "\n")
	session.Stdout = output
	session.Stderr = output

	return session.Run(commandToExecute)
}

// Prepare does all the preparations before actually running a job
func (s *SSHJobExecutor) Prepare(ctx context.Context) error {
	preparations := []func(context.Context) error{
		s.prepareJobRuntime,
		s.prepareFileContext,
	}
	for _, f := range preparations {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}

// prepareJobRuntime creates a working dir for this job.
func (s *SSHJobExecutor) prepareJobRuntime(ctx context.Context) error {
	digest, err := helpers.Sha1HexDigest(fmt.Sprintf("%s:%s", s.pistage.Name(), s.job.Name))
	if err != nil {
		return err
	}

	workingDir := filepath.Join(s.home, sshExecutorRootWorkingDir, digest)
	cmd := fmt.Sprintf("mkdir -p %s", workingDir)
	if err := executeCommand(s.client, cmd, s.workingDir, nil, io.Discard); err != nil {
		return err
	}

	s.workingDir = workingDir
	return nil
}

func (s *SSHJobExecutor) prepareFileContext(ctx context.Context) error {
	dependentJobs := s.pistage.GetJobs(s.job.DependsOn)
	for _, job := range dependentJobs {
		fc := job.GetFileCollector()
		if fc == nil {
			continue
		}
		if err := fc.CopyTo(ctx, s.workingDir, nil); err != nil {
			return err
		}
	}
	return nil
}

// defaultEnvironmentVariables sets some useful information into environment variables.
// This will be set to the whole running context within the workload.
func (s *SSHJobExecutor) defaultEnvironmentVariables() map[string]string {
	return map[string]string{
		"PHISTAGE_WORKING_DIR": s.workingDir,
		"PHISTAGE_JOB_NAME":    s.job.Name,
	}
}

// Execute will execute all steps within this job one by one
func (s *SSHJobExecutor) Execute(ctx context.Context) error {
	return s.executeSteps(ctx, s.job.Steps)
}

// executeSteps will execute steps, steps can be steps or rollback_steps
func (s *SSHJobExecutor) executeSteps(ctx context.Context, steps []*common.Step) error {
	for _, step := range steps {
		var err error
		switch step.Uses {
		case "":
			err = s.executeStep(ctx, step)
		default:
			// step, err = e.replaceStepWithUses(ctx, step)
			// if err != nil {
			// 	return err
			// }
			err = s.executeKhoriumStep(ctx, step)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// executeStep executes a step.
// It first replace the step with uses if uses is given,
// then prepare the arguments and environments to the command.
// Then execute the command, retrieve the output, the execution will stop if any error occurs.
// It then retries to execute the OnError commands, also with the arguments and environments.
func (s *SSHJobExecutor) executeStep(ctx context.Context, step *common.Step) error {
	var (
		err  error
		vars map[string]string
	)

	environment := command.MergeVariables(s.jobEnvironment, step.Environment)

	defer func() {
		if !errors.Is(err, common.ErrExecutionError) {
			return
		}
		if err := s.executeCommands(ctx, step.OnError, step.With, environment, vars); err != nil {
			logrus.WithField("step", step.Name).WithError(err).Errorf("[EruJobExecutor] error when executing on_error")
		}
	}()

	err = s.executeCommands(ctx, step.Run, step.With, environment, vars)
	return err
}

// executeKhoriumStep executes a KhoriumStep defined by step.Uses.
func (s *SSHJobExecutor) executeKhoriumStep(ctx context.Context, step *common.Step) error {
	ks, err := s.store.GetRegisteredKhoriumStep(ctx, step.Uses)
	if err != nil {
		return err
	}

	arguments, err := variable.RenderArguments(step.With, step.Environment, map[string]string{})
	if err != nil {
		return err
	}

	ksEnv, err := ks.BuildEnvironmentVariables(arguments)
	if err != nil {
		return err
	}
	envs := command.MergeVariables(s.defaultEnvironmentVariables(), step.Environment)
	envs = command.MergeVariables(envs, ksEnv)

	// Prepare KhoriumStep environment.
	// Make the proper working dir, and copy the files to this dir.
	digest, err := helpers.Sha1HexDigest(fmt.Sprintf("%s:%s", s.pistage.Name(), s.job.Name))
	if err != nil {
		return err
	}
	khoriumStepWorkingDir := filepath.Join(s.home, sshExecutorKhoriumStepRootWorkingDir, digest)
	defer s.cleanupDir(khoriumStepWorkingDir)

	fc := NewSSHFileCollector(s.client)
	fc.SetFiles(ks.Files)
	if err := fc.CopyTo(ctx, khoriumStepWorkingDir, nil); err != nil {
		return err
	}

	// Now we can execute the script written in specification.
	if err := executeCommand(s.client, ks.Run.Main, khoriumStepWorkingDir, envs, s.output); err != nil {
		return errors.WithMessagef(common.ErrExecutionError, "exec error: %v", err)
	}
	return nil
}

// executeCommands executes cmd with given arguments, environments and variables.
// use args, envs, and reserved vars to build the cmd.
func (s *SSHJobExecutor) executeCommands(ctx context.Context, cmds []string, args, env, vars map[string]string) error {
	if len(cmds) == 0 {
		return nil
	}

	var commands []string
	for _, cmd := range cmds {
		c, err := command.RenderCommand(cmd, args, env, vars)
		if err != nil {
			return err
		}
		commands = append(commands, c)
	}

	shell, err := command.RenderShell(commands)
	if err != nil {
		return err
	}

	envs := command.MergeVariables(s.defaultEnvironmentVariables(), env)
	if err := executeCommand(s.client, shell, s.workingDir, envs, s.output); err != nil {
		return errors.WithMessagef(common.ErrExecutionError, "exec error: %v", err)
	}
	return nil
}

// beforeCleanup collects files
func (s *SSHJobExecutor) beforeCleanup(ctx context.Context) error {
	if len(s.job.Files) == 0 {
		return nil
	}

	fc := NewSSHFileCollector(s.client)
	if err := fc.Collect(ctx, s.workingDir, s.job.Files); err != nil {
		return err
	}

	s.job.SetFileCollector(fc)
	return nil
}

func (s *SSHJobExecutor) cleanupDir(dir string) error {
	cmd := fmt.Sprintf("rm -rf %s", dir)
	return executeCommand(s.client, cmd, s.workingDir, nil, io.Discard)
}

// cleanup removes the working dir.
func (s *SSHJobExecutor) cleanup(ctx context.Context) error {
	if s.workingDir == "" {
		return nil
	}
	return s.cleanupDir(s.workingDir)
}

// Cleanup does all the cleanup work
func (ls *SSHJobExecutor) Cleanup(ctx context.Context) error {
	cleanups := []func(context.Context) error{
		ls.beforeCleanup,
		ls.cleanup,
	}
	for _, f := range cleanups {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Rollback is a function can rollback steps by rollback_steps which defined in YAML
func (s *SSHJobExecutor) Rollback(ctx context.Context) error {
	return s.executeSteps(ctx, s.job.RollbackSteps)
}
