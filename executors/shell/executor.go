package shell

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/projecteru2/phistage/common"
	"github.com/projecteru2/phistage/helpers/command"
	"github.com/projecteru2/phistage/helpers/variable"
	"github.com/projecteru2/phistage/store"
)

type ShellJobExecutor struct {
	store  store.Store
	config *common.Config

	job      *common.Job
	phistage *common.Phistage

	output         io.Writer
	workingDir     string
	jobEnvironment map[string]string
}

// NewEruJobExecutor creates an ERU executor for this job.
// Since job needs to know its context, phistage is assigned too.
func NewShellJobExecutor(job *common.Job, phistage *common.Phistage, output io.Writer, store store.Store, config *common.Config) (*ShellJobExecutor, error) {
	return &ShellJobExecutor{
		store:          store,
		config:         config,
		job:            job,
		phistage:       phistage,
		output:         output,
		jobEnvironment: phistage.Environment,
	}, nil
}

// Prepare creates a temp working dir for this job.
func (ls *ShellJobExecutor) Prepare(ctx context.Context) error {
	var err error
	ls.workingDir, err = ioutil.TempDir("", "phistage-*")
	return err
}

// defaultEnvironmentVariables sets some useful information into environment variables.
// This will be set to the whole running context within the workload.
func (ls *ShellJobExecutor) defaultEnvironmentVariables() map[string]string {
	return map[string]string{
		"PHISTAGE_WORKING_DIR": ls.workingDir,
		"PHISTAGE_JOB_NAME":    ls.job.Name,
	}
}

// Execute will execute all steps within this job one by one
func (ls *ShellJobExecutor) Execute(ctx context.Context) error {
	for _, step := range ls.job.Steps {
		var err error
		switch step.Uses {
		case "":
			err = ls.executeStep(ctx, step)
		default:
			// step, err = e.replaceStepWithUses(ctx, step)
			// if err != nil {
			// 	return err
			// }
			err = ls.executeKhoriumStep(ctx, step)
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
func (ls *ShellJobExecutor) replaceStepWithUses(ctx context.Context, step *common.Step) (*common.Step, error) {
	if step.Uses == "" {
		return step, nil
	}

	uses, err := ls.store.GetRegisteredStep(ctx, step.Uses)
	if err != nil {
		return nil, err
	}
	s := &common.Step{
		Name:        step.Name,
		Run:         uses.Run,
		OnError:     uses.OnError,
		Environment: command.MergeVariables(uses.Environment, step.Environment),
		With:        command.MergeVariables(uses.With, step.With),
	}
	// in case name of this step is empty
	if s.Name == "" {
		s.Name = uses.Name
	}
	return s, nil
}

// executeStep executes a step.
// It first replace the step with uses if uses is given,
// then prepare the arguments and environments to the command.
// Then execute the command, retrieve the output, the execution will stop if any error occurs.
// It then retries to execute the OnError commands, also with the arguments and environments.
func (ls *ShellJobExecutor) executeStep(ctx context.Context, step *common.Step) error {
	var (
		err  error
		vars map[string]string
	)

	vars, err = ls.store.GetVariablesForPhistage(ctx, ls.phistage.Name)
	if err != nil {
		return err
	}

	environment := command.MergeVariables(ls.jobEnvironment, step.Environment)

	defer func() {
		if !errors.Is(err, common.ErrExecutionError) {
			return
		}
		if err := ls.executeCommands(ctx, step.OnError, step.With, environment, vars); err != nil {
			logrus.WithField("step", step.Name).WithError(err).Errorf("[EruJobExecutor] error when executing on_error")
		}
	}()

	err = ls.executeCommands(ctx, step.Run, step.With, environment, vars)
	return err
}

// executeKhoriumStep executes a KhoriumStep defined by step.Uses.
func (ls *ShellJobExecutor) executeKhoriumStep(ctx context.Context, step *common.Step) error {
	ks, err := ls.store.GetRegisteredKhoriumStep(ctx, step.Uses)
	if err != nil {
		return err
	}

	vars, err := ls.store.GetVariablesForPhistage(ctx, ls.phistage.Name)
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
	envs := command.MergeVariables(ls.defaultEnvironmentVariables(), step.Environment)
	envs = command.MergeVariables(envs, ksEnv)

	khoriumStepWorkingDir, err := ioutil.TempDir("", "phistage-khoriumstep-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(khoriumStepWorkingDir)

	files := map[string][]byte{}
	for name, content := range ks.Files {
		files[filepath.Join(khoriumStepWorkingDir, name)] = content
	}
	if err := ls.createNecessaryDirs(ctx, files); err != nil {
		return err
	}

	if err := ls.writeFiles(files); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", ks.Run.Main)
	cmd.Dir = khoriumStepWorkingDir
	cmd.Env = command.ToEnvironmentList(envs)
	cmd.Stdout = ls.output
	cmd.Stderr = ls.output
	if err := cmd.Run(); err != nil {
		return errors.WithMessagef(common.ErrExecutionError, "exec error: %v", err)
	}
	return nil
}

func (ls *ShellJobExecutor) writeFiles(files map[string][]byte) error {
	for path, content := range files {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		if _, err := f.Write(content); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}

// createNecessaryDirs creates essential dirs for files,
// otherwise error occurs when we open files for writting.
func (ls *ShellJobExecutor) createNecessaryDirs(ctx context.Context, files map[string][]byte) error {
	// golang is really, really stupid
	dirs := map[string]struct{}{}
	for path := range files {
		dirs[filepath.Dir(path)] = struct{}{}
	}

	dirnames := []string{"-p"}
	for path := range dirs {
		dirnames = append(dirnames, path)
	}

	return exec.CommandContext(ctx, "/bin/mkdir", dirnames...).Run()
}

// executeCommands executes cmd with given arguments, environments and variables.
// use args, envs, and reserved vars to build the cmd.
func (ls *ShellJobExecutor) executeCommands(ctx context.Context, cmds []string, args, env, vars map[string]string) error {
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

	for _, c := range commands {
		cmd := exec.CommandContext(ctx, "/bin/sh", "-c", c)
		cmd.Dir = ls.workingDir
		cmd.Env = command.ToEnvironmentList(command.MergeVariables(ls.defaultEnvironmentVariables(), env))
		cmd.Stdout = ls.output
		cmd.Stderr = ls.output
		if err := cmd.Run(); err != nil {
			return errors.WithMessagef(common.ErrExecutionError, "exec error: %v", err)
		}
	}
	return nil
}

// Cleanup removes the temp working dir.
func (ls *ShellJobExecutor) Cleanup(ctx context.Context) error {
	return os.RemoveAll(ls.workingDir)
}
