package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/helpers"
	"github.com/projecteru2/phistage/helpers/command"
	"github.com/projecteru2/phistage/helpers/variable"
	"github.com/projecteru2/phistage/store"
)

var (
	sshExecutorRootWorkingDir            = "_phistage"
	sshExecutorKhoriumStepRootWorkingDir = "_phistage_khoriumstep"

	ErrorNoHome = errors.New("No SSH Home")
)

type SSHJobExecutor struct {
	store  store.Store
	config *common.Config

	client *ssh.Client
	home   string

	job      *common.Job
	phistage *common.Phistage

	output         io.Writer
	workingDir     string
	jobEnvironment map[string]string
}

// NewEruJobExecutor creates an ERU executor for this job.
// Since job needs to know its context, phistage is assigned too.
func NewSSHJobExecutor(job *common.Job, phistage *common.Phistage, output io.Writer, client *ssh.Client, store store.Store, config *common.Config) (*SSHJobExecutor, error) {
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
		phistage:       phistage,
		output:         output,
		jobEnvironment: phistage.Environment,
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

// Prepare creates a working dir for this job.
func (s *SSHJobExecutor) Prepare(ctx context.Context) error {
	digest, err := helpers.Sha1HexDigest(fmt.Sprintf("%s:%s", s.phistage.Name, s.job.Name))
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
	for _, step := range s.job.Steps {
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

// replaceStepWithUses replaces the step with the actual step identified by uses.
// Name will not be replaced, the commands to execute aka Run and OnError will be replaced,
// also Environment and With will be merged, uses' environment and with has a lower priority,
// will be overridden by step's environment and with,
// If uses is not given, directly return the original step.
func (s *SSHJobExecutor) replaceStepWithUses(ctx context.Context, step *common.Step) (*common.Step, error) {
	if step.Uses == "" {
		return step, nil
	}

	uses, err := s.store.GetRegisteredStep(ctx, step.Uses)
	if err != nil {
		return nil, err
	}
	realStep := &common.Step{
		Name:        step.Name,
		Run:         uses.Run,
		OnError:     uses.OnError,
		Environment: command.MergeVariables(uses.Environment, step.Environment),
		With:        command.MergeVariables(uses.With, step.With),
	}
	// in case name of this step is empty
	if realStep.Name == "" {
		realStep.Name = uses.Name
	}
	return realStep, nil
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

	vars, err = s.store.GetVariablesForPhistage(ctx, s.phistage.Name)
	if err != nil {
		return err
	}

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

	vars, err := s.store.GetVariablesForPhistage(ctx, s.phistage.Name)
	if err != nil {
		return err
	}

	arguments, err := variable.RenderArguments(step.With, step.Environment, vars)
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
	digest, err := helpers.Sha1HexDigest(fmt.Sprintf("%s:%s", s.phistage.Name, s.job.Name))
	if err != nil {
		return err
	}
	khoriumStepWorkingDir := filepath.Join(s.home, sshExecutorKhoriumStepRootWorkingDir, digest)
	defer s.cleanupDir(khoriumStepWorkingDir)

	files := map[string][]byte{}
	for name, content := range ks.Files {
		files[filepath.Join(khoriumStepWorkingDir, name)] = content
	}
	if err := s.createNecessaryDirs(ctx, files); err != nil {
		return err
	}

	if err := s.sendFiles(files); err != nil {
		return err
	}

	// Now we can execute the script written in specification.
	if err := executeCommand(s.client, ks.Run.Main, khoriumStepWorkingDir, envs, s.output); err != nil {
		return errors.WithMessagef(common.ErrExecutionError, "exec error: %v", err)
	}
	return nil
}

// createNecessaryDirs creates essential dirs for files.
func (s *SSHJobExecutor) createNecessaryDirs(ctx context.Context, files map[string][]byte) error {
	// golang is really, really stupid
	dirs := map[string]struct{}{}
	for path := range files {
		dirs[filepath.Dir(path)] = struct{}{}
	}

	dirnames := []string{}
	for path := range dirs {
		dirnames = append(dirnames, path)
	}
	paths := strings.Join(dirnames, " ")
	cmd := fmt.Sprintf("mkdir -p %s", paths)
	return executeCommand(s.client, cmd, s.workingDir, nil, io.Discard)
}

func (s *SSHJobExecutor) sendFiles(files map[string][]byte) error {
	client, err := sftp.NewClient(s.client)
	if err != nil {
		return err
	}
	defer client.Close()

	for path, content := range files {
		local := bytes.NewBuffer(content)
		remote, err := client.Create(path)
		if err != nil {
			return err
		}
		io.Copy(remote, local)
		remote.Close()
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

func (s *SSHJobExecutor) cleanupDir(dir string) error {
	cmd := fmt.Sprintf("rm -rf %s", dir)
	return executeCommand(s.client, cmd, s.workingDir, nil, io.Discard)
}

// Cleanup removes the working dir.
func (s *SSHJobExecutor) Cleanup(ctx context.Context) error {
	if s.workingDir == "" {
		return nil
	}
	return s.cleanupDir(s.workingDir)
}
